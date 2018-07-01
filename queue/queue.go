package queue

import (
	"bytes"
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

//go:generate protoc --go_out=plugins=grpc:. queue.proto
//go:generate mockgen -destination mockqueue/mockqueue.go -package mockqueue -source=queue.go

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNilReq          = errors.Wrap(ErrInvalidArgument, "req is nil")
)

type (
	Queue interface {
		Peek(context.Context, string, int64) ([]*WebRequest, error)
		Pop(context.Context, string, time.Duration) (*WebRequest, error)
		Push(context.Context, string, []*WebRequest) error
	}

	GRPCHandler struct {
		q Queue
	}
)

func NewGRPCHandler(q Queue) *GRPCHandler {
	return &GRPCHandler{q: q}
}

func (g *GRPCHandler) Pop(ctx context.Context, request *PopRequest) (*PopResponse, error) {
	var duration time.Duration
	timeout := request.GetTimeout()
	if timeout != nil {
		duration = time.Duration(time.Duration(timeout.GetNanos()) + time.Duration(timeout.GetSeconds())*time.Second)
	}
	webRequest, err := g.q.Pop(ctx, request.GetQueueName(), duration)
	return &PopResponse{WebRequest: webRequest}, err
}

func (g *GRPCHandler) Peek(ctx context.Context, request *PeekRequest) (*PeekResponse, error) {
	webRequests, err := g.q.Peek(ctx, request.GetQueueName(), request.GetCount())
	return &PeekResponse{WebRequest: webRequests}, err
}

func getHeadersFromHttpRequest(req *http.Request) []*Header {
	headers := []*Header{}
	if req != nil {
		for k, v := range req.Header {
			headers = append(headers, &Header{Name: k, Value: v})
		}
	}
	return headers
}

func readBodyFromHttpRequest(req *http.Request) (string, error) {
	if req == nil {
		return "", ErrNilReq
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed reading body")
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	return string(body), nil
}

func NewWebRequestFromHttpRequest(req *http.Request, receivedAt time.Time) (*WebRequest, error) {
	if req == nil {
		return nil, ErrNilReq
	}
	body, err := readBodyFromHttpRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading request body")
	}
	ts, err := ptypes.TimestampProto(receivedAt)
	if err != nil {
		return nil, err
	}
	return &WebRequest{
		ReceivedAt: ts,
		Header:     getHeadersFromHttpRequest(req),
		Body:       body,
		Host:       req.Host,
	}, nil
}

func (w *WebRequest) MarshalJSON() ([]byte, error) {
	s, err := new(jsonpb.Marshaler).MarshalToString(w)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (w *WebRequest) UnmarshalJSON(src []byte) error {
	return jsonpb.UnmarshalString(string(src), w)
}
