package operations

import (
	"io"
)

func (do *DualOperator) Copy(src, dst string) (ch chan *ObjectResult, err error) {
	obj, err := do.src.Stat(src)
	if err != nil {
		return nil, err
	}

	size := obj.MustGetContentLength()

	// assign dst by src if blank
	if dst == "" {
		dst = src
	}

	ch = make(chan *ObjectResult, 16)
	go func() {
		defer close(ch)

		r, w := io.Pipe()

		go func() {
			defer func() {
				cErr := w.Close()
				if cErr != nil {
					ch <- &ObjectResult{Error: cErr}
				}
			}()
			_, err = do.src.Read(src, w)
			if err != nil {
				ch <- &ObjectResult{Error: err}
				return
			}
		}()

		_, err = do.dst.Write(dst, r, size)
		if err != nil {
			ch <- &ObjectResult{Error: err}
			return
		}
	}()

	return ch, nil
}
