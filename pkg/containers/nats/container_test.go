package nats_test

import (
	"context"
	"testing"

	"github.com/hyphengolang/with-jetstream/pkg/containers/nats"
	gonats "github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNATSContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()

	c, err := nats.RunNATSContainer(ctx,
		nats.WithWaitStrategy(wait.ForLog("Listening for client connections on 0.0.0.0:4222")))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	url, err := c.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// connect to nats
	nc, err := gonats.Connect(url)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
	}

	// create a jetstream context
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("failed to create jetstream context: %v", err)
	}

	// add stream
	_, err = js.AddStream(&gonats.StreamConfig{
		Name:     "hello",
		Subjects: []string{"hello"},
	})
	if err != nil {
		t.Fatalf("failed to add stream: %v", err)
	}

	// subscribe and publish a hello message
	sub, err := js.SubscribeSync("hello", gonats.Durable("worker"))
	if err != nil {
		t.Fatalf("failed to subscribe to hello: %v", err)
	}

	// publish a message
	_, err = js.Publish("hello", []byte("world"))
	if err != nil {
		t.Fatalf("failed to publish hello: %v", err)
	}

	// wait for the message
	msg, err := sub.NextMsg(1000)
	if err != nil {
		t.Fatalf("failed to get message: %v", err)
	}

	if string(msg.Data) != "world" {
		t.Fatalf("expected world, got %s", msg.Data)
	}
}
