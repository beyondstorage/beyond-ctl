package operations

import (
	"fmt"

	"github.com/beyondstorage/go-storage/v4/types"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type SingleOperator struct {
	store types.Storager
}

func NewSingleOperator(store types.Storager) (oo *SingleOperator) {
	return &SingleOperator{store: store}
}

type DualOperator struct {
	src        types.Storager
	dst        types.Storager
	readPairs  []types.Pair
	writePairs []types.Pair
	pool       *ants.Pool
	logger     *zap.Logger
}

func NewDualOperator(src, dst types.Storager) (do *DualOperator) {
	// TODO: we will support setting workers via command line.
	pool, err := ants.NewPool(4)
	if err != nil {
		panic(fmt.Errorf("inti worker pool: %w", err))
	}

	// TODO: we will allow user config log level.
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}

	return &DualOperator{
		src:    src,
		dst:    dst,
		pool:   pool,
		logger: logger,
	}
}

func (do *DualOperator) WithWorkers(workers int) *DualOperator {
	pool, err := ants.NewPool(workers)
	if err != nil {
		panic(fmt.Errorf("inti worker pool: %w", err))
	}

	do.pool = pool
	return do
}

func (do *DualOperator) WithReadPairs(ps ...types.Pair) *DualOperator {
	do.readPairs = ps
	return do
}

func (do *DualOperator) WithWritePairs(ps ...types.Pair) *DualOperator {
	do.writePairs = ps
	return do
}
