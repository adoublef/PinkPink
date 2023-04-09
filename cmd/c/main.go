package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	js "github.com/hyphengolang/with-jetstream/internal/format/nats"
	natsUtil "github.com/hyphengolang/with-jetstream/internal/nats"
)

var (
	natsURL  = os.Getenv("NATS_URL")
	natsJWT  = os.Getenv("NATS_USER_JWT")
	natsNKey = os.Getenv("NATS_NKEY")
)

func init() {
	log.SetPrefix("consumer: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)

	go func() { <-q; cancel() }()

	if err := run(ctx); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(ctx context.Context) error {
	nc, err := natsUtil.NewConn(ctx, natsURL, natsJWT, natsNKey)
	if err != nil {
		return fmt.Errorf("newNATSConnection: %w", err)
	}
	defer nc.Close()

	w, err := js.NewWorker(nc, 2, 1, "worker", js.SubjectFoo) // quantity of workers
	if err != nil {
		return fmt.Errorf("js.NewWorker: %w", err)
	}

	e := make(chan error, 1)
	go func() { log.Printf("consumer running..."); e <- w.Listen(ctx) }()

	select {
	case err := <-e:
		return err
	case <-ctx.Done(): // graceful shutdown
		log.Printf("shutting down...")
		if err := w.Close(); err != nil {
			return fmt.Errorf("w.Close: %w", err)
		}
		return nil
	}

}
