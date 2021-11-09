package operations

// MoveFileViaWrite will move a file via Write operation.
func (do *DualOperator) MoveFileViaWrite(src, dst string, size int64) (err error) {
	cch, err := do.CopyFileViaWrite(src, dst, size)
	if err != nil {
		return err
	}

	for v := range cch {
		if v.Error != nil {
			return v.Error
		}
	}

	so := NewSingleOperator(do.src)
	err = so.Delete(src)
	if err != nil {
		return err
	}

	return nil
}

// MoveFileViaMultipart will move a file via Multipart related operation.
func (do *DualOperator) MoveFileViaMultipart(src, dst string, totalSize int64) (err error) {
	cch, err := do.CopyFileViaMultipart(src, dst, totalSize)
	if err != nil {
		return err
	}

	for v := range cch {
		if v.Error != nil {
			return v.Error
		}
	}

	so := NewSingleOperator(do.src)
	err = so.Delete(src)
	if err != nil {
		return err
	}

	return nil
}

// MoveRecursively will move directories recursively.
func (do *DualOperator) MoveRecursively(src, dst string, multipartThreshold int64) (err error) {
	cch, err := do.CopyRecursively(src, dst, multipartThreshold)
	if err != nil {
		return err
	}

	for v := range cch {
		if v.Error != nil {
			return v.Error
		}
	}

	so := NewSingleOperator(do.src)
	dch, err := so.DeleteRecursively(src)
	if err != nil {
		return err
	}

	for v := range dch {
		if v.Error != nil {
			return v.Error
		}
	}

	return nil
}
