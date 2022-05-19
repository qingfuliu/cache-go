package cache_go

import "context"

type PeerConnector interface {
	Connect(ctx context.Context) (PeerGetter, error)
}
