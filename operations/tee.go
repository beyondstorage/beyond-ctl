package operations

import (
	"bufio"
	"bytes"
	"fmt"
	"go.beyondstorage.io/v5/types"
	"go.uber.org/zap"
	"io"
	"os"
	"sort"
	"sync"
)

func (so *SingleOperator) TeeRunViaPipe(path string, expectSize int64) (err error) {
	ch, err := so.teeViaMultipart(path, expectSize)
	if err != nil {
		return err
	}
	for v := range ch {
		if v.Error != nil {
			return v.Error
		}
	}

	return err
}

func (so *SingleOperator) teeViaMultipart(path string, expectSize int64) (errch chan *EmptyResult, err error) {
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

	partSize, err := calculatePartSize(so.store, expectSize)
	if err != nil {
		return nil, err
	}

	go func() {
		// Close partch to inform that all parts have been done.
		defer close(partch)

		wg := &sync.WaitGroup{}
		var index int

		r := bufio.NewReader(os.Stdin)
		b := make([]byte, partSize)

		for {
			wg.Add(1)

			taskIndex := index

			n, err := io.ReadFull(r, b)
			if err == io.EOF {
				wg.Done()
				break
			}
			if err != nil && n != 0 {
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

	go func() {
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
			errch <- &EmptyResult{Error: err}
			return
		}
	}()

	return errch, nil
}

func (so *SingleOperator) TeeRun(path string) (err error) {
	var inputs []byte

	r := bufio.NewScanner(os.Stdin)

	for r.Scan() {
		input := r.Bytes()
		input = append(input, '\n')

		output := string(input)
		fmt.Print(output)

		inputs = append(inputs, input...)
	}

	rd := bytes.NewReader(inputs)
	_, err = so.store.Write(path, rd, rd.Size())
	if err != nil {
		return err
	}

	return
}
