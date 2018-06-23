package hooks

import (
	"encoding/json"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Service struct {
	queue              queue.Queue
	receivedAtOverride int64
}

func New(queue queue.Queue) *Service {
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
	sr.HandleFunc("", s.peekHandler).Methods(http.MethodGet)
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

	err = s.queue.Push(r.Context(), key, []*queue.WebRequest{webRequest})
	if err != nil {
		http.Error(w, "failed adding to queue", http.StatusInternalServerError)
		return
	}
}

func (s *Service) peekHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	webRequests, err := s.queue.Peek(r.Context(), key, 0)
	if err != nil {
		http.Error(w, "failed querying queue", http.StatusInternalServerError)
		return
	}

	if webRequests == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		err = json.NewEncoder(w).Encode(webRequests)
		if err != nil {
			http.Error(w, "failed encoding json", http.StatusInternalServerError)
			return
		}
	}
}
