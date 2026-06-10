package qr

import (
	"testing"
)

func TestCalculateCRC16(t *testing.T) {

	input := "123456789"
	expected := "29B1"
	result := CalculateCRC16(input)
	if result != expected {
		t.Errorf("CalculateCRC16(%q) = %q; want %q", input, result, expected)
	}
}

func TestGeneratePromptPayBillPayment(t *testing.T) {
	billerID := "099400016485800"
	ref1 := "123456701"
	ref2 := "20260500"
	amount := 750.00

	qrContent, err := GeneratePromptPayBillPayment(billerID, ref1, ref2, amount)
	if err != nil {
		t.Fatalf("Unexpected error generating QR: %v", err)
	}

	if !testing.Short() {
		t.Logf("Generated QR Content: %s", qrContent)
	}

	if len(qrContent) < 20 {
		t.Errorf("Generated QR content too short: %s", qrContent)
	}
}
