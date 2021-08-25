package operations

import (
	"io"
	"sync"

	"github.com/beyondstorage/go-storage/v4/pairs"
	"github.com/beyondstorage/go-storage/v4/types"
	"go.uber.org/zap"
)

func (do *DualOperator) Copy(src, dst string) (ch chan *EmptyResult, err error) {
	// assign dst by src if blank
	if dst == "" {
		dst = src
	}

	ch = make(chan *EmptyResult, 4)
	go func() {
		defer close(ch)

		err = do.CopyFile(src, dst)
		if err != nil {
			ch <- &EmptyResult{Error: err}
		}
	}()

	return ch, nil
}

func (do *DualOperator) CopyFile(src, dst string) error {
	obj, err := do.src.Stat(src)
	if err != nil {
		return err
	}

	size := obj.MustGetContentLength()
	if _, ok := do.dst.(types.Multiparter); size > defaultMultipartThreshold && ok {
		return do.CopyLargeFile(src, dst, size)
	}

	return do.CopySmallFile(src, dst, size)
}

func (do *DualOperator) CopyLargeFile(src, dst string, totalSize int64) error {
	logger, _ := zap.NewDevelopment()

	multiparter := do.dst.(types.Multiparter)

	obj, err := multiparter.CreateMultipart(dst)
	if err != nil {
		return err
	}

	partSize, err := calculatePartSize(do.src, totalSize)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var offset int64
	var index uint32
	parts := make([]*types.Part, 0)

	multipartID := obj.MustGetMultipartID()
	for {
		wg.Add(1)
		// handle size for the last part
		if offset+partSize > totalSize {
			partSize = totalSize - offset
		}

		parts = append(parts, &types.Part{
			Index: int(index),
			Size:  partSize,
		})

		go func(size int64, offset int64, index uint32) {
			defer wg.Done()

			r, w := io.Pipe()

			go func() {
				defer func() {
					cErr := w.Close()
					if cErr != nil {
						logger.Error("close writer", zap.Error(err))
					}
				}()
				_, err = do.src.Read(src, w, pairs.WithSize(size), pairs.WithOffset(offset))
				if err != nil {
					logger.Error("read from", zap.String("path", src),
						zap.Int64("size", size), zap.Int64("offset", offset), zap.Error(err))
				}
			}()

			o := do.dst.Create(dst, pairs.WithMultipartID(multipartID))
			_, part, err := multiparter.WriteMultipart(o, r, size, int(index))
			if err != nil {
				logger.Error("write multipart", zap.String("path", dst), zap.Int64("size", size),
					zap.Uint32("index", index), zap.String("id", multipartID), zap.Error(err))
				return
			}
			parts[int(index)].ETag = part.ETag
		}(partSize, offset, index)

		offset += partSize
		if offset >= totalSize {
			break
		}
		index++
	}

	wg.Wait()

	err = multiparter.CompleteMultipart(obj, parts)
	if err != nil {
		return err
	}
	return nil
}

func (do *DualOperator) CopySmallFile(src, dst string, size int64) error {
	logger, _ := zap.NewDevelopment()
	r, w := io.Pipe()

	go func() {
		defer func() {
			cErr := w.Close()
			if cErr != nil {
				logger.Error("close writer", zap.Error(cErr))
			}
		}()
		_, err := do.src.Read(src, w)
		if err != nil {
			logger.Error("read from source", zap.String("path", src), zap.Int64("size", size))
		}
	}()

	_, err := do.dst.Write(dst, r, size)
	if err != nil {
		return err
	}
	return nil
}
