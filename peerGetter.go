package cache_go

import (
	"context"
	"github.com/golang/protobuf/proto"
)

type peerGetter interface{
	Get(ctx context.Context,in proto.Message,out proto.Message)error
}