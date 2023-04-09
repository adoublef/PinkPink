package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NewConn returns a new nats connection. This is a u
func NewConn(ctx context.Context, natsURL, natsJWT, natsNKey string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL, nats.UserJWTAndSeed(natsJWT, natsNKey))
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return nc, nil
}
