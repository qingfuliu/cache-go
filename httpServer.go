package cache_go

import (
	"cache-go/byteString"
	"cache-go/msg"
	"context"
	"github.com/golang/protobuf/proto"
	"net/http"
)

var DefaultRouterPath = "cache_go"

type httpServer struct {
	localAddr string
	hP        *httpPeerPicker
}

func NewHttpServer(localAddr string) *httpServer {
	return &httpServer{
		localAddr: localAddr,
		hP:        NewHttpPeerPicker(localAddr),
	}
}

func (hs *httpServer) Start(localAddr string) {
	_ = http.ListenAndServe(localAddr, hs)
}

func (hs *httpServer) ServeHTTP(rw http.ResponseWriter, re *http.Request) {
	var (
		getResponse    *msg.GetResponse
		message        []byte
		key, cacheName string
		err            error
	)

	getResponse = &msg.GetResponse{}
	key, cacheName = re.Header.Get("key"), re.Header.Get("cacheName")

	var val string
	err = Get(context.Background(), key, cacheName, byteString.NewStringSkin(&val))
	getResponse.Error = err.Error()
	getResponse.Val = val

	message, _ = proto.Marshal(getResponse)
	_, _ = rw.Write(message)
	rw.WriteHeader(http.StatusOK)
	return
}
