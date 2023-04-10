package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hyphengolang/with-jetstream/internal/smtp"
	"github.com/nats-io/nats.go"
)

type Consumer interface {
	NextMsg(ctx context.Context) (*smtp.Email, error)
	Listen(ctx context.Context) error
	Close() error
}

var _ Consumer = (*Worker)(nil)

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

// I don't really need to know this info I don't think
func (w *Worker) NextMsg(ctx context.Context) (*smtp.Email, error) {
	msg, err := w.sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sub.NextMsgWithContext: %w", err)
	}

	var email smtp.Email
	if err := unmarshal(msg, &email); err != nil {
		return nil, fmt.Errorf("email.UnmarshalJSON: %w", err)
	}

	// do something with the email

	// ack the message
	if err := msg.Ack(); err != nil {
		return nil, fmt.Errorf("msg.Ack: %w", err)
	}

	return &email, nil
}

// Actual handler I care for
func (w *Worker) Listen(ctx context.Context) error {
	for {
		msg, err := w.sub.NextMsgWithContext(ctx)
		if err != nil {
			return fmt.Errorf("sub.NextMsgWithContext: %w", err)
		}

		var email smtp.Email
		if err := unmarshal(msg, &email); err != nil {
			return fmt.Errorf("email.UnmarshalJSON: %w", err)
		}

		// do something with the email

		// ack the message
		if err := msg.Ack(); err != nil {
			return fmt.Errorf("msg.Ack: %w", err)
		}
	}
}

func (w *Worker) Close() error {
	if err := w.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("sub.Unsubscribe: %w", err)
	}
	return nil
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

// unmarshal is a helper function to unmarshal the data
func unmarshal(msg *nats.Msg, v any) error {
	if err := json.Unmarshal(msg.Data, &v); err != nil {
		return fmt.Errorf("email.UnmarshalJSON: %w", err)
	}

	return nil
}

// marshal is a helper function to marshal the data
func marshal(v any) ([]byte, error) {
	p, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	return p, nil
}
