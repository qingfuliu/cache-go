package cache_go

import (
	"cache-go/consistentHash"
	"context"
)

type httpPeerPicker struct {
	cMap      *consistentHash.ConsistentMap
	peers     map[string]*httpGetter
	localAddr string
}

func NewHttpPeerPicker(localAddr string) *httpPeerPicker {
	hP := &httpPeerPicker{
		cMap:      consistentHash.NewConsistMap(),
		peers:     make(map[string]*httpGetter),
		localAddr: localAddr,
	}
	_ = hP.addAddr(hP.localAddr)
	RegisterGetPeerPickerFunc(hP)
	return hP
}

func (hP *httpPeerPicker) GetPeer(key string, _ context.Context) (PeerGetter, bool) {
	addr, ok := hP.cMap.Get(key)
	if !ok || addr == hP.localAddr {
		return nil, false
	}
	return hP.peers[addr], ok
}

func (hP *httpPeerPicker) addAddr(addr ...string) error {
	for _, val := range addr {
		if _, ok := hP.peers[val]; !ok {
			hP.cMap.Add(val)
			hP.peers[val] = NewHttpGetter(val)
		}
	}
	return nil
}
