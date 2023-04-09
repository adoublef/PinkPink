package nats

import (
	"context"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

var (
	natsURL string
	natsJWT string
	natsNKey string
)

func init() {
	if natsURL = os.Getenv("NATS_URL"); natsURL == "" {
		panic("NATS_URL is not set")
	}
	if natsJWT = os.Getenv("NATS_USER_JWT"); natsJWT == "" {
		panic("NATS_USER_JWT is not set")
	}
	if natsNKey = os.Getenv("NATS_NKEY"); natsNKey == "" {
		panic("NATS_NKEY is not set")
	}
}

// NewConn returns a new nats connection. This is a utility function
func NewConn(ctx context.Context) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL, nats.UserJWTAndSeed(natsJWT, natsNKey))
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return nc, nil
}
