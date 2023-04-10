package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	natsUtil "github.com/adoublef/pinkpink/internal/nats"
	smtpHTTP "github.com/adoublef/pinkpink/internal/smtp/http"
	smtpNATS "github.com/adoublef/pinkpink/internal/smtp/nats"
)

var (
	natsURL  = os.Getenv("NATS_URL")
	natsJWT  = os.Getenv("NATS_USER_JWT")
	natsNKey = os.Getenv("NATS_NKEY")

	port = os.Getenv("PORT")
)

func init() {
	log.SetPrefix("producer: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if port == "" {
		port = "8080"
	}
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

	p, err := smtpNATS.NewProducer(nc, 2, 1024)
	if err != nil {
		return fmt.Errorf("js.NewProducer: %w", err)
	}

	formatHTTP := smtpHTTP.New(p)

	srv := &http.Server{Addr: ":" + port, Handler: formatHTTP}

	e := make(chan error, 1)
	go func() { log.Printf("listening to :%s...", port); e <- srv.ListenAndServe() }()

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
