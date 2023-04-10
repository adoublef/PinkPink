package nats

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

// debug is used for testing
var debug bool

func init() {
	if  os.Getenv("DEBUG") == "t" {
		debug = true
	}
}

const (
	streamName = "FORMAT"
)

var (
	streamSubjects = []string{"*.smtp.send"}
)

type Subject string

func (s Subject) String() string {
	if debug {
		return "debug.smtp." + string(s)
	}
	
	return "internal.smtp." + string(s)
}

// subjects
const (
	SubjectAll Subject = ">"
	SubjectSend Subject = "send"
)

type Producer struct {
	nc nats.JetStreamContext
}

// publish event to stream
func (p *Producer) Publish(subject Subject, data []byte) error {
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
