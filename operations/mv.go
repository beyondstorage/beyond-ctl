package operations

// MoveFileViaWrite will move a file via Write operation.
func (do *DualOperator) MoveFileViaWrite(src, dst string, size int64) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	cch, err := do.CopyFileViaWrite(src, dst, size)
	if err != nil {
		ch <- &EmptyResult{Error: err}
		return nil, err
	}

	go func() {
		defer close(ch)

		for v := range cch {
			if v.Error != nil {
				ch <- &EmptyResult{Error: err}
				return
			}
		}

		so := NewSingleOperator(do.src)
		err = so.Delete(src)
		if err != nil {
			ch <- &EmptyResult{Error: err}
			return
		}
	}()

	return ch, nil
}

// MoveFileViaMultipart will move a file via Multipart related operation.
func (do *DualOperator) MoveFileViaMultipart(src, dst string, totalSize int64) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	cch, err := do.CopyFileViaMultipart(src, dst, totalSize)
	if err != nil {
		ch <- &EmptyResult{Error: err}
		return nil, err
	}

	go func() {
		defer close(ch)

		for v := range cch {
			if v.Error != nil {
				ch <- &EmptyResult{Error: err}
				return
			}
		}

		so := NewSingleOperator(do.src)
		err = so.Delete(src)
		if err != nil {
			ch <- &EmptyResult{Error: err}
			return
		}
	}()

	return ch, nil
}

// MoveRecursively will move directories recursively.
func (do *DualOperator) MoveRecursively(src, dst string, multipartThreshold int64) (ch chan *EmptyResult, err error) {
	ch = make(chan *EmptyResult, 4)

	cch, err := do.CopyRecursively(src, dst, multipartThreshold)
	if err != nil {
		ch <- &EmptyResult{Error: err}
		return nil, err
	}

	go func() {
		defer close(ch)

		for v := range cch {
			if v.Error != nil {
				ch <- &EmptyResult{Error: err}
				return
			}
		}

		so := NewSingleOperator(do.src)
		dch, err := so.DeleteRecursively(src)
		if err != nil {
			ch <- &EmptyResult{Error: err}
			return
		}

		for v := range dch {
			if v.Error != nil {
				ch <- &EmptyResult{Error: err}
				return
			}
		}
	}()

	return ch, nil
}
