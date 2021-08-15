package operations

import "github.com/beyondstorage/go-storage/v4/types"

type OneOperator struct {
	store types.Storager
	errch chan error
}

func NewOneOneOperator(store types.Storager) (oo *OneOperator, err error) {
	return &OneOperator{store: store}, nil
}

func (oo *OneOperator) Errors() chan error {
	return oo.errch
}
