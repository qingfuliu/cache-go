package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

const (
	prime              = "16777619"
	defaultReplication = 100
)

type HashFunc func(str string) int32

type Option func(CMap *ConsistentMap)

func WithReplication(countReplication int) Option {
	return Option(func(cMap *ConsistentMap) {
		cMap.countReplication = countReplication
	})
}

func WithHash(hash func(str string) int32) Option {
	return Option(func(cMap *ConsistentMap) {
		cMap.hash = hash
	})
}

func defaultHash(str string) int32 {
	return int32(crc32.ChecksumIEEE([]byte(str)))
}

type ConsistentMap struct {
	m                map[string]struct{}
	rings            map[int32][]string
	mKey             []int32
	countReplication int
	hash             HashFunc
	mu               sync.RWMutex
}

func NewConsistMap() (cMap *ConsistentMap) {
	return &ConsistentMap{
		hash:             defaultHash,
		countReplication: defaultReplication,
		rings:            make(map[int32][]string),
		m:                make(map[string]struct{}),
	}
}

func NewConsistentMapWithOptions(options ...Option) (cMap *ConsistentMap) {
	cMap = NewConsistMap()
	for _, fn := range options {
		fn(cMap)
	}
	return
}

func (CMap *ConsistentMap) Add(key string) {
	CMap.AddWithReplication(key, CMap.countReplication)
}

func (CMap *ConsistentMap) AddWithReplication(key string, countReplication int) {
	if countReplication > CMap.countReplication {
		countReplication = CMap.countReplication
	}
	CMap.mu.Lock()
	defer CMap.mu.Unlock()
	CMap.del(key)
	for i := 0; i < countReplication; i++ {
		hashKey := CMap.hash(key + strconv.Itoa(i))
		CMap.rings[hashKey] = append(CMap.rings[hashKey], key)
	}
	CMap.m[key] = struct{}{}
	CMap.rearrangement()
}
func (CMap *ConsistentMap) Del(key string) {
	CMap.mu.Lock()
	defer CMap.mu.Unlock()
	CMap.del(key)
}
func (CMap *ConsistentMap) Get(val string) (string, bool) {
	CMap.mu.RLock()
	defer CMap.mu.RUnlock()

	if len(CMap.mKey) == 0 {
		return "", false
	}

	hashKey := CMap.hash(val)

	index := sort.Search(len(CMap.mKey), func(i int) bool {
		return CMap.mKey[i] >= hashKey
	})

	if index == len(CMap.mKey) {
		index = 0
	}

	nodeSlice := CMap.rings[CMap.mKey[index]]
	switch len(nodeSlice) {
	case 0:
		return "", false
	case 1:
		return nodeSlice[0], true
	default:
		index = int(CMap.hash(val + prime))
		index %= len(nodeSlice)
		return nodeSlice[index], true
	}
}
func (CMap *ConsistentMap) adjust() {
	sort.Slice(CMap.mKey, func(i, j int) bool {
		return CMap.mKey[i] < CMap.mKey[j]
	})
}
func (CMap *ConsistentMap) del(key string) {
	if _, ok := CMap.m[key]; !ok {
		return
	}
	for i := 0; i <= CMap.countReplication; i++ {
		hashKey := CMap.hash(key + strconv.Itoa(i))
		nodeSlice, ok := CMap.rings[hashKey]
		if ok {
			newNodeSlice := nodeSlice[:0]
			for i := 0; i < len(nodeSlice); i++ {
				if nodeSlice[i] != key {
					newNodeSlice = append(newNodeSlice, nodeSlice[i])
				}
			}
			if len(nodeSlice) == 0 {
				delete(CMap.rings, hashKey)
			} else {
				CMap.rings[hashKey] = newNodeSlice
			}
		}
	}
	CMap.rearrangement()
	delete(CMap.m, key)
}

func (CMap *ConsistentMap) rearrangement() {
	CMap.mKey = CMap.mKey[:0]
	for key, _ := range CMap.rings {
		CMap.mKey = append(CMap.mKey, key)
	}
	CMap.adjust()
}
