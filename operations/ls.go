package operations

import (
	"errors"
	"runtime"
	"strings"

	"github.com/gobwas/glob"

	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) List(path string) (ch chan *ObjectResult, err error) {
	ch = make(chan *ObjectResult, 16)
	go func() {
		defer close(ch)

		objects, err := so.GlobList(path)
		if err != nil {
			ch <- &ObjectResult{Error: err}
			return
		}

		for _, object := range objects {
			if object.Mode.IsDir() {
				it, err := so.store.List(object.Path, pairs.WithListMode(types.ListModeDir))
				if err != nil {
					ch <- &ObjectResult{Error: err}
					break
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

					ch <- &ObjectResult{Object: o}
				}
			} else {
				ch <- &ObjectResult{Object: object}
			}
		}
	}()

	return ch, nil
}

func hasMeta(path string) bool {
	magicChars := `*?[{`
	if runtime.GOOS != "windows" {
		magicChars = `*?[{\`
	}
	return strings.ContainsAny(path, magicChars)
}

func splitPath(path string) (prefixPath, pattern string) {
	if path == "" {
		return
	}

	subpaths := strings.Split(path, "/")
	for _, subpath := range subpaths {
		if hasMeta(subpath) {
			pattern = path[len(prefixPath):]
			break
		} else {
			prefixPath += subpath + "/"
		}
	}

	prefixPath = strings.TrimSuffix(prefixPath, "/")

	return
}

func (so *SingleOperator) globList(path, pattern string) (objects []*types.Object, err error) {
	var g glob.Glob
	if pattern != "" {
		g = glob.MustCompile(pattern)
	}

	it, err := so.store.List(path, pairs.WithListMode(types.ListModeDir))
	if err != nil {
		return nil, err
	}

	for {
		o, err := it.Next()
		if err != nil && errors.Is(err, types.IterateDone) {
			break
		}
		if err != nil {
			return nil, err
		}

		if (g != nil && g.Match(o.Path)) || g == nil {
			objects = append(objects, o)
		}
	}

	return
}

func (so *SingleOperator) DirWalk(workPath string, pattern string) (objects []*types.Object, err error) {
	if pattern == "" {
		objs, err := so.globList(workPath, "")
		if err != nil {
			return nil, err
		}
		if len(objs) == 1 && !objs[0].Mode.IsDir() {
			objects = append(objects, objs...)
		} else {
			objects = append(objects, &types.Object{
				Mode: types.ModeDir,
				Path: workPath,
				ID:   workPath,
			})
		}

		return objects, nil
	}

	subpaths := strings.Split(pattern, "/")
	currentPattern := subpaths[0]
	objs, err := so.globList(workPath, workPath+currentPattern)
	if err != nil {
		return nil, err
	}

	if len(subpaths) == 1 {
		objects = append(objects, objs...)
		return objects, nil
	}

	postfixPath := pattern[len(currentPattern)+1:]
	for _, obj := range objs {
		if !obj.Mode.IsDir() {
			continue
		}
		objPrefix, objPattern := splitPath(postfixPath)
		objWorkPath := obj.Path + "/" + objPrefix
		ret, err := so.DirWalk(objWorkPath, objPattern)
		if err != nil {
			return nil, err
		}
		objects = append(objects, ret...)
	}

	return
}

func (so *SingleOperator) GlobList(path string) (objects []*types.Object, err error) {
	unixPath := strings.ReplaceAll(path, "\\", "/")

	workPath, pattern := splitPath(unixPath)
	return so.DirWalk(workPath, pattern)
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
