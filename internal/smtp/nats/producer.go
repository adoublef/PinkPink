package nats

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

// debug is used for testing
var debug bool

// loc is the region location of server (e.g. us-west-2)
// 
// This will be found using the "FLY_REGION" environment variable
// var loc string 

func init() {
	if env := os.Getenv("DEBUG"); env != "" {
		debug = true
	}
}

const (
	streamName = "SMTP"
)

var (
	streamSubjects = []string{"*.smtp.subscribe"}
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
	SubjectAll       Subject = ">"
	SubjectSubscribe Subject = "subscribe"
)

type Producer interface {
	// will update so that the publisher 
	// will marshal the data to json
	Publish(subject Subject, data any) error
}

var _ Producer = (*Stream)(nil)

type Stream struct {
	nc nats.JetStreamContext
}

// publish event to stream.
// important to note that the data will need to be a pointer
func (s *Stream) Publish(subject Subject, data any) error {
	p, err := marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if _, err := s.nc.Publish(subject.String(), p); err != nil {
		return fmt.Errorf("nc.Publish: %w", err)
	}
	return nil
}

func NewProducer(nc *nats.Conn, retention nats.RetentionPolicy, maxBytes int64) (*Stream, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nc.JetStream: %w", err)
	}

	// TODO: update stream with new subjects if they change

	// check if stream already exists
	if _, err = js.StreamInfo(streamName); err == nil {
		return &Stream{js}, nil
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  streamSubjects, // wildcard
		Retention: retention,
		MaxBytes:  maxBytes, // 1kb
	}); err != nil {
		return nil, fmt.Errorf("addStream: %w", err)
	}

	return &Stream{js}, nil
}
