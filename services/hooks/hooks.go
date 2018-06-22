package hooks

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Service struct {
	queue              queue.QueueServer
	receivedAtOverride int64
}

func New(queue queue.QueueServer) *Service {
	return &Service{
		queue: queue,
	}
}

func (s *Service) receivedAt() int64 {
	if s.receivedAtOverride != 0 {
		return s.receivedAtOverride
	}
	return time.Now().UnixNano()
}

func (s *Service) Router() *mux.Router {
	r := mux.NewRouter()
	sr := r.PathPrefix("/{key}").Subrouter()
	sr.HandleFunc("", s.postHandler).Methods(http.MethodPost)
	sr.HandleFunc("/pop", s.popHandler).Methods(http.MethodGet)
	return r
}

func (s *Service) postHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	webRequest, err := queue.NewWebRequestFromHttpRequest(r, s.receivedAt())
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	pushRequest := &queue.PushRequest{
		QueueName:  key,
		WebRequest: []*queue.WebRequest{webRequest},
	}
	_, err = s.queue.Push(r.Context(), pushRequest)
	if err != nil {
		http.Error(w, "failed adding to queue", http.StatusInternalServerError)
		return
	}
}

func (s *Service) popHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	popRequest := &queue.PopRequest{QueueName: key}

	popResponse, err := s.queue.Pop(r.Context(), popRequest)
	if err != nil {
		http.Error(w, "failed querying queue", http.StatusInternalServerError)
		return
	}
	webRequest := popResponse.WebRequest

	if webRequest == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		err := new(jsonpb.Marshaler).Marshal(w, webRequest)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
