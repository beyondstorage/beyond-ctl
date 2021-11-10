package operations

import (
	"errors"
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

func splitPattern(path string) (base, pattern string) {
	hasMeta := false

	splitIdx := -1
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '\\' {
			i++
		} else if c == '/' {
			splitIdx = i
		} else if c == '*' || c == '?' || c == '[' || c == '{' {
			hasMeta = true
			break
		}
	}

	if hasMeta {
		if splitIdx >= 0 {
			return path[:splitIdx+1], path[splitIdx+1:]
		}

		return "", path
	}

	return path, ""
}

func (so *SingleOperator) ListWithGlob(path string) (ch chan *ObjectResult, err error) {
	ch = make(chan *ObjectResult, 16)

	go func() {
		defer close(ch)

		so.listWithGlob(ch, path)
	}()

	return ch, nil
}

func (so *SingleOperator) listWithGlob(
	ch chan *ObjectResult,
	path string,
) {
	base, pattern := splitPattern(path)
	it, err := so.store.List(base, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		ch <- &ObjectResult{Error: err}
		return
	}

	subPatterns := strings.SplitN(pattern, "/", 2)
	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			ch <- &ObjectResult{Error: err}
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
					so.listWithGlob(ch, o.Path)
				}
			}
			break
		case 2:
			ok1, _ := doublestar.Match(base+subPatterns[0], o.Path)
			ok2, _ := doublestar.Match(base+subPatterns[0]+"/", o.Path)
			if ok1 || ok2 {
				if o.Mode.IsDir() {
					so.listWithGlob(ch, strings.TrimSuffix(o.Path, "/")+"/"+subPatterns[1])
				}
			}
			break
		}
	}
}
