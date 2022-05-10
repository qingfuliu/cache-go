package net

import (
	"fmt"
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
