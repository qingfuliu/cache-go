package cache_go

import (
	"context"
	"sync"
)

func getPeerPicker() PeerPicker {
	return DefaultGetPeerPickerFunc()
}

var once sync.Once

func RegisterGetPeerPickerFunc(p PeerPicker) {
	once.Do(func() {
		DefaultGetPeerPickerFunc = func() PeerPicker {
			return p
		}
	})
}

var DefaultPeeker PeerPicker = &defaultPeeker{}

var DefaultGetPeerPickerFunc func() PeerPicker = func() PeerPicker {
	return DefaultPeeker
}

type PeerPicker interface {
	GetPeer(key string, ctx context.Context) (PeerGetter, bool)
}

type defaultPeeker struct {
}

func (p *defaultPeeker) GetPeer(key string, ctx context.Context) (PeerGetter, bool) {
	return nil, false
}
