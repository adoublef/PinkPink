package nats

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Worker struct {
	sub *nats.Subscription
}

func NewWorker(nc *nats.Conn, ack nats.AckPolicy, maxPending int, consumer string, subj Subject) (*Worker, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nc.JetStream: %w", err)
	}

	sub, err := newWorker(js, ack, maxPending, consumer, subj.String())
	if err != nil {
		return nil, fmt.Errorf("newConsumer: %w", err)
	}

	return &Worker{sub}, nil
}

func (w *Worker) Close() error {
	if err := w.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("sub.Unsubscribe: %w", err)
	}
	return nil
}

func (w *Worker) NextMsg(ctx context.Context) (string, error) {
	msg, err := w.sub.NextMsgWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("sub.NextMsgWithContext: %w", err)
	}

	return string(msg.Data), msg.Ack()
}

// Actual handler I care for
func (w *Worker) Listen(ctx context.Context) error {
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

func newWorker(js nats.JetStreamContext, ack nats.AckPolicy, pending int, consumer, subject string) (*nats.Subscription, error) {
	// if consumer already exists return it
	if _, err := js.ConsumerInfo(streamName, consumer); err == nil {
		return js.QueueSubscribeSync(subject, consumer, nats.Bind(streamName, consumer))
	}

	_, err := js.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable: consumer,
		// convention is to use the same name as the durable
		DeliverGroup: consumer,
		// This for now can just be random, I don't really care
		DeliverSubject: nats.NewInbox(),
		FilterSubject:  subject,
		AckPolicy:      ack,
		MaxAckPending:  pending,
	})
	if err != nil {
		return nil, fmt.Errorf("js.AddConsumer: %w", err)
	}

	return js.QueueSubscribeSync(subject, consumer, nats.Bind(streamName, consumer))
}
