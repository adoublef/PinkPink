package smtp_test

import (
	"os"
	"testing"

	"github.com/adoublef/pinkpink/internal/smtp"
	"github.com/hyphengolang/prelude/testing/is"
)

var smtpUrl = os.Getenv("SMTP_URL")

func TestParseURL(t *testing.T) {
	is := is.New(t)

	_, err := smtp.ParseURL(smtpUrl)
	is.NoErr(err) // parse url
}

func TestNewClient(t *testing.T) {
	is := is.New(t)

	cfg, err := smtp.ParseURL(smtpUrl)
	is.NoErr(err) // parse url

	c := smtp.NewClient(cfg)

	to := "kristopherab@gmail.com"

	subject := "Dynamic HTML Email"
	msg := "<h1>Hello World, this is a new message!</h1>"

	// **** Send via TLS ****
	err = c.Send(subject, msg, to)
	is.NoErr(err) // send email
}
