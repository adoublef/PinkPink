package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

var (
	natsURL  = os.Getenv("NATS_URL")
	natsJWT  = os.Getenv("NATS_USER_JWT")
	natsNKey = os.Getenv("NATS_NKEY")
)

func init() {
	log.SetPrefix("producer: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)

	go func() { <-q; cancel() }()

	log.Printf("starting...")

	if err := run(ctx); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(ctx context.Context) error {
	nc, err := newNATSConnection(ctx, natsURL, natsJWT, natsNKey)
	if err != nil {
		return fmt.Errorf("newNATSConnection: %w", err)
	}
	defer nc.Close()

	// setup a jetstream context
	p, err := newProducer(nc, 2, 1024, "TEST", "test.>")
	if err != nil {
		return fmt.Errorf("newProducer: %w", err)
	}

	mux := http.NewServeMux()

	var count int
	mux.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		count++
		if _, err := p.nc.Publish("test.foo", []byte(fmt.Sprintf("message %d", count))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "published message %d", count)
	})

	srv := &http.Server{Addr: ":8080", Handler: mux}

	e := make(chan error, 1)
	go func() { log.Printf("server running..."); e <- srv.ListenAndServe() }()

	select {
	case err := <-e:
		return fmt.Errorf("srv.ListenAndServe: %w", err)
	case <-ctx.Done(): // graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("srv.Shutdown: %w", err)
		}

		log.Printf("shutting down...")
		return nil
	}
}

func newNATSConnection(ctx context.Context, natsURL, natsJWT, natsNKey string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL, nats.UserJWTAndSeed(natsJWT, natsNKey))
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return nc, nil
}

type producer struct {
	nc nats.JetStreamContext
}

func newProducer(nc *nats.Conn, retention nats.RetentionPolicy, maxBytes int64, stream string, subjects ...string) (*producer, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nc.JetStream: %w", err)
	}

	// check if stream already exists
	if _, err = js.StreamInfo(stream); err == nil {
		// update stream
		log.Printf("stream already exists: %s", stream)
		return &producer{js}, nil
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:      stream,
		Subjects:  subjects, // wildcard
		Retention: retention,
		MaxBytes:  maxBytes, // 1kb
	}); err != nil {
		return nil, fmt.Errorf("AddStream: %w", err)
	}

	log.Printf("stream created: %s", stream)
	return &producer{js}, nil
}
