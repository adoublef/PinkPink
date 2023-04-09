package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	streamName = "FORMAT"
)

var (
	streamSubjects = []string{"format.>"}
)

type Subject string

func (s Subject) String() string {
	return "format." + string(s)
}

// subjects
const (
	SubjectAll Subject = ">"
	SubjectFoo Subject = "foo"
	SubjectBar Subject = "bar"
)

func validSubject(s Subject) error {
	switch s {
	case SubjectAll, SubjectFoo, SubjectBar:
		return nil
	}
	return fmt.Errorf("invalid subject: %s", s)
}

type Producer struct {
	nc nats.JetStreamContext
}

// publish event to stream
func (p *Producer) Publish(subject Subject, data []byte) error {
	if err := validSubject(subject); err != nil {
		return fmt.Errorf("validSubject: %w", err)
	}
	
	if _, err := p.nc.Publish(subject.String(), data); err != nil {
		return fmt.Errorf("nc.Publish: %w", err)
	}
	return nil
}

func NewProducer(nc *nats.Conn, retention nats.RetentionPolicy, maxBytes int64) (*Producer, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nc.JetStream: %w", err)
	}

	// check if stream already exists
	if _, err = js.StreamInfo(streamName); err == nil {
		return &Producer{js}, nil
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  streamSubjects, // wildcard
		Retention: retention,
		MaxBytes:  maxBytes, // 1kb
	}); err != nil {
		return nil, fmt.Errorf("addStream: %w", err)
	}

	return &Producer{js}, nil
}
