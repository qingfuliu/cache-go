package net

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAsyncQueue(t *testing.T) {
	const count = 5000
	var i int64
	task := func(i interface{}) error {
		k := i.(*int64)
		atomic.AddInt64(k, 1)
		return nil
	}
	queue := NewAsyncQueue()
	var wg sync.WaitGroup
	fn := func() {
		queue.Push(task, &i)
		wg.Done()
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go fn()
	}
	wg.Wait()
	if queue.Len() != count {
		t.Fatalf("queue length should be %d ,but is %d", count, queue.Len())
	}

	if loadNode(queue.head).next == nil {
		t.Fatalf("queue.head.next should not be nil")
	}

	if loadNode(queue.tail).next != nil {
		t.Fatalf("queue.head.next should  be nil")
	}

	test := loadNode(queue.head)
	k := 0
	for test != nil {
		k++
		test = loadNode(test.next)
	}
	if k != count+1 {
		t.Fatalf("k should be %d ,but is %d", count+1, k)
	}
	fn2 := func() {
		//_ = queue.Pop()
		node := queue.Pop()
		if node == nil {
			log.Fatalf("node should not be nil,ans the length of the queue is %d", queue.Len())
		}
		err := node.Run()
		if err != nil {
			t.Fatal(err)
		} else {
			fmt.Println(atomic.LoadInt64(&i))
		}
		wg.Done()
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go fn2()
	}

	wg.Wait()
	if i != int64(count) {
		t.Fatalf("i should be %d,but is %d", i, count)
	}

	if loadNode(queue.head).next != nil {
		t.Fatalf("queue.head.next should not be nil")
	}

	if loadNode(queue.tail).next != nil {
		t.Fatalf("queue.head.next should  be nil")
	}

}

func TestAsyncQueuePop(t *testing.T) {
	const count = 5000
	var i int64
	task := func(i interface{}) error {
		k := i.(*int64)
		atomic.AddInt64(k, 1)
		return nil
	}
	queue := NewAsyncQueue()
	var wg sync.WaitGroup
	fn := func() {
		queue.Push(task, &i)
		wg.Done()
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go fn()
	}
	wg.Wait()
	if queue.Len() != count {
		t.Fatalf("queue length should be %d ,but is %d", count, queue.Len())
	}

	if loadNode(queue.head).next == nil {
		t.Fatalf("queue.head.next should not be nil")
	}

	if loadNode(queue.tail).next != nil {
		t.Fatalf("queue.head.next should  be nil")
	}
	for i := 0; i < count+1000; i++ {
		n := queue.Pop()
		if i < count && n == nil {
			zap.L().Fatal("n should not be nil", zap.Int("i", i))
		} else if i > count && n != nil {
			zap.L().Fatal("n should  be nil", zap.Int("i", i))
		}
	}

	for i := 0; i < count; i++ {
		wg.Add(1)
		go fn()
	}
	wg.Wait()
	if queue.Len() != count {
		t.Fatalf("queue length should be %d ,but is %d", count, queue.Len())
	}

	if loadNode(queue.head).next == nil {
		t.Fatalf("queue.head.next should not be nil")
	}

	if loadNode(queue.tail).next != nil {
		t.Fatalf("queue.head.next should  be nil")
	}

	temp := queue.Pop()
	nums := 0
	for temp != nil {
		for i := 0; i < 200 && temp != nil; i++ {
			nums++
			temp = queue.Pop()
		}
	}

	if nums != count {
		t.Fatalf("nums should  be 1000")
	} else {
		zap.L().Debug("nums", zap.Int("nums", nums))
	}
}

func TestAsyncQueue2(t *testing.T) {
	const count = 100000
	var i int64
	task := func(i interface{}) error {
		k := i.(*int64)
		atomic.AddInt64(k, 1)
		return nil
	}
	queue := NewAsyncQueue()
	var wg sync.WaitGroup
	fn := func() {
		queue.Push(task, &i)
	}
	fn2 := func() {
		atomic.AddInt64(&i, -2)
		task := queue.Pop()
		task.Run()
		task = queue.Pop()
		task.Run()

		wg.Done()
		wg.Done()
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go fn()
		if i >= count/2 {
			go fn2()
		}
	}
	wg.Wait()
	if i != 0 {
		t.Fatalf("i %d", i)
	} else {
		fmt.Println(i)
	}
}
