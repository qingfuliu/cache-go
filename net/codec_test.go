package net

import (
	"bytes"
	"cache-go/net/pool/slicePool"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"testing"
	"time"
)

type testConn struct {
	bytes []byte
	size  int
}

func (tC *testConn) Get(ctx context.Context, in proto.Message, out proto.Message) error {
	return nil
}
func (tC *testConn) Chan() <-chan proto.Message {
	return nil
}
func (tC *testConn) Write(p []byte) (n int, err error) {
	return 0, nil
}
func (tC *testConn) Close() error {
	return nil
}
func (tC *testConn) PeekAll() []byte {
	return tC.bytes
}
func (tC *testConn) ShiftN(n int) {
	tC.bytes = tC.bytes[n:]
}

func TestLengthFieldBasedFrameCodec(t *testing.T) {
	//-----------------lengthField=16----------------------//
	{
		data := make([]byte, 0, 256)
		for i := 0; i <= 255; i++ {
			data = append(data, byte(i))
		}
		codec := NewLengthFieldBasedFrameCodec(&EncoderConfig{
			IncludeLengthFieldLength: false,
			LengthAdjustment:         0,
			LengthFieldLength:        16,
		})
		tC := &testConn{}

		var err error
		tC.bytes, err = codec.Encode(data)
		if err != nil {
			t.Fatal(err)
		}
		var ans []byte
		ans, err = codec.Decode(tC)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(ans, data) {
			fmt.Println("ans", ans)
			fmt.Println("data", data)
			t.Fatalf("ans should be same as data")
		}

		if len(tC.bytes) != 0 {
			t.Fatalf("len(tC.bytes) should be 0")
		}
	}
	{
		nums := rand.Intn(256)
		data := make([]byte, 0, nums)
		for i := 1; i <= nums; i++ {
			data = append(data, byte(i))
		}
		codec := NewLengthFieldBasedFrameCodec(&EncoderConfig{
			IncludeLengthFieldLength: true,
			LengthAdjustment:         100,
			LengthFieldLength:        16,
		})
		tC := &testConn{}

		loop := 100

		var err error
		var nbyte []byte
		for i := 0; i < loop; i++ {
			nbyte, err = codec.Encode(data)
			if err != nil {
				t.Fatal(err)
			} else {
				tC.bytes = append(tC.bytes, nbyte...)
			}
		}

		for i := 0; i < loop; i++ {
			var ans []byte
			ans, err = codec.Decode(tC)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ans, data) {
				fmt.Println("ans", ans)
				fmt.Println("data", data)
				t.Fatalf("ans should be same as data")
			}
		}
		if len(tC.bytes) != 0 {
			t.Fatalf("len(tC.bytes) should be 0")
		}
	}
}

func TestLengthFieldBasedFrameCodec64(t *testing.T) {
	rand.Seed(time.Now().UnixMilli())
	//-----------------lengthField=64----------------------//
	{
		nums := rand.Intn(1 << 25)
		nums = 91005025
		fmt.Println(nums)

		data := slicePool.Get(nums)
		for i := 0; i < nums; i++ {
			data[i] = byte(i)
		}
		codec := NewLengthFieldBasedFrameCodec(&EncoderConfig{
			IncludeLengthFieldLength: false,
			LengthAdjustment:         0,
			LengthFieldLength:        64,
		})
		tC := &testConn{}

		loop := 10
		var err error
		var nbyte []byte

		tC.bytes = slicePool.Get(loop * (nums + 8))

		for i := 0; i < loop; i++ {
			nbyte, err = codec.Encode(data)
			if err != nil {
				t.Fatal(err)
			} else {
				copy(tC.bytes[tC.size:], nbyte)
				tC.size += len(nbyte)
			}

			slicePool.Put(nbyte)
		}

		for i := 0; i < loop; i++ {
			var ans []byte
			ans, err = codec.Decode(tC)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ans, data) {
				fmt.Println("ans", ans)
				fmt.Println("data", data)
				t.Fatalf("ans should be same as data")
			}
			slicePool.Put(ans)
		}

		if len(tC.bytes) != 0 {
			t.Fatalf("len(tC.bytes) should be 0,but is %d", len(tC.bytes))
		}

	}
}

func TestLengthFieldBasedFrameCodec64NoPool(t *testing.T) {
	rand.Seed(time.Now().UnixMilli())
	//-----------------lengthField=64----------------------//
	{
		nums := rand.Intn(1 << 27)
		nums = 91005025
		fmt.Println(nums)

		data := make([]byte, nums)
		for i := 0; i < nums; i++ {
			data[i] = byte(i)
		}
		codec := NewLengthFieldBasedFrameCodec(&EncoderConfig{
			IncludeLengthFieldLength: false,
			LengthAdjustment:         0,
			LengthFieldLength:        64,
		})
		tC := &testConn{}

		loop := 10
		var err error
		var nbyte []byte

		tC.bytes = make([]byte, loop*(nums+8))

		for i := 0; i < loop; i++ {
			nbyte, err = codec.Encode(data)
			if err != nil {
				t.Fatal(err)
			} else {
				copy(tC.bytes[tC.size:], nbyte)
				tC.size += len(nbyte)
			}

			//slicePool.Put(nbyte)
		}

		for i := 0; i < loop; i++ {
			var ans []byte
			ans, err = codec.Decode(tC)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ans, data) {
				fmt.Println("ans", ans)
				fmt.Println("data", data)
				t.Fatalf("ans should be same as data")
			}
			//slicePool.Put(ans)
		}

		if len(tC.bytes) != 0 {
			t.Fatalf("len(tC.bytes) should be 0,but is %d", len(tC.bytes))
		}

	}
}
