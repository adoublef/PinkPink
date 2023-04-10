package smtp

import (
	"encoding/json"
	"net/mail"
)

type Recipient struct {
	// Address is required
	Address Address `json:"address"`
	// FirstName is required, must be between 1 to 50 characters
	FirstName string `json:"firstName"`
	// LastName is optional, must be between 1 to 50 characters if provided
	LastName string `json:"lastName"`
}

type Email struct {
	// Subject must be between 1 to 50 characters
	Subject string `json:"subject"`
	// Message must be between 1 to 255 characters
	Message    string      `json:"message"`
	Recipients []Recipient `json:"recipient"`
}

// Address wraps the mail.Address and adds custom encoding/decoding
// This ignores the Name field
type Address mail.Address

func (a *Address) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	addr, err := mail.ParseAddress(s)
	if err != nil {
		return err
	}

	*a = Address(*addr)

	return nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a Address) String() string { return a.Address }
