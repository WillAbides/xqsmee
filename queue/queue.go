package queue

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
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
		Pop(context.Context, string, int64) (*WebRequest, error)
		Push(context.Context, string, []*WebRequest) error
	}

	GRPCHandler struct {
		q Queue
	}
)

func NewGRPCHandler(q Queue) *GRPCHandler {
	return &GRPCHandler{q}
}

func (g *GRPCHandler) Pop(ctx context.Context, request *PopRequest) (*PopResponse, error) {
	webRequest, err := g.q.Pop(ctx, request.GetQueueName(), request.GetTimeout())
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

func NewWebRequestFromHttpRequest(req *http.Request, receivedAt int64) (*WebRequest, error) {
	if req == nil {
		return nil, ErrNilReq
	}
	body, err := readBodyFromHttpRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading request body")
	}
	return &WebRequest{
		ReceivedAt: receivedAt,
		Header:     getHeadersFromHttpRequest(req),
		Body:       body,
		Host:       req.Host,
	}, nil
}
