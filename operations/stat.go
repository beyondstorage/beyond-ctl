package operations

import (
	"errors"
	"strings"

	"go.beyondstorage.io/v5/pairs"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) Stat(path string) (o *types.Object, err error) {
	o, err = so.store.Stat(path)

	// so.store.Features().VirtualDir
	if err != nil && errors.Is(err, services.ErrObjectNotExist) {
		it, cerr := so.store.List(path, pairs.WithListMode(types.ListModeDir))
		if cerr == nil {
			for {
				// FIXME: We should check if the directory exists by whether the object list is empty after bumping to the new version of services.
				obj, cerr := it.Next()
				if cerr != nil && errors.Is(cerr, types.IterateDone) {
					break
				}
				if cerr != nil {
					err = cerr
					break
				}
				if (obj.Mode.IsDir() && strings.HasPrefix(obj.Path, strings.TrimSuffix(path, "/")+"/")) ||
					(!obj.Mode.IsDir() && strings.HasPrefix(obj.Path, strings.TrimSuffix(path, "/")+"/")) {
					o = so.store.Create(path, pairs.WithObjectMode(types.ModeDir))
					err = nil
					break
				}
			}
		}
	}

	// in progress multipart upload
	if err != nil && errors.Is(err, services.ErrObjectNotExist) {
		// so.store.Features().CreateMultipart
		if _, ok := so.store.(types.Multiparter); ok {
			it, cerr := so.store.List(path, pairs.WithListMode(types.ListModePart))
			if cerr == nil {
				for {
					obj, cerr := it.Next()
					if cerr != nil {
						if !errors.Is(cerr, types.IterateDone) {
							err = cerr
						}
						break
					}
					if obj.Path == path {
						o = so.store.Create(path, pairs.WithMultipartID(obj.MustGetMultipartID()))
						err = nil
						break
					}
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return o, nil
}

func (so SingleOperator) StatStorager() (meta *types.StorageMeta) {
	meta = so.store.Metadata()
	return meta
}
