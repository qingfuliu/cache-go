package slicePool

import (
	"fmt"
	"math/bits"
	"testing"
)

func TestPool_Get(t *testing.T) {
	index := func(n uint32) uint32 {
		return uint32(bits.Len32(n - 1))
	}

	fmt.Println(index(1))
	fmt.Println(index(2))
	fmt.Println(index(3))
	fmt.Println(index(4))
	fmt.Println(index(5))
	fmt.Println(index(6))
	fmt.Println(index(7))
	fmt.Println(index(8))

}
