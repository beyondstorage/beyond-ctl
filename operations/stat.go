package operations

import (
	"github.com/beyondstorage/go-storage/v4/types"
)

func (oo *SingleOperator) Stat(path string) (o *types.Object, err error) {
	o, err = oo.store.Stat(path)
	if err != nil {
		return nil, err
	}

	return o, nil
}
