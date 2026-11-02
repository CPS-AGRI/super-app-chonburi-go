package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
)

// DefaultDevKey is used only in development if PID_ENCRYPTION_KEY is unset.
const DefaultDevKey = "7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a"
const DefaultDevHmacSecret = "chonburi_pid_hmac_default_secret_key_2026"

// PIDCipher provides high-performance, thread-safe AES-256-GCM encryption/decryption
// and HMAC-SHA256 blind indexing for Thai Citizen IDs and sensitive identity fields.
type PIDCipher struct {
	mu         sync.RWMutex
	gcm        cipher.AEAD
	rawKey     []byte
	hmacSecret []byte
	isReady    bool
}

var (
	defaultCipher *PIDCipher
	onceDefault   sync.Once
)

// InitPIDCipher initializes the global PIDCipher instance with the given 64-character hex key and HMAC secret.
func InitPIDCipher(keyHex, hmacSecret string) error {
	c, err := NewPIDCipher(keyHex, hmacSecret)
	if err != nil {
		return err
	}
	defaultCipher = c
	return nil
}

// GetPIDCipher returns the singleton PIDCipher instance. If not explicitly initialized,
// it lazily initializes with safe defaults.
func GetPIDCipher() *PIDCipher {
	onceDefault.Do(func() {
		if defaultCipher == nil {
			c, _ := NewPIDCipher(DefaultDevKey, DefaultDevHmacSecret)
			defaultCipher = c
		}
	})
	return defaultCipher
}

// NewPIDCipher creates a new PIDCipher instance.
// keyHex must be a 64-char hex string (32 bytes = 256 bits).
func NewPIDCipher(keyHex, hmacSecret string) (*PIDCipher, error) {
	cleanKey := strings.TrimSpace(keyHex)
	if cleanKey == "" {
		cleanKey = DefaultDevKey
	}

	keyBytes, err := hex.DecodeString(cleanKey)
	if err != nil || len(keyBytes) != 32 {
		return nil, fmt.Errorf("PID_ENCRYPTION_KEY must be a valid 64-char hex string (32 bytes)")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AES-GCM AEAD: %w", err)
	}

	sec := strings.TrimSpace(hmacSecret)
	if sec == "" {
		sec = DefaultDevHmacSecret
	}

	return &PIDCipher{
		gcm:        gcm,
		rawKey:     keyBytes,
		hmacSecret: []byte(sec),
		isReady:    true,
	}, nil
}

// Encrypt encrypts plaintext into Base64(nonce + ciphertext + tag) using AES-256-GCM.
func (c *PIDCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	c.mu.RLock()
	gcm := c.gcm
	c.mu.RUnlock()

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	// Seal appends ciphertext and authentication tag directly to nonce
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt intelligently decrypts a stored identity string.
// Handles:
// 1. AES-256-GCM Base64 ciphertext (produced by Mobile App / Encrypt)
// 2. Legacy seed/mock prefix ("ENC_1100112233441" -> "1100112233441")
// 3. Raw plaintext 13-digit numbers (left untouched)
// 4. Invalid or corrupted ciphertext (safely falls back to masked representation without panic)
func (c *PIDCipher) Decrypt(storedVal string) string {
	trimmed := strings.TrimSpace(storedVal)
	if trimmed == "" || trimmed == "-" {
		return trimmed
	}

	// 1. Handle Legacy mock prefix "ENC_"
	if strings.HasPrefix(trimmed, "ENC_") {
		return trimmed[4:]
	}

	// 2. If it's already a plain 13-digit Thai National ID, return directly
	if isPlainThaiID(trimmed) {
		return trimmed
	}

	// 3. Attempt AES-256-GCM Base64 decode and decryption
	c.mu.RLock()
	gcm := c.gcm
	c.mu.RUnlock()

	rawCipher, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil && len(rawCipher) >= gcm.NonceSize()+16 {
		nonceSize := gcm.NonceSize()
		nonce := rawCipher[:nonceSize]
		ciphertextWithTag := rawCipher[nonceSize:]

		plaintextBytes, decErr := gcm.Open(nil, nonce, ciphertextWithTag, nil)
		if decErr == nil {
			return string(plaintextBytes)
		}
	}

	// 4. If decryption fails or not recognized, return as safe masked string
	return MaskPID(trimmed)
}

// Hash generates an HMAC-SHA256 hex string for blind-indexing and indexed database searches.
func (c *PIDCipher) Hash(plaintext string) string {
	clean := strings.TrimSpace(plaintext)
	if clean == "" {
		return ""
	}
	h := hmac.New(sha256.New, c.hmacSecret)
	h.Write([]byte(clean))
	return hex.EncodeToString(h.Sum(nil))
}

// Global convenience wrappers using defaultCipher:

// DecryptPID decrypts a stored PID using the singleton cipher.
func DecryptPID(val string) string {
	return GetPIDCipher().Decrypt(val)
}

// EncryptPID encrypts a PID using the singleton cipher.
func EncryptPID(plaintext string) (string, error) {
	return GetPIDCipher().Encrypt(plaintext)
}

// HashPID hashes a PID using HMAC-SHA256.
func HashPID(plaintext string) string {
	return GetPIDCipher().Hash(plaintext)
}

// MaskPID masks a Citizen ID for safe display in UI/logs, e.g. "1-1001-•••••-44-1".
func MaskPID(id string) string {
	digits := extractDigits(id)
	if len(digits) == 13 {
		// Format: 1-1001-•••••-44-1
		return fmt.Sprintf("%s-%s-•••••-%s-%s", digits[0:1], digits[1:5], digits[10:12], digits[12:13])
	}
	if len(id) > 6 {
		return id[:3] + "••••••" + id[len(id)-3:]
	}
	return "•••••••••••••"
}

// FormatThaiID formats a 13-digit ID into standard "x-xxxx-xxxxx-xx-x".
func FormatThaiID(id string) string {
	digits := extractDigits(id)
	if len(digits) == 13 {
		return fmt.Sprintf("%s-%s-%s-%s-%s", digits[0:1], digits[1:5], digits[5:10], digits[10:12], digits[12:13])
	}
	return id
}

func isPlainThaiID(s string) bool {
	if len(s) != 13 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ExtractDigits returns only numeric digits from a string.
func ExtractDigits(s string) string {
	return extractDigits(s)
}

func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

