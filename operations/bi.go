package operations

import "github.com/beyondstorage/go-storage/v4/types"

type BiOperator struct {
	src   types.Storager
	dst   types.Storager
	errCh chan error
}

func NewBiOperator(src, dst types.Storager) (oo *BiOperator, err error) {
	return &BiOperator{src: src, dst: dst}, nil
}

func (bo *BiOperator) Errors() chan error {
	return bo.errCh
}
