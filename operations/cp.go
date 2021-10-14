package operations

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/types"
)

// CopyFileViaWrite will copy a file via Write operation.
func (do *DualOperator) CopyFileViaWrite(src, dst string, size int64) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	r, w := io.Pipe()

	go func() {
		defer func() {
			err := w.Close()
			if err != nil {
				do.logger.Error("close pipe writer", zap.Error(err))
				ch <- &EmptyResult{Error: err}
			}
		}()

		_, err := do.src.Read(src, w, do.readPairs...)
		if err != nil {
			do.logger.Error("pipe read", zap.String("path", src), zap.Error(err))
			ch <- &EmptyResult{Error: err}
		}
	}()

	go func() {
		defer close(ch)

		_, err := do.dst.Write(dst, r, size, do.writePairs...)
		if err != nil {
			do.logger.Error("pipe write", zap.String("path", dst), zap.Error(err))
			ch <- &EmptyResult{Error: err}
		}
	}()

	return ch, nil
}

// CopyFileViaMultipart will copy a file via Multipart related operation.
//
// We will:
// - Create a multipart object.
// - Write into this multipart object via split source file into parts (read by offset)
// - Complete the multipart object.
//
// We have two channels have:
// - errch is returned to cmd and used as an error channel.
// - partch is used internally to control the part copy multipart logic.
func (do *DualOperator) CopyFileViaMultipart(src, dst string, totalSize int64) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	partch := make(chan *PartResult, 4)

	dstMultiparter, ok := do.dst.(types.Multiparter)
	if !ok {
		return nil, fmt.Errorf("dst is not a dstMultiparter")
	}

	dstObj, err := dstMultiparter.CreateMultipart(dst)
	if err != nil {
		return nil, fmt.Errorf("create multipart: %w", err)
	}

	partSize, err := calculatePartSize(do.dst, totalSize)
	if err != nil {
		return nil, fmt.Errorf("calculate part size: %w", err)
	}

	go func() {
		// Close partch to inform that all parts have been done.
		defer close(partch)

		wg := &sync.WaitGroup{}
		var offset int64
		var index int

		for {
			wg.Add(1)

			// Reallocate var here to prevent closure catch.
			taskSize := partSize
			taskIndex := index
			taskOffset := offset

			err = do.pool.Submit(func() {
				do.copyMultipart(partch, wg, src, dstObj, taskSize, taskOffset, taskIndex)
			})
			if err != nil {
				do.logger.Error("submit task", zap.Error(err))
				errch <- &EmptyResult{Error: err}
				break
			}

			index++
			offset += partSize
			// Offset >= totalSize means we have read all content
			if offset >= totalSize {
				break
			}
			// Handle the last part
			if offset+partSize > totalSize {
				partSize = totalSize - offset
			}
		}

		wg.Wait()
	}()

	go func() {
		// Close errch to inform that this copy operation has been done.
		defer close(errch)

		parts := make([]*types.Part, 0)
		for v := range partch {
			if v.Error != nil {
				errch <- &EmptyResult{Error: v.Error}
				continue
			}
			parts = append(parts, v.Part)
		}

		sort.SliceStable(parts, func(i, j int) bool {
			return parts[i].Index < parts[j].Index
		})

		err = dstMultiparter.CompleteMultipart(dstObj, parts)
		if err != nil {
			errch <- &EmptyResult{Error: err}
			return
		}
	}()

	return errch, nil
}

func (do *DualOperator) copyMultipart(
	ch chan *PartResult, wg *sync.WaitGroup,
	src string, dstObj *types.Object,
	size, offset int64, index int,
) {
	defer wg.Done()

	r, w := io.Pipe()

	go func() {
		defer func() {
			err := w.Close()
			if err != nil {
				do.logger.Error("close pipe writer", zap.Error(err))
				ch <- &PartResult{Error: err}
			}
		}()

		ps := make([]types.Pair, 0, len(do.readPairs)+2)
		ps = append(ps, pairs.WithSize(size), pairs.WithOffset(offset))
		ps = append(ps, do.readPairs...)

		_, err := do.src.Read(src, w, ps...)
		if err != nil {
			do.logger.Error("pipe read", zap.String("path", src), zap.Error(err))
			ch <- &PartResult{Error: err}
		}
	}()

	defer func() {
		err := r.Close()
		if err != nil {
			do.logger.Error("close pipe reader", zap.Error(err))
			ch <- &PartResult{Error: err}
		}
	}()

	multiparter := do.dst.(types.Multiparter)

	_, p, err := multiparter.WriteMultipart(dstObj, r, size, index, do.writePairs...)
	if err != nil {
		do.logger.Error("pipe write", zap.String("path", dstObj.Path), zap.Error(err))
		ch <- &PartResult{Error: err}
		return
	}
	ch <- &PartResult{Part: p}
}

// CopyRecursively will copy directories recursively.
func (do *DualOperator) CopyRecursively(src, dst string, multipartThreshold int64) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)

	so := NewSingleOperator(do.src)
	och, err := so.ListRecursively(src)
	if err != nil {
		errch <- &EmptyResult{Error: err}
	}

	if !strings.HasSuffix(dst, "/") {
		dst += "/"
	}

	go func() {
		defer close(errch)

		wg := &sync.WaitGroup{}

		for or := range och {
			if or.Error != nil {
				errch <- &EmptyResult{Error: err}
				break
			}
			object := or.Object

			if object.Mode.IsDir() {
				continue
			}

			path := dst + strings.TrimPrefix(object.Path, src)
			size := object.MustGetContentLength()

			wg.Add(1)
			err = so.pool.Submit(func() {
				defer wg.Done()

				if size < multipartThreshold {
					ch, err := do.CopyFileViaWrite(object.Path, path, size)
					if err != nil {
						errch <- &EmptyResult{Error: err}
					}
					for er := range ch {
						if er.Error != nil {
							errch <- &EmptyResult{Error: err}
						}
					}
				} else {
					ch, err := do.CopyFileViaMultipart(object.Path, path, size)
					if err != nil {
						errch <- &EmptyResult{Error: err}
					}
					for er := range ch {
						if er.Error != nil {
							errch <- &EmptyResult{Error: err}
						}
					}
				}
			})
			if err != nil {
				errch <- &EmptyResult{Error: err}
				break
			}
		}

		wg.Wait()
	}()

	return errch, nil
}
