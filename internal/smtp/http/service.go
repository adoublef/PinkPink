package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyphengolang/with-jetstream/internal/smtp"
	smtpNATS "github.com/hyphengolang/with-jetstream/internal/smtp/nats"
)

type Service struct {
	mux *http.ServeMux
	p   smtpNATS.Producer
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func New(p smtpNATS.Producer) *Service {
	s := &Service{
		mux: http.NewServeMux(),
		p:   p,
	}

	s.routes()

	return s
}

func (s *Service) routes() {
	s.mux.HandleFunc("/subscribe", s.handleCount(0))
}

func (s *Service) handleCount(count int) http.HandlerFunc {
	type response struct {
		Message string `json:"message"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var e smtp.Email
		if err := s.decode(w, r, &e); err != nil {
			s.error(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.p.Publish(smtpNATS.SubjectSubscribe, e); err != nil {
			s.error(w, r, err, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, &response{Message: "Email sent"}, http.StatusOK)
	}
}

func (s *Service) decode(w http.ResponseWriter, r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return fmt.Errorf("json.NewDecoder: %w", err)
	}
	return nil
}

func (s *Service) respond(w http.ResponseWriter, r *http.Request, v any, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			http.Error(w, "Could not encode in json", status)
		}
	}
}

func (s *Service) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	http.Error(w, err.Error(), status)
}
