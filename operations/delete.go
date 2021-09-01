package operations

import (
	"sync"
)

func (so *SingleOperator) Delete(path string) (err error) {
	err = so.store.Delete(path)
	if err != nil {
		return err
	}

	return nil
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
