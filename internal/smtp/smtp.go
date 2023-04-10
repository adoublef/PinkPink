package smtp

type Recipient struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	// LastName is optional
	LastName string `json:"lastName"`
}

type Email struct {
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Recipient []Recipient `json:"recipient"`
}
