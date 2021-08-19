package operations

import (
	"errors"

	"github.com/beyondstorage/go-storage/v4/types"
)

func (so *SingleOperator) List(path string) chan *types.Object {
	it, err := so.store.List(path)
	if err != nil {
		so.errCh <- err
		return nil
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
				so.errCh <- err
				return
			}
			ch <- o
		}
	}()

	return ch
}
