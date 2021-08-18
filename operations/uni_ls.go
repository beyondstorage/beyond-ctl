package operations

import (
	"errors"

	"github.com/beyondstorage/go-storage/v4/types"
)

func (uo *UniOperator) List(path string) chan *types.Object {
	it, err := uo.store.List(path)
	if err != nil {
		uo.errCh <- err
	}

	ch := make(chan *types.Object, 16)
	go func() {
		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				close(ch)
				break
			}
			if err != nil {
				uo.errCh <- err
				return
			}
			ch <- o
		}
	}()

	return ch
}
