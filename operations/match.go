package operations

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"

	"go.beyondstorage.io/v5/types"
)

// Glob returns the names of all files matching pattern or nil
// if there is no matching file.
func (so *SingleOperator) Glob(path string) (matches []*types.Object, err error) {
	unixPath := filepath.ToSlash(path)
	base, _ := doublestar.SplitPattern(unixPath)
	if base == "." {
		base = ""
	}

	ch, err := so.ListRecursively(base)
	if err != nil {
		return
	}

	for v := range ch {
		if v.Error != nil {
			return nil, v.Error
		}

		if ok, _ := doublestar.Match(unixPath, v.Object.Path); ok {
			matches = append(matches, v.Object)
		}
	}

	return matches, nil
}
