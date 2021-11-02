package operations

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"go.beyondstorage.io/v5/types"
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

	fmt.Printf("Stdin is saved to <%s>\n", path)
	os.Exit(0)

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

func (so *SingleOperator) TeeRun(path string) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	ch := make(chan os.Signal, 1)

	var inputs []byte
	flag := false

	go func() {
		defer close(errch)
		r := bufio.NewReader(os.Stdin)

		for {
			line, err := r.ReadBytes('\n')
			if err != nil && errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				errch <- &EmptyResult{Error: err}
				flag = true
				return
			}

			output := string(line)
			fmt.Print(output)

			inputs = append(inputs, line...)
		}
	}()

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		if !flag {
			r := bytes.NewReader(inputs)
			_, err = so.store.Write(path, r, r.Size())
			if err != nil {
				return nil, err
			}
			fmt.Printf("\nStdin is saved to <%s>\n", path)
			os.Exit(0)
		} else {
			return errch, err
		}
	}

	return
}
