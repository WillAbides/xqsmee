package queue

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
)

//go:generate protoc --go_out=plugins=grpc:. queue.proto
//go:generate mockgen -destination mockqueue/mockqueue.go -package mockqueue -source=queue.go

var (
	errInvalidArgument = errors.New("invalid argument")
	errNilReq          = errors.Wrap(errInvalidArgument, "req is nil")
)

type (
	//Queue is a queue
	Queue interface {
		Peek(context.Context, string, int64) ([]*WebRequest, error)
		Pop(context.Context, string, time.Duration) (*WebRequest, error)
		Push(context.Context, string, []*WebRequest) error
	}

	//GRPCHandler handle grpc requests
	GRPCHandler struct {
		q Queue
	}
)

//NewGRPCHandler returns a new GRPCHandler
func NewGRPCHandler(q Queue) *GRPCHandler {
	return &GRPCHandler{q: q}
}

//Pop pops an item off the queue
func (g *GRPCHandler) Pop(ctx context.Context, request *PopRequest) (*PopResponse, error) {
	var duration time.Duration
	timeout := request.GetTimeout()
	if timeout != nil {
		duration = time.Duration(timeout.GetNanos()) + time.Duration(timeout.GetSeconds())*time.Second
	}
	webRequest, err := g.q.Pop(ctx, request.GetQueueName(), duration)
	return &PopResponse{WebRequest: webRequest}, err
}

//Peek shows the next few items in the queue
func (g *GRPCHandler) Peek(ctx context.Context, request *PeekRequest) (*PeekResponse, error) {
	webRequests, err := g.q.Peek(ctx, request.GetQueueName(), request.GetCount())
	return &PeekResponse{WebRequest: webRequests}, err
}

func getHeadersFromHTTPRequest(req *http.Request) []*Header {
	headers := []*Header{}
	if req != nil {
		for k, v := range req.Header {
			headers = append(headers, &Header{Name: k, Value: v})
		}
	}
	return headers
}

func readBodyFromHTTPRequest(req *http.Request) (string, error) {
	if req == nil {
		return "", errNilReq
	}
	defer func() {
		err := req.Body.Close()
		if err != nil {
			log.Println("failed closing request body: ", err)
		}
	}()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed reading body")
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	return string(body), nil
}

//NewWebRequestFromHTTPRequest is a helper to build a WebRequest from an HTTP request
func NewWebRequestFromHTTPRequest(req *http.Request, receivedAt time.Time) (*WebRequest, error) {
	if req == nil {
		return nil, errNilReq
	}
	body, err := readBodyFromHTTPRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading request body")
	}
	ts, err := ptypes.TimestampProto(receivedAt)
	if err != nil {
		return nil, err
	}
	return &WebRequest{
		ReceivedAt: ts,
		Header:     getHeadersFromHTTPRequest(req),
		Body:       body,
		Host:       req.Host,
	}, nil
}

// MarshalJSON creates a json representation of q WebRequest
func (w *WebRequest) MarshalJSON() ([]byte, error) {
	s, err := new(jsonpb.Marshaler).MarshalToString(w)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

//UnmarshalJSON builds a WebRequest from JSON
func (w *WebRequest) UnmarshalJSON(src []byte) error {
	return jsonpb.UnmarshalString(string(src), w)
}
