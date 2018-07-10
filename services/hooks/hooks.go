package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/WillAbides/idcheck"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/gorilla/mux"
)

type (
	IDChecker interface {
		NewID() (*idcheck.ID, error)
		ValidID(*idcheck.ID) bool
	}

	Service struct {
		queue              queue.Queue
		receivedAtOverride *time.Time
		idChecker          IDChecker
	}
)

func New(queue queue.Queue, idChecker IDChecker) *Service {
	return &Service{
		idChecker: idChecker,
		queue:     queue,
	}
}

func (s *Service) receivedAt() time.Time {
	if s.receivedAtOverride != nil {
		return *s.receivedAtOverride
	}
	return time.Now()
}

func (s *Service) idCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := idcheck.FromBase64(mux.Vars(r)["key"])
		if err != nil || !s.idChecker.ValidID(id) {
			http.NotFound(w, r)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (s *Service) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/_ping", s.pingHandler)

	r.HandleFunc("/q/new", s.newQueueHandler).Methods(http.MethodGet)

	sr := r.PathPrefix("/q/{key}").Subrouter()

	sr.HandleFunc("", s.postHandler).Methods(http.MethodPost)
	sr.HandleFunc("", s.peekHandler).Methods(http.MethodGet)
	sr.Use(s.idCheckMiddleware)
	return r
}

func (s *Service) pingHandler(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "pong")
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (s *Service) newQueueHandler(w http.ResponseWriter, r *http.Request) {
	id, err := s.idChecker.NewID()
	if err != nil {
		http.Error(w, "failed to create new queue id", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/q/"+id.Base64(), http.StatusFound)
}

func (s *Service) postHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	webRequest, err := queue.NewWebRequestFromHttpRequest(r, s.receivedAt())
	if err != nil || key == "" {
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
