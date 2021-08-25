package operations

import (
	"errors"

	"github.com/beyondstorage/go-storage/v4/types"
)

// ObjectResult is the result for Object.
// Only one of Object or Error will be valid.
// We need to check Error before use Object.
type ObjectResult struct {
	Object *types.Object
	Error  error
}

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
