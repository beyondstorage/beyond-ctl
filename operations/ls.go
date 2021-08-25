package operations

import (
	"errors"

	"github.com/beyondstorage/go-storage/v4/types"
)

type ListResult struct {
	Object *types.Object
	Error  error
}

func (oo *SingleOperator) List(path string) (ch chan *ListResult, err error) {
	it, err := oo.store.List(path)
	if err != nil {
		return nil, err
	}

	ch = make(chan *ListResult, 16)
	go func() {
		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				close(ch)
				break
			}
			if err != nil {
				ch <- &ListResult{Error: err}
				close(ch)
				return
			}
			ch <- &ListResult{Object: o}
		}
	}()

	return ch, nil
}
