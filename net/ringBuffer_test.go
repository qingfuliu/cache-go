package net

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	test := "1asdfasdfaer3245234562"
	ringBuffer := NewRingBuffer(1024)
	//fmt.Println(ringBuffer.Size())
	for i := 0; i < 100000; i++ {
		n, _ := ringBuffer.WriteString(test)
		if n != len(test) {
			t.Fatal("n should be same with len(test)")
		}
	}

	fmt.Println(ringBuffer.Len())

	if ringBuffer.IsEmpty() {
		t.Fatal("ringBuffer should be empty")
	}

	for i := 0; i < 100000; i++ {
		head, tail := ringBuffer.PeekN(len(test))
		if string(head)+string(tail) != test {
			t.Fatal("PeekN should be same with test")
		} else {
			ringBuffer.ShiftN(len(test))
		}
	}

	if !ringBuffer.IsEmpty() {
		t.Fatal("ringBuffer should be empty")
	}

	//
	for i := 0; i < 100000; i++ {
		_, _ = ringBuffer.WriteString(test)
	}
	fmt.Println(ringBuffer.Len())
	if ringBuffer.IsEmpty() {
		t.Fatal("ringBuffer should be empty")
	}
	testRead := []byte(test)
	for i := 0; i < 100000-1; i++ {
		n, _ := ringBuffer.Read(testRead)
		if n != len(testRead) {
			t.Fatal("n should be same with len(testRead)")
		}
		if string(testRead) != test {
			t.Fatal("testRead should be same with test")
		}
	}
	//
	if ringBuffer.IsEmpty() {
		t.Fatal("ringBuffer should be empty")
	}

	testReadWithbut := ringBuffer.ReadWithBytes([]byte(test))
	if testReadWithbut.String() != test+test {
		t.Fatal("testRead should be same with test")
	}

	if !ringBuffer.IsEmpty() {
		t.Fatal("ringBuffer should be empty")
	}
}

func TestRingBufferBigBytes(t *testing.T) {
	file, err := os.Open("/home/lqf/MPackage/protoc-3.20.1-linux-x86_64.zip")
	if err != nil {
		t.Fatal(err)
	}
	testRingBuffer := NewDefaultRingBuffer()
	testString, err := ioutil.ReadAll(file)
	fmt.Println(len(testString))
	var sum int
	for {

		if sum+MaximumQuantityPreTime < len(testString) {
			testRingBuffer.Write(testString[sum : sum+MaximumQuantityPreTime])
			fmt.Println(testRingBuffer.Len())
		} else {
			testRingBuffer.Write(testString[sum:])
			fmt.Println(testRingBuffer.Len())
			break
		}
		sum += MaximumQuantityPreTime
	}
	data := testRingBuffer.PeekAllWithBytes([]byte{})
	if !bytes.Equal(data.Bytes(), testString) {
		t.Fatal("err!")
	}

}
