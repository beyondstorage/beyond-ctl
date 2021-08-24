package operations

import "github.com/beyondstorage/go-storage/v4/types"

type SingleOperator struct {
	store types.Storager
	errch chan error
}

func NewSingleOperator(store types.Storager) (oo *SingleOperator, err error) {
	return &SingleOperator{store: store}, nil
}

func (oo *SingleOperator) Errors() chan error {
	return oo.errch
}
