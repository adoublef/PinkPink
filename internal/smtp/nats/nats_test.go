package nats_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	smtpNATS "github.com/hyphengolang/with-jetstream/internal/smtp/nats"
	containers "github.com/hyphengolang/with-jetstream/pkg/containers/nats"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	natsContainer *containers.NATSContainer
	natsConn      *nats.Conn
)

func TestConsumer(t *testing.T) {
	// we are testing so should be in debug mode
	t.Setenv("DEBUG", "t")

	// setup nats producer
	producer, err := smtpNATS.NewProducer(natsConn, nats.WorkQueuePolicy, 1024)
	require.NoError(t, err, "failed to create nats producer")

	// setup nats consumer
	consumer, err := smtpNATS.NewWorker(natsConn, nats.AckExplicitPolicy, 1, "tester", smtpNATS.SubjectSend)
	require.NoError(t, err, "failed to create nats consumer")

	err = producer.Publish(smtpNATS.SubjectSend, []byte("hello world"))
	require.NoError(t, err, "failed to publish message")

	// TODO: wait for consumer to receive message
	msg, err := consumer.NextMsg(context.Background())
	require.NoError(t, err, "failed to get next message")

	require.Equal(t, "hello world", string(msg), "message data does not match")
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	if err := setup(ctx); err != nil {
		log.Fatalf("setup: %v", err)
	}

	code := m.Run()

	if err := teardown(ctx); err != nil {
		log.Fatalf("teardown: %v", err)
	}

	os.Exit(code)
}

func setup(ctx context.Context) (err error) {
	natsContainer, err = containers.RunNATSContainer(ctx,
		containers.WithWaitStrategy(wait.ForLog("Listening for client connections on 0.0.0.0:4222")))
	if err != nil {
		return fmt.Errorf("failed to run nats container: %w", err)
	}

	// get connection string
	url, err := natsContainer.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	// connect to nats
	natsConn, err = nats.Connect(url)
	if err != nil {
		return fmt.Errorf("failed to connect to nats: %w", err)
	}

	return nil
}

func teardown(ctx context.Context) error {
	return natsContainer.Terminate(ctx)
}
