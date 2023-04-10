package http_test

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	smtpHTTP "github.com/adoublef/pinkpink/internal/smtp/http"
	smtpNATS "github.com/adoublef/pinkpink/internal/smtp/nats"
	containers "github.com/adoublef/pinkpink/pkg/containers/nats"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	natsContainer *containers.NATSContainer
	natsConn      *nats.Conn
)

func TestService(t *testing.T) {
	t.Setenv("DEBUG", "t")

	producer, err := smtpNATS.NewProducer(natsConn, nats.WorkQueuePolicy, 1024)
	require.NoError(t, err, "failed to create nats producer")

	// setup http service
	srv := newTestServer(t, producer)

	// setup nats consumer
	consumer, err := smtpNATS.NewWorker(natsConn, nats.AckExplicitPolicy, 1, "tester", smtpNATS.SubjectSubscribe)
	require.NoError(t, err, "failed to create nats consumer")

	t.Cleanup(func() { srv.Close() })

	// make a post request to the service at /subscribe
	body := `
	{
		"email":"kristopherab@gmail.com", 
		"subject":"Test",
		"message":"Hello World",
		"firstName":"Kristopher"
	}`

	resp, err := srv.Client().Post(srv.URL+"/api/subscribe", "application/json", strings.NewReader(body))
	require.NoError(t, err, "failed to make post request")

	require.Equal(t, 200, resp.StatusCode, "response status code does not match")

	email, err := consumer.NextMsg(context.Background())
	require.NoError(t, err, "failed to get next message")

	require.Equal(t, "Test", email.Subject, "email subject does not match")

	// email error
	body = `
	{
		"email":"test#gmail.com", 
		"subject":"Test",
		"message":"Hello World",
		"firstName":"Kristopher"
	}`

	resp, err = srv.Client().Post(srv.URL+"/api/subscribe", "application/json", strings.NewReader(body))
	require.NoError(t, err, "failed to make post request")

	require.Equal(t, 400, resp.StatusCode, "response status code does not match")
	
	// firstname error
	body = `
	{
		"email":"test#gmail.com", 
		"subject":"Test",
		"message":"Hello World"
	}`
	
	resp, err = srv.Client().Post(srv.URL+"/api/subscribe", "application/json", strings.NewReader(body))
	require.NoError(t, err, "failed to make post request")
	
	require.Equal(t, 400, resp.StatusCode, "response status code does not match")
}

func newTestServer(t *testing.T, p smtpNATS.Producer) *httptest.Server {
	t.Helper()

	return httptest.NewServer(smtpHTTP.New(p))
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
