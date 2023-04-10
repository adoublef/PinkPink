package nats

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultNATSImage = "nats:latest"
)

type NATSContainer struct {
	testcontainers.Container
}

func (c *NATSContainer) ConnectionString(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, "4222/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := "nats://" + host + ":" + port.Port()
	return connStr, nil
}

type NATSContainerOption func(req *testcontainers.ContainerRequest)

func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithImage(image string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if image == "" {
			image = defaultNATSImage
		}

		req.Image = image
	}
}

func RunNATSContainer(ctx context.Context, opts ...NATSContainerOption) (*NATSContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultNATSImage,
		ExposedPorts: []string{"4222/tcp", "6222/tcp", "8222/tcp"},
		Cmd:          []string{"-DV", "-js"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &NATSContainer{Container: c}, nil
}
