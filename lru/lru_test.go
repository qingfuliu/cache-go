package lru

import (
	"fmt"
	"sync/atomic"
	"testing"
)

var data_test []int = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

func TestCacheLRU(t *testing.T) {

	var i int32
	var capacity int
	capacity = 5

	cache := NewCacheLru(capacity, func(key, value interface{}) {
		fmt.Println("key: ", key, " value: ", value)
		atomic.AddInt32(&i, 1)
	})
	for i := range data_test {
		cache.Add(i, data_test[i])
	}
	fmt.Println(cache.Len())
	if len(data_test) > capacity && i != int32(len(data_test)-capacity) {
		t.Fatalf("i shoule be %d !!,but is %d", len(data_test)-capacity, i)
	}

	for i := range data_test {
		if i < 5 {
			if _, ok := cache.Get(i); ok {
				t.Fatalf("should not be ok!!!")
			}
		} else {
			if val, ok := cache.Get(i); ok {
				if ii, ok := val.(int); ok {
					if ii == i {
						k := cache.l.Front().Value.(*Entry).value
						if k != ii {
							t.Fatalf("kk should  be equal with i!!!,but k is %d i is %d, and the key is %d", k, i, cache.l.Front().Value.(*Entry).key)
						}
					} else {
						t.Fatalf("ii should  be equal with i!!!,but ii is %d i is %d", ii, i)
					}
				} else {
					t.Fatalf("should  be ok!!!")
				}
			} else {
				t.Fatalf("should  be ok!!!")
			}
		}
	}

}
