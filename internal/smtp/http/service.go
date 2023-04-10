package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyphengolang/with-jetstream/internal/openapi"
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
	// s.mux.Handle("/", svelte.FileServer("/")) // serve the svelte app

	s.mux.Handle("/api", openapi.FileServer("/"))
	s.mux.HandleFunc("/api/subscribe", s.handleSubscribe())
}

func (s *Service) handleSubscribe() http.HandlerFunc {
	type request struct {
		Email     smtp.Address `json:"email"`
		Subject   string       `json:"subject"`
		Message   string       `json:"message"`
		FirstName string       `json:"firstName"`
		LastName  string       `json:"lastName"`
	}

	type response struct {
		Message string `json:"message"`
	}

	parseEmail := func(w http.ResponseWriter, r *http.Request) (*smtp.Email, error) {
		var req request
		if err := s.decode(w, r, &req); err != nil {
			return nil, err
		}

		// subject must be between 1 to 50 characters
		if len(req.Subject) < 1 || len(req.Subject) > 50 {
			return nil, fmt.Errorf("subject must be between 1 to 50 characters")
		}

		// message must be between 1 to 255 characters
		if len(req.Message) < 1 || len(req.Message) > 255 {
			return nil, fmt.Errorf("message must be between 1 to 255 characters")
		}

		// firstname must be between 1 to 50 characters
		if len(req.FirstName) < 1 || len(req.FirstName) > 50 {
			return nil, fmt.Errorf("firstName must be between 1 to 50 characters")
		}

		// lastName must be between 1 to 50 characters if provided
		if len(req.LastName) > 0 && (len(req.LastName) < 1 || len(req.LastName) > 50) {
			return nil, fmt.Errorf("lastName must be between 1 to 50 characters if provided")
		}

		return &smtp.Email{
			Subject:    req.Subject,
			Message:    req.Message,
			Recipients: []smtp.Recipient{{Address: req.Email, FirstName: req.FirstName, LastName: req.LastName}},
		}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		e, err := parseEmail(w, r)
		if err != nil {
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
