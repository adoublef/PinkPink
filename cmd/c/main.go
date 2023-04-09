package main

import (
	"context"
	"fmt"
	"log"
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
	log.SetPrefix("consumer: ")
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

	w, err := newWorker(nc, 2, 1, "worker", "test.foo") // quantity of workers
	if err != nil {
		return fmt.Errorf("newWorker: %w", err)
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

func newNATSConnection(ctx context.Context, natsURL, natsJWT, natsNKey string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL, nats.UserJWTAndSeed(natsJWT, natsNKey))
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return nc, nil
}

type worker struct {
	sub *nats.Subscription
}

func newWorker(nc *nats.Conn, ack nats.AckPolicy, maxPending int, consumer, subj string) (*worker, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nc.JetStream: %w", err)
	}

	// get stream name from subject
	info, err := streamInfoFromSubject(js, subj)
	if err != nil {
		return nil, fmt.Errorf("js.StreamNameBySubject: %w", err)
	}

	sub, err := newConsumer(js, ack, maxPending, consumer, subj, info)
	if err != nil {
		return nil, fmt.Errorf("newConsumer: %w", err)
	}

	return &worker{sub}, nil
}

func (w *worker) Close() error {
	if err := w.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("sub.Unsubscribe: %w", err)
	}
	return nil
}

func (w *worker) Listen(ctx context.Context) error {
	for {
		msg, err := w.sub.NextMsgWithContext(ctx)
		if err != nil {
			return fmt.Errorf("sub.NextMsgWithContext: %w", err)
		}

		log.Printf("received message: %s", string(msg.Data))

		// ack the message
		if err := msg.Ack(); err != nil {
			return fmt.Errorf("msg.Ack: %w", err)
		}
	}
}

func streamInfoFromSubject(js nats.JetStreamContext, subj string) (*nats.StreamInfo, error) {
	stream, err := js.StreamNameBySubject(subj)
	if err != nil {
		return nil, fmt.Errorf("js.StreamNameBySubject: %w", err)
	}

	info, err := js.StreamInfo(stream)
	if err != nil {
		return nil, fmt.Errorf("js.StreamInfo: %w", err)
	}

	return info, nil
}

func newConsumer(js nats.JetStreamContext, ack nats.AckPolicy, pending int, consumer, subj string, stream *nats.StreamInfo) (*nats.Subscription, error) {
	// if consumer already exists return it
	if info, err := js.ConsumerInfo(stream.Config.Name, consumer); err == nil {
		return js.QueueSubscribeSync(subj, consumer, nats.Bind(stream.Config.Name, info.Config.Durable))
	}

	info, err := js.AddConsumer(stream.Config.Name, &nats.ConsumerConfig{
		Durable: consumer,
		// convention is to use the same name as the durable
		DeliverGroup: consumer,
		// This for now can just be random, I don't really care
		DeliverSubject: nats.NewInbox(),
		FilterSubject:  subj,
		AckPolicy:      ack,
		MaxAckPending:  pending,
	})
	if err != nil {
		return nil, fmt.Errorf("js.AddConsumer: %w", err)
	}

	return js.QueueSubscribeSync(subj, consumer, nats.Bind(stream.Config.Name, info.Config.Durable))
}
