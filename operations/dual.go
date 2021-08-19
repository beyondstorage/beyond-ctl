package operations

import "github.com/beyondstorage/go-storage/v4/types"

type DualOperator struct {
	src   types.Storager
	dst   types.Storager
	errCh chan error
}

func NewDualOperator(src, dst types.Storager) (oo *DualOperator, err error) {
	return &DualOperator{
		src:   src,
		dst:   dst,
		errCh: make(chan error),
	}, nil
}

func (do *DualOperator) Errors() chan error {
	return do.errCh
}
