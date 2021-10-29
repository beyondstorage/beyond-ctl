package operations

import (
	"errors"
	"sync"

	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) Delete(path string, pairs ...types.Pair) (err error) {
	err = so.store.Delete(path, pairs...)
	if err != nil {
		return err
	}

	return nil
}

func (so *SingleOperator) DeleteMultipart(path string) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	it, err := so.store.List(path, pairs.WithListMode(types.ListModePart))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)

		wg := sync.WaitGroup{}

		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				break
			}
			if err != nil {
				ch <- &EmptyResult{Error: err}
				break
			}

			wg.Add(1)

			err = so.pool.Submit(func() {
				defer wg.Done()

				if o.Path == path {
					err = so.Delete(path, pairs.WithMultipartID(o.MustGetMultipartID()))
					if err != nil {
						ch <- &EmptyResult{Error: err}
						return
					}
				}
			})
			if err != nil {
				ch <- &EmptyResult{Error: err}
				break
			}
		}

		wg.Wait()
	}()

	return
}

func (so *SingleOperator) DeleteMultipartViaRecursively(path string) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	it, err := so.store.List(path, pairs.WithListMode(types.ListModePart))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)

		wg := sync.WaitGroup{}

		for {
			o, err := it.Next()
			if err != nil && errors.Is(err, types.IterateDone) {
				break
			}
			if err != nil {
				ch <- &EmptyResult{Error: err}
				break
			}

			wg.Add(1)

			err = so.pool.Submit(func() {
				defer wg.Done()

				err = so.Delete(o.Path, pairs.WithMultipartID(o.MustGetMultipartID()))
				if err != nil {
					if err != nil {
						ch <- &EmptyResult{Error: err}
						return
					}
				}
			})
			if err != nil {
				ch <- &EmptyResult{Error: err}
				break
			}
		}

		wg.Wait()
	}()

	return
}

func (so *SingleOperator) DeleteRecursively(path string) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	och, err := so.ListRecursively(path)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)

		wg := &sync.WaitGroup{}

		for or := range och {
			if or.Error != nil {
				ch <- &EmptyResult{Error: or.Error}
				break
			}
			object := or.Object

			wg.Add(1)
			err = so.pool.Submit(func() {
				defer wg.Done()

				err = so.Delete(object.Path)
				if err != nil {
					ch <- &EmptyResult{Error: err}
				}
			})
			if err != nil {
				ch <- &EmptyResult{Error: err}
				break
			}
		}

		wg.Wait()
	}()

	return ch, nil
}
