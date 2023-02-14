package services

import (
	"fmt"
	"testing"
)

func TestNewGinlongProvider(t *testing.T) {
	provider := NewGinlongProvider("Pierre", "p.jongejan93@upcmail.nl", "Cct3ugwk", "172533", 10, nil)
	if provider == nil {
		t.Error("Expected provider")
	}
	status, err := provider.GetSolarStatus()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	fmt.Printf("%+v", status)
}
