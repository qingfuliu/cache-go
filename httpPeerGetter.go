package cache_go

import (
	"cache-go/msg"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
)

type httpGetter struct {
	url string
}

func NewHttpGetter(url string) *httpGetter {
	return &httpGetter{
		url: url,
	}
}

func (h *httpGetter) Get(ctx context.Context, in *msg.GetRequest, out *msg.GetResponse) (err error) {
	var request *http.Request
	if request, err = http.NewRequest("GET", h.url, nil); err != nil {
		return
	}
	request = request.WithContext(ctx)
	request.Header.Add("key", in.Key)
	request.Header.Add("cacheName", in.CacheName)
	tr := http.DefaultTransport
	response, err := tr.RoundTrip(request)
	defer response.Body.Close()

	if err != nil {
		return err
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %v", response.Status)
	}

	var data []byte
	if data, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	err = proto.Unmarshal(data, out)

	return
}

func (h *httpGetter) Close() error {
	return nil
}
