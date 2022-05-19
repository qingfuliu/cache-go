package net

import (
	"cache-go/msg"
	_ "database/sql/driver"
	"sync"
)

type message struct {
	msg *msg.GetResponse
	err error
}
type tcpGetterAttachment struct {
	mu      *sync.Mutex
	msgChan chan *message
	wait    int32
}
