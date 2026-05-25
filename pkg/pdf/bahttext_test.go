package pdf

import (
	"fmt"
	"testing"
)

func TestConvertToThaiBaht(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{350.00, "สามร้อยห้าสิบบาทถ้วน"},
		{2560775.46, "สองล้านห้าแสนหกหมื่นเจ็ดร้อยเจ็ดสิบห้าบาทสี่สิบหกสตางค์"},
		{4086.00, "สี่พันแปดสิบหกบาทถ้วน"},
		{0.05, "ห้าสตางค์"},
		{10.00, "สิบบาทถ้วน"},
		{11.00, "สิบเอ็ดบาทถ้วน"},
		{21.00, "ยี่สิบเอ็ดบาทถ้วน"},
		{1000000.00, "หนึ่งล้านบาทถ้วน"},
	}

	for _, test := range tests {
		result := ConvertToThaiBaht(test.amount)
		if result != test.expected {
			t.Errorf("ConvertToThaiBaht(%f) = %q, expected %q", test.amount, result, test.expected)
		} else {
			fmt.Printf("PASS: %f -> %s\n", test.amount, result)
		}
	}
}
