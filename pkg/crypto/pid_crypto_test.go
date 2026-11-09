package crypto

import (
	"testing"
)

func TestPIDCrypto_EncryptDecrypt(t *testing.T) {
	cipher, err := NewPIDCipher(DefaultDevKey, DefaultDevHmacSecret)
	if err != nil {
		t.Fatalf("Failed to initialize cipher: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid 13-digit Thai ID",
			input:    "1209700123456",
			expected: "1209700123456",
		},
		{
			name:     "Bangkok Citizen ID",
			input:    "1100112233441",
			expected: "1100112233441",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := cipher.Encrypt(tc.input)
			if err != nil {
				t.Fatalf("Encrypt error: %v", err)
			}
			if tc.input != "" && enc == tc.input {
				t.Errorf("Encrypted value should not equal plaintext")
			}

			dec := cipher.Decrypt(enc)
			if dec != tc.expected {
				t.Errorf("Decrypt mismatch: expected %q, got %q", tc.expected, dec)
			}
		})
	}
}

func TestPIDCrypto_LegacyAndEdgeCases(t *testing.T) {
	cipher, _ := NewPIDCipher(DefaultDevKey, DefaultDevHmacSecret)

	// 1. Legacy mock "ENC_..."
	legacy := "ENC_1100112233441"
	if got := cipher.Decrypt(legacy); got != "1100112233441" {
		t.Errorf("Legacy prefix failed: expected 1100112233441, got %s", got)
	}

	// 2. Already plaintext 13-digit
	plain := "1209700123456"
	if got := cipher.Decrypt(plain); got != "1209700123456" {
		t.Errorf("Plain ID failed: expected 1209700123456, got %s", got)
	}
}

func TestPIDCrypto_SafeMaskOnUnknown(t *testing.T) {
	cipher, _ := NewPIDCipher(DefaultDevKey, DefaultDevHmacSecret)
	unknown := "N3q87sDti01LJOswzig6IndrmnlKz9Fv/BjqY17idP3CPtF8bhBAK7U="
	masked := cipher.Decrypt(unknown)
	if masked == "" {
		t.Errorf("Expected masked fallback, got empty")
	}
}
