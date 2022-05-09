package net

import (
	"context"
	"github.com/golang/protobuf/proto"
)

type Conn interface {
	Get(ctx context.Context, in proto.Message, out proto.Message) error
	Chan() <-chan proto.Message
	Write(p []byte) (n int, err error)
	Close() error
}
