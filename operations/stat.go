package operations

import (
	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) Stat(path string) (o *types.Object, err error) {
	o, err = so.store.Stat(path)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (so SingleOperator) StatStorager() (meta *types.StorageMeta) {
	meta = so.store.Metadata()
	return meta
}
