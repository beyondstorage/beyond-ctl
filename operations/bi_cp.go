package operations

import (
	"io"
)

func (bo *BiOperator) Copy(src, dst string) {
	obj, err := bo.src.Stat(src)
	if err != nil {
		bo.errCh <- err
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
				bo.errCh <- err
			}
		}()
		_, err := bo.src.Read(src, w)
		if err != nil {
			bo.errCh <- err
			return
		}
	}()

	_, err = bo.dst.Write(dst, r, size)
	if err != nil {
		bo.errCh <- err
		return
	}
}
