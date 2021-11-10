package operations

import (
	"errors"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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

func hasMeta(path string) bool {
	magicChars := `*?[{`
	if runtime.GOOS != "windows" {
		magicChars = `*?[{\`
	}
	return strings.ContainsAny(path, magicChars)
}

func splitPattern(path string) (base, pattern string) {
	if !hasMeta(path) {
		return path, ""
	}

	base, pattern = doublestar.SplitPattern(path)
	if base == "." {
		base = ""
	} else {
		base += "/"
	}

	return
}

func (so *SingleOperator) ListWithGlob(path string) (ch chan *ObjectResult, err error) {
	ch = make(chan *ObjectResult, 16)

	go func() {
		defer close(ch)

		so.listWithGlob(ch, path, true)
	}()

	return ch, nil
}

func (so *SingleOperator) listWithGlob(
	ch chan *ObjectResult,
	path string,
	errFlag bool,
) {
	base, pattern := splitPattern(path)
	it, err := so.store.List(base, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		if errFlag {
			ch <- &ObjectResult{Error: err}
		}
		return
	}

	subPatterns := strings.SplitN(pattern, "/", 2)
	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			if errFlag {
				ch <- &ObjectResult{Error: err}
			}
			break
		}

		switch len(subPatterns) {
		case 1:
			if subPatterns[0] == "" {
				ch <- &ObjectResult{Object: o}
			} else if ok, _ := doublestar.Match(base+subPatterns[0], o.Path); ok {
				if !o.Mode.IsDir() {
					ch <- &ObjectResult{Object: o}
				} else {
					so.listWithGlob(ch, o.Path, false)
				}
			}
			break
		case 2:
			ok1, _ := doublestar.Match(base+subPatterns[0], o.Path)
			ok2, _ := doublestar.Match(base+subPatterns[0]+"/", o.Path)
			if ok1 || ok2 {
				if o.Mode.IsDir() {
					so.listWithGlob(ch, strings.TrimSuffix(o.Path, "/")+"/"+subPatterns[1], false)
				}
			}
			break
		}
	}
}

func (so *SingleOperator) ListRecursivelyWithGlob(path string) (ch chan *ObjectResult, err error) {
	ch = make(chan *ObjectResult, 16)

	go func() {
		defer close(ch)

		so.listRecursivelyWithGlob(ch, path, true)
	}()

	return ch, nil
}

func (so *SingleOperator) listRecursivelyWithGlob(
	ch chan *ObjectResult,
	path string,
	errFlag bool,
) {
	base, pattern := splitPattern(path)
	it, err := so.store.List(base, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		if errFlag {
			ch <- &ObjectResult{Error: err}
		}
		return
	}

	subPatterns := strings.SplitN(pattern, "/", 2)
	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			if errFlag {
				ch <- &ObjectResult{Error: err}
			}
			break
		}

		switch len(subPatterns) {
		case 1:
			if subPatterns[0] == "" {
				if o.Mode.IsDir() {
					so.listRecursively(ch, o.Path)
				} else {
					ch <- &ObjectResult{Object: o}
				}
			} else if ok, _ := doublestar.Match(base+subPatterns[0], o.Path); ok {
				if !o.Mode.IsDir() {
					ch <- &ObjectResult{Object: o}
				} else {
					so.listRecursively(ch, o.Path)
				}
			}
			break
		case 2:
			ok1, _ := doublestar.Match(base+subPatterns[0], o.Path)
			ok2, _ := doublestar.Match(base+subPatterns[0]+"/", o.Path)
			if ok1 || ok2 {
				if o.Mode.IsDir() {
					so.listRecursivelyWithGlob(ch, strings.TrimSuffix(o.Path, "/")+"/"+subPatterns[1], false)
				}
			}
			break
		}
	}
}
