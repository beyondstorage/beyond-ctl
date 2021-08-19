package operations

import (
	"io"
)

func (do *DualOperator) Copy(src, dst string) {
	obj, err := do.src.Stat(src)
	if err != nil {
		do.errCh <- err
		return
	}
	size := obj.MustGetContentLength()

	// assign dst by src if blank
	if dst == "" {
		dst = src
	}

	r, w := io.Pipe()

	go func() {
		defer func() {
			err := w.Close()
			if err != nil {
				do.errCh <- err
			}
		}()
		_, err := do.src.Read(src, w)
		if err != nil {
			do.errCh <- err
			return
		}
	}()

	_, err = do.dst.Write(dst, r, size)
	if err != nil {
		do.errCh <- err
		return
	}
}
