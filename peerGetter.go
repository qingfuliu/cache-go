package cache_go

import (
	"cache-go/msg"
	"context"
)

type PeerGetter interface {
	Get(ctx context.Context, in *msg.GetRequest, out *msg.GetResponse) error
	Close() error
}
