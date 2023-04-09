package http

import (
	"fmt"
	"net/http"

	js "github.com/hyphengolang/with-jetstream/internal/format/nats"
)

type Service struct {
	mux *http.ServeMux
	p   *js.Producer
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func New(p *js.Producer) (*Service) {
	s := &Service{
		mux: http.NewServeMux(),
		p:   p,
	}

	s.routes()

	return s
}

func (s *Service) routes() {
	s.mux.HandleFunc("/format", s.handleCount(0))
}

func (s *Service) handleCount(count int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count++
		if err := s.p.Publish(js.SubjectFoo, []byte(fmt.Sprintf("message %d", count))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
