package hooks

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/WillAbides/idcheck"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
)

const (
	htmlHeader = "text/html"
	jsonHeader = "application/json"
)

var (
	static        = packr.NewBox("./static")
	tpl           = packr.NewBox("./tpl")
	queueTemplate = template.Must(template.New("").Parse(tpl.String("queue.gohtml")))
	indexTemplate = template.Must(template.New("").Parse(tpl.String("index.gohtml")))
)

type (
	queueTemplateData struct {
		QueueURL string
		Items    []string
	}

	//IDChecker checks queue IDs
	IDChecker interface {
		NewID() (*idcheck.ID, error)
		ValidID(*idcheck.ID) bool
	}

	//Service is a hooks service
	Service struct {
		publicURL          string
		queue              queue.Queue
		receivedAtOverride *time.Time
		idChecker          IDChecker
	}
)

//New returns a new hooks service
func New(queue queue.Queue, idChecker IDChecker, publicURL string) *Service {
	return &Service{
		idChecker: idChecker,
		queue:     queue,
		publicURL: publicURL,
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

//Router is the mux router
func (s *Service) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(w, struct{}{})
		if err != nil {
			http.Error(w, "failed serving html", http.StatusInternalServerError)
		}
	})
	r.HandleFunc("/_ping", s.pingHandler)

	r.HandleFunc("/q/new", s.newQueueHandler).Methods(http.MethodGet)

	sr := r.PathPrefix("/q/{key}").Subrouter()

	sr.HandleFunc("", s.postHandler).Methods(http.MethodPost)
	sr.HandleFunc("", s.peekHandler).Methods(http.MethodGet)
	sr.HandleFunc("/{subkey}", s.postHandler).Methods(http.MethodPost)
	sr.HandleFunc("/{subkey}", s.peekHandler).Methods(http.MethodGet)

	sr.Use(s.idCheckMiddleware)
	staticServer := http.FileServer(static)
	r.PathPrefix("/static").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		staticServer.ServeHTTP(w, r)
	})
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
	vars := mux.Vars(r)
	key := vars["key"]
	subkey := vars["subkey"]
	if subkey != "" {
		key = key + "/" + subkey
	}

	webRequest, err := queue.NewWebRequestFromHTTPRequest(r, s.receivedAt())
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

func probablyWantsHTML(r *http.Request) bool {
	accepts := textproto.MIMEHeader(r.Header)["Accept"]
	for _, accept := range accepts {
		if strings.Contains(strings.ToLower(accept), "html") {
			return true
		}
	}
	return false
}

func (s *Service) peekHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	subkey := vars["subkey"]
	if subkey != "" {
		key = key + "/" + subkey
	}

	webRequests, err := s.queue.Peek(r.Context(), key, 0)
	if err != nil {
		http.Error(w, "failed querying queue", http.StatusInternalServerError)
		return
	}

	if webRequests == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if probablyWantsHTML(r) {
		var items []string
		for _, item := range webRequests {
			jb, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				http.Error(w, "failed encoding json", http.StatusInternalServerError)
				return
			}
			items = append(items, string(jb))
		}
		w.Header().Set("Content-Type", htmlHeader)
		err := queueTemplate.Execute(w, queueTemplateData{
			QueueURL: strings.TrimRight(s.publicURL, "/") + "/q/" + key,
			Items:    items,
		})
		if err != nil {
			http.Error(w, "failed serving html", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", jsonHeader)
	err = json.NewEncoder(w).Encode(webRequests)
	if err != nil {
		http.Error(w, "failed encoding json", http.StatusInternalServerError)
		return
	}
}
