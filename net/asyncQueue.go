package net

import (
	"sync/atomic"
	"unsafe"
)

type TaskFunc func(interface{}) error

type Task struct {
	task TaskFunc
	arg  interface{}
}

func (t *Task) Run() error {
	return t.task(t.arg)
}

type node struct {
	next unsafe.Pointer
	//prev unsafe.Pointer
	task *Task
}

func loadNode(ptr unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(&ptr))
}

type AsyncQueue struct {
	head   unsafe.Pointer
	tail   unsafe.Pointer
	length int64
}

func NewAsyncQueue() *AsyncQueue {
	n := unsafe.Pointer(&node{})
	return &AsyncQueue{
		head: n,
		tail: n,
	}
}

func (q *AsyncQueue) Len() int {
	return int(atomic.LoadInt64(&q.length))
}
func (q *AsyncQueue) IsEmpty() bool {
	return q.head == q.tail
}
func (q *AsyncQueue) Push(fn TaskFunc, arg interface{}) *Task {
	n := &node{
		task: &Task{
			task: fn,
			arg:  arg,
		},
	}
	tail := loadNode(q.tail)
	next := loadNode(tail.next)
retry:
	if next == nil {
		if loadNode(q.tail) == tail {
			if atomic.CompareAndSwapPointer(&tail.next, nil, unsafe.Pointer(n)) {
				atomic.AddInt64(&q.length, 1)
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(n))
				return n.task
			}
		}
	} else {
		atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
	}
	tail = loadNode(q.tail)
	next = loadNode(tail.next)
	goto retry
}

func (q *AsyncQueue) Pop() *Task {
	head := loadNode(q.head)
	tail := loadNode(q.tail)
	next := loadNode(head.next)
retry:
	if loadNode(q.head) == head {
		if head == tail {
			if next == nil {
				return nil
			}
			atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
		} else {
			n := next
			if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
				head.next = nil
				atomic.AddInt64(&q.length, -1)
				return n.task
			}
		}
	}
	head = loadNode(q.head)
	tail = loadNode(q.tail)
	next = loadNode(head.next)
	goto retry
}
