package cache_go

func GetPeerPicker() peerPicker {
	return DefaultGetPeerPickerFunc()
}

func RegisterGetPeerPickerFunc(fn func() peerPicker) {
	DefaultGetPeerPickerFunc = fn
}

var DefaultPeeker peerPicker = &defaultPeeker{}

var DefaultGetPeerPickerFunc func() peerPicker = func() peerPicker {
	return DefaultPeeker
}

type peerPicker interface {
	GetPeer(key string) (PeerGetter, bool)
}

type defaultPeeker struct {
}

func (p *defaultPeeker) GetPeer(key string) (PeerGetter, bool) {
	return nil, false
}
