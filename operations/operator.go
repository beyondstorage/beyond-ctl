package operations

import (
	"fmt"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"

	"go.beyondstorage.io/v5/types"
)

type SingleOperator struct {
	store  types.Storager
	pool   *ants.Pool
	logger *zap.Logger
}

func NewSingleOperator(store types.Storager) (oo *SingleOperator) {
	pool, err := ants.NewPool(4)
	if err != nil {
		panic(fmt.Errorf("inti worker pool: %w", err))
	}

	// TODO: we will allow user config log level.
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}

	return &SingleOperator{
		store:  store,
		pool:   pool,
		logger: logger,
	}
}

func (so *SingleOperator) WithWorkers(workers int) *SingleOperator {
	pool, err := ants.NewPool(workers)
	if err != nil {
		panic(fmt.Errorf("inti worker pool: %w", err))
	}

	so.pool = pool
	return so
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
