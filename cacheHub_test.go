package cache_go

import (
	"cache-go/byteString"
	"cache-go/msg"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

type testPeerPicker struct {
}

func (t *testPeerPicker) GetPeer(key string) (PeerGetter, bool) {
	return &testPeerGetter{}, true
}

type testPeerGetter struct {
}

var m = make(map[string]string)

func (t *testPeerGetter) Get(ctx context.Context, in proto.Message, out proto.Message) error {
	request := msg.GetRequest{}
	bytes, _ := proto.Marshal(in)
	_ = proto.Unmarshal(bytes, &request)
	if val, ok := m[request.Key]; ok {
		reponse := msg.GetResponse{
			Val: val,
		}
		bytes, _ = proto.Marshal(&reponse)
		fmt.Println(len(bytes))
		_ = proto.Unmarshal(bytes, out)
		return nil
	}
	return KeyDoesNotExists
}

func TestNewCacheHub(t *testing.T) {
	ctx := context.Background()
	RegisterGetPeerPickerFunc(func() peerPicker {
		return &testPeerPicker{}
	})

	for i := 0; i < 10; i++ {
		str := "lqf" + strconv.Itoa(i)
		m[str] = str
	}

	var getFunc GetterFunc = func(ctx context.Context, key string, skin byteString.Skin) error {
		if val, ok := m[key]; ok {
			skin.SetString(val)
			return nil
		}
		return KeyDoesNotExists
	}
	cache := NewCacheHub("test", getFunc, 300)
	//var rs []byte
	skiner := byteString.NewByteStringSkin()
	for i := 0; i < 10; i++ {
		str := "lqf" + strconv.Itoa(i)
		err := cache.Get(ctx, str, skiner)
		if err != nil || skiner.View().String() != str {
			//fmt.Println("address of rs is: ", unsafe.Pointer(&rs))
			fmt.Println(skiner.View().String(), str)
			fmt.Println(len(skiner.View().String()), len(str))
			t.Fatal(err)
		} else {
			fmt.Println(skiner.View().String())
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		for j := 0; j <= 100; j++ {
			wg.Add(1)
			go func(k int) {
				skiner := byteString.NewByteStringSkin()
				str := "lqf" + strconv.Itoa(k)
				err := cache.Get(ctx, str, skiner)
				if err != nil || skiner.View().String() != str {
					log.Fatal(str, "   ", err, "  ", skiner.View().String())
				} else {
					fmt.Println(skiner.View().String())
				}
				//if i > 3 && cache.localCache.nEliminate < 1 {
				//	t.Fatalf("cache.localCache.nEliminate should more then ,but is %d", cache.localCache.nEliminate)
				//}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()

	if cache.State().NumsGet() != 1020 {
		t.Fatalf("cache.State().NumsGet() should be 1010,but is %d", cache.State().NumsGet())
	}
	if cache.State().NumsHit() != 1020 {
		t.Fatalf("cache.State().NumsGet() should be 1010,but is %d", cache.State().NumsHit())
	}

	for i := 10; i < 20; i++ {
		for j := 0; j <= 100; j++ {
			wg.Add(1)
			go func(k int) {
				skiner := byteString.NewByteStringSkin()
				str := "lqf" + strconv.Itoa(k)
				err := cache.Get(ctx, str, skiner)
				if err == nil {
					log.Fatal("err should not be nil")
				}
				//if i > 3 && cache.localCache.nEliminate < 1 {
				//	t.Fatalf("cache.localCache.nEliminate should more then ,but is %d", cache.localCache.nEliminate)
				//}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()

	if cache.State().NumPeerGet() >= 1010 {
		t.Fatalf("cache.State().NumsGet() should be 1010,but is %d", cache.State().NumPeerGet())
	}

	if cache.State().NumsPeerMiss() >= 1010 {
		t.Fatalf("cache.State().NumsGet() should be 1010,but is %d", cache.State().NumsPeerMiss())
	}

	fmt.Println(runtime.NumCPU())

}
