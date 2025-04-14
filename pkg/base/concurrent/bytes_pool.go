package concurrent

import (
	"fmt"
	"sort"
)

var (
	ErrBufferTooLarge    = fmt.Errorf("requested buffer size exceeds maximum pool size")
	ErrInvalidBufferSize = fmt.Errorf("buffer size does not match any pool size")
)

type BytesPool struct {
	sizes []int
	pools []*Pool[[]byte]
}

func NewBytesPool(sizes []int) *BytesPool {
	sort.Ints(sizes)
	pools := make([]*Pool[[]byte], len(sizes))

	for i, size := range sizes {
		pools[i] = NewPool(func() []byte {
			return make([]byte, size)
		})
	}

	return &BytesPool{
		sizes: sizes,
		pools: pools,
	}
}

func (bp *BytesPool) Get(size int) ([]byte, error) {
	idx := sort.Search(len(bp.sizes), func(i int) bool {
		return size <= bp.sizes[i]
	})

	if idx >= len(bp.sizes) {
		return nil, ErrBufferTooLarge
	}

	return bp.pools[idx].Get()[:size], nil
}

func (bp *BytesPool) Put(b []byte) error {
	size := cap(b)

	idx := sort.Search(len(bp.sizes), func(i int) bool {
		return size == bp.sizes[i]
	})

	if idx >= len(bp.sizes) {
		return ErrInvalidBufferSize
	}

	bp.pools[idx].Put(b)
	return nil
}
