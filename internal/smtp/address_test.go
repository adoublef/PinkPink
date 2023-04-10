package smtp

import (
	"encoding/json"
	"testing"
)

// Test the Address type
func TestEncoding(t *testing.T) {
	// Test unmarshaling
	data := `"kristopherab@gmail.com"`
	
	var addr Address
	if err := json.Unmarshal([]byte(data), &addr); err != nil {
		t.Fatal(err)
	}

	if addr.String() != "kristopherab@gmail.com" {
		t.Fatalf("Expected %s, got %s", data, addr.String())
	}

	// Invalid address
	data = `"test#gmail.com"`

	if err := json.Unmarshal([]byte(data), &addr); err == nil {
		t.Fatal("Expected error")
	}
}