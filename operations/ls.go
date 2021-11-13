package operations

import (
	"errors"

	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) List(path string) (ch chan *ObjectResult, err error) {
	it, err := so.store.List(path, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		return nil, err
	}

	ch = make(chan *ObjectResult, 16)
	go func() {
		defer close(ch)

		for {
			o, err := it.Next()
			if err != nil {
				ch <- &ObjectResult{Error: err}
				break
			}

			ch <- &ObjectResult{Object: o}
		}
	}()

	return ch, nil
}
func (so *SingleOperator) ListRecursively(path string) (ch chan *ObjectResult, err error) {
	ch = make(chan *ObjectResult, 16)

	go func() {
		defer close(ch)

		so.listRecursively(ch, path)
	}()

	return ch, nil
}

func (so *SingleOperator) listRecursively(
	ch chan *ObjectResult,
	path string,
) {
	it, err := so.store.List(path, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		ch <- &ObjectResult{Error: err}
		return
	}

	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			ch <- &ObjectResult{Error: err}
			break
		}

		if o.Mode.IsDir() {
			so.listRecursively(ch, o.Path)
		}
		ch <- &ObjectResult{Object: o}
	}
}
