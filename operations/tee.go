package operations

import (
	"bufio"
	"bytes"
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

func (so *SingleOperator) TeeRunViaPipe(key string, multipartThreshold int64) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	defer close(errch)

	var inputs []byte

	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		b := s.Bytes()
		b = append(b, '\n')
		inputs = append(inputs, b...)
	}

	if int64(len(inputs)) < multipartThreshold {
		r := bytes.NewReader(inputs)
		_, err = so.store.Write(key, r, r.Size())
		if err != nil {
			return nil, err
		}
	} else {
		ch, err := so.teeViaMultipart(key, inputs)
		if err != nil {
			return nil, err
		}

		for v := range ch {
			if v.Error != nil {
				errch <- &EmptyResult{Error: v.Error}
			}
		}
	}
	fmt.Printf("Stdin is saved to <%s>", key)
	fmt.Print("\n")
	os.Exit(0)

	return
}

func (so *SingleOperator) teeViaMultipart(path string, inputs []byte) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	partch := make(chan *PartResult, 4)

	size := int64(len(inputs))

	multiparter, ok := so.store.(types.Multiparter)
	if !ok {
		return nil, fmt.Errorf("multiparter")
	}

	mo, err := multiparter.CreateMultipart(path)
	if err != nil {
		return nil, err
	}

	partSize, err := calculatePartSize(so.store, size)
	if err != nil {
		return nil, err
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

			err = so.pool.Submit(func() {
				defer wg.Done()

				input := inputs[taskOffset : taskSize+taskOffset-1]
				r := bytes.NewReader(input)
				_, part, err := multiparter.WriteMultipart(mo, r, r.Size(), taskIndex)
				if err != nil {
					partch <- &PartResult{Error: err}
					return
				}
				partch <- &PartResult{Part: part}
			})
			if err != nil {
				so.logger.Error("submit task", zap.Error(err))
				break
			}

			index++
			offset += partSize
			// Offset >= totalSize means we have read all content
			if offset >= size {
				break
			}
			// Handle the last part
			if offset+partSize > size {
				partSize = size - offset
			}
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

func (so *SingleOperator) TeeRun(key string) (errch chan *EmptyResult, err error) {
	errch = make(chan *EmptyResult, 4)
	ch := make(chan os.Signal, 1)

	var inputs []byte

	go func() {
		defer close(errch)

		wg := &sync.WaitGroup{}
		r := bufio.NewReader(os.Stdin)

		for {
			flag := false
			wg.Add(1)

			err = so.pool.Submit(func() {
				defer wg.Done()

				line, err := r.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						flag = true
						return
					} else {
						errch <- &EmptyResult{Error: err}
						return
					}
				}

				input := string(line)
				fmt.Print(input)

				inputs = append(inputs, line...)
			})
			if err != nil {
				so.logger.Error("submit task", zap.Error(err))
				break
			}
			if flag {
				break
			}
		}

		wg.Wait()
	}()

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		r := bytes.NewReader(inputs)
		_, err = so.store.Write(key, r, r.Size())
		if err != nil {
			return nil, err
		}
		fmt.Print("\n")
		fmt.Printf("Stdin is saved to <%s>", key)
		fmt.Print("\n")
		os.Exit(0)
	}

	return
}
