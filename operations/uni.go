package operations

import "github.com/beyondstorage/go-storage/v4/types"

type UniOperator struct {
	store types.Storager
	errCh chan error
}

func NewUniOperator(store types.Storager) (oo *UniOperator, err error) {
	return &UniOperator{store: store}, nil
}

func (uo *UniOperator) Errors() chan error {
	return uo.errCh
}
