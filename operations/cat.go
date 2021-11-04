package operations

import (
	"io"
	"os"
)

func (so *SingleOperator) CatFile(path string) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	r, w := io.Pipe()

	go func() {
		defer func() {
			err = w.Close()
			if err != nil {
				ch <- &EmptyResult{Error: err}
				return
			}
			close(ch)
		}()

		_, err = so.store.Read(path, w)
		if err != nil {
			ch <- &EmptyResult{Error: err}
			return
		}
	}()

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		return nil, err
	}

	return
}
