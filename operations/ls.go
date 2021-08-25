package operations

import (
	"errors"

	"github.com/beyondstorage/go-storage/v4/types"
)

func (oo *SingleOperator) List(path string) (ch chan *ObjectResult, err error) {
	it, err := oo.store.List(path)
	if err != nil {
		return nil, err
	}

	ch = make(chan *ObjectResult, 16)
	go func() {
		defer close(ch)

		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				break
			}
			if err != nil {
				ch <- &ObjectResult{Error: err}
				break
			}
			ch <- &ObjectResult{Object: o}
		}
	}()

	return ch, nil
}
