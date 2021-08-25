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
		defer close(ch)

		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				break
			}
			if err != nil {
				ch <- &ListResult{Error: err}
				break
			}
			ch <- &ListResult{Object: o}
		}
	}()

	return ch, nil
}
