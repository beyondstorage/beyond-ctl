package operations

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"sync"

	"go.uber.org/zap"

	"go.beyondstorage.io/v5/types"
)

// TeeRun will write a file fetched from standard input through Multipart related operations.
//
// We will:
// - Create a multipart object.
// - Write into this multipart object via split source file into parts (read by offset)
// - Complete the multipart object.
//
// We have two channels have:
// - errch is returned to cmd and used as an error channel.
// - partch is used internally to control the part write multipart logic.
func (so *SingleOperator) TeeRun(path string, expectedSize int64, r io.Reader) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	partch := make(chan *PartResult, 4)

	multiparter, ok := so.store.(types.Multiparter)
	if !ok {
		return nil, fmt.Errorf("multiparter")
	}

	mo, err := multiparter.CreateMultipart(path)
	if err != nil {
		return nil, err
	}

	partSize, err := calculatePartSize(so.store, expectedSize)
	if err != nil {
		return nil, err
	}

	go func() {
		// Close partch to inform that all parts have been done.
		defer close(partch)

		wg := &sync.WaitGroup{}
		var index int

		b := make([]byte, partSize)
		// When this flag is true, it means that the data has been read and it is time to exit the read loop.
		flag := false

		for {
			if flag {
				break
			}

			wg.Add(1)

			taskIndex := index

			n, err := io.ReadFull(r, b)
			if err == io.EOF {
				wg.Done()
				break
			}
			if err != nil && n != 0 {
				flag = true
				err = nil
			}
			if err != nil {
				partch <- &PartResult{Error: err}
				return
			}

			rd := bytes.NewReader(b[:n])

			err = so.pool.Submit(func() {
				defer wg.Done()

				_, part, err := multiparter.WriteMultipart(mo, rd, rd.Size(), taskIndex)
				if err != nil {
					partch <- &PartResult{Error: err}
					return
				}
				partch <- &PartResult{Part: part}
			})
			if err != nil {
				so.logger.Error("submit task", zap.Error(err))
				return
			}

			index++
		}

		wg.Wait()
	}()

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

	err = multiparter.CompleteMultipart(mo, parts)
	if err != nil {
		return nil, err
	}

	return
}
