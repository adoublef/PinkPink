package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"net/url"
	"strings"
)

const (
	GoogleHost = "smtp.gmail.com"
)

// Sender is an interface for sending emails.
type Sender interface {
	// Send sends an email to the given recipients.
	Send(subject string, msg string, to ...string) error
}

// Client is a client for sending emails.
// All fields are exported in-case, you want to set them manually.
type Client struct {
	cfg *Config
}

type Config struct {
	// Username is client's email address Username
	Username string
	// Password is client's app Password for their email
	// currently supporting gmail.
	Password string
	// Hostname is the Hostname of the smtp server
	// currently supporting `smtp.gmail.com`.
	Hostname string
	// Port is the Port that the smtp server is listening on
	Port string
}

func ParseURL(s string) (*Config, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	var hostname string
	switch u.Scheme {
	case "google":
		hostname = GoogleHost
	default:
		return nil, fmt.Errorf("unknown scheme: %s", u.Scheme)
	}

	return &Config{
		Username: u.User.Username() + "@" + u.Hostname(),
		Password: password,
		Hostname: hostname,
		Port:     u.Port(),
	}, nil
}

// NewClient returns a new Client based on the connection string
// which should be in the format:
//
//	`<scheme>://<username>:<password>@<host>:<port>`
func NewClient(cfg *Config) Sender {
	return &Client{cfg}
}

func (c *Client) build(subject, msg string, to ...string) []byte {
	var buf bytes.Buffer
	// **** From: <email address> \r\n ****
	buf.WriteString(fmt.Sprintf("From: %s", c.cfg.Username))
	buf.WriteString("\r\n")
	// **** To: <email address> \r\n ****
	buf.WriteString(fmt.Sprintf("To: %s", strings.Join(to, ",")))
	buf.WriteString("\r\n")
	// **** Subject: <subject> \r\n ****
	buf.WriteString(fmt.Sprintf("Subject: %s", subject))
	buf.WriteString("\r\n")
	// **** header fields ****
	hdr := []string{"MIME-version: 1.0;", "Content-Type: text/html; charset=\"UTF-8\";"}
	buf.WriteString(strings.Join(hdr, "\r\n"))
	// **** \r\n\r\n ****
	buf.WriteString("\r\n")
	buf.WriteString("\r\n")
	// **** <message> ****
	buf.WriteString(msg)

	return buf.Bytes()
}

// MEthod for pinging the client, rather than sending an email.

func (c *Client) Send(subject, msg string, to ...string) error {
	s, err := smtp.Dial(c.cfg.Hostname + ":" + c.cfg.Port)
	if err != nil {
		return err
	}
	defer s.Close()

	tls := &tls.Config{InsecureSkipVerify: true, ServerName: c.cfg.Hostname}
	if err = s.StartTLS(tls); err != nil {
		return err
	}

	a := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Hostname)
	if err := s.Auth(a); err != nil {
		return err
	}

	if err := s.Mail(c.cfg.Username); err != nil {
		return err
	}
	
	for _, r := range to {
		if err := s.Rcpt(r); err != nil {
			return err
		}
	}

	w, err := s.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(c.build(subject, msg, to...)); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return s.Quit()
}
