package operations

import (
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/types"
	"go.uber.org/zap"
)

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

		_, err := do.src.Read(src, w)
		if err != nil {
			do.logger.Error("pipe read", zap.String("path", src), zap.Error(err))
			ch <- &EmptyResult{Error: err}
		}
	}()

	go func() {
		defer close(ch)

		_, err := do.dst.Write(dst, r, size)
		if err != nil {
			do.logger.Error("pipe write", zap.String("path", dst), zap.Error(err))
			ch <- &EmptyResult{Error: err}
		}
	}()

	return ch, nil
}

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
		defer close(partch)

		wg := &sync.WaitGroup{}
		var offset int64
		var index int

		for {
			// handle size for the last part
			if offset+partSize > totalSize {
				partSize = totalSize - offset
			}

			wg.Add(1)
			err = do.pool.Submit(func() {
				do.copyMultipart(partch, wg, src, dstObj, partSize, offset, index)
			})
			if err != nil {
				do.logger.Error("submit task", zap.Error(err))
				errch <- &EmptyResult{Error: err}
				break
			}

			offset += partSize
			if offset >= totalSize {
				break
			}
			index++
		}

		wg.Wait()
	}()

	go func() {
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

		_, err := do.src.Read(src, w, pairs.WithSize(size), pairs.WithOffset(offset))
		if err != nil {
			do.logger.Error("pipe read", zap.String("path", src), zap.Error(err))
			ch <- &PartResult{Error: err}
		}
	}()

	multiparter := do.dst.(types.Multiparter)

	_, p, err := multiparter.WriteMultipart(dstObj, r, size, index)
	if err != nil {
		do.logger.Error("pipe write", zap.String("path", dstObj.Path), zap.Error(err))
		ch <- &PartResult{Error: err}
	}
	ch <- &PartResult{Part: p}
}
