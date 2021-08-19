package operations

import "github.com/beyondstorage/go-storage/v4/types"

type SingleOperator struct {
	store types.Storager
	errCh chan error
}

func NewSingleOperator(store types.Storager) (oo *SingleOperator, err error) {
	return &SingleOperator{
		store: store,
		errCh: make(chan error),
	}, nil
}

func (so *SingleOperator) Errors() chan error {
	return so.errCh
}
