package lru

import (
	"container/list"
)

type CacheLRU struct {
	m        map[interface{}]*list.Element
	l        *list.List
	capacity int
	OnDelete func(key, value interface{})
}

var funcDelete = func(key, value interface{}) {}

func NewCacheLru(capacity int, OnDelete func(key, value interface{})) *CacheLRU {
	if capacity <= 0 {
		capacity = 100
	}
	if OnDelete == nil {
		OnDelete = funcDelete
	}
	return &CacheLRU{
		capacity: capacity,
		OnDelete: OnDelete,
	}
}

type Entry struct {
	key, value interface{}
}

func (c *CacheLRU) Get(key interface{}) (interface{}, bool) {
	if c.l == nil {
		return nil, false
	}
	if e, ok := c.m[key]; ok {
		c.l.MoveToFront(e)
		return e.Value.(*Entry).value, true
	}
	return nil, false
}
func (c *CacheLRU) Add(key, value interface{}) {
	if c.l == nil {
		c.l = list.New()
		c.m = make(map[interface{}]*list.Element)
	}
	e, ok := c.m[key]
	if ok {
		e.Value.(*Entry).value = value
		c.l.MoveToFront(e)
		return
	}
	if c.capacity <= c.Len() {
		c.RemoveOldest()
	}
	ee := &Entry{key, value}
	e = c.l.PushFront(ee)
	c.m[key] = e
	return
}
func (c *CacheLRU) Remove(key interface{}) bool {
	if c.l == nil {
		return false
	}
	if e, ok := c.m[key]; ok {
		c.removeEntry(e)
	}
	return false
}

func (c *CacheLRU) Len() int {
	if c.l == nil {
		return 0
	}
	return c.l.Len()
}

func (c *CacheLRU) Release() {
	if c.l == nil {
		return
	}
	if c.OnDelete != nil {
		for _, val := range c.m {
			c.OnDelete(val.Value.(*Entry).key, val.Value.(*Entry).value)
		}
	}
	c.l, c.m = nil, nil
}

func (c *CacheLRU) RemoveOldest() interface{} {
	e := c.l.Back()
	c.removeEntry(e)
	return e.Value.(*Entry).value
}

func (c *CacheLRU) removeEntry(e *list.Element) {
	c.l.Remove(e)
	delete(c.m, e.Value.(*Entry).key)
	if c.OnDelete != nil {
		c.OnDelete(e.Value.(*Entry).key, e.Value.(*Entry).value)
	}
}
