package protocol

import (
	"crypto/rand"
	"encoding/hex"
)

// NewTraceID generates a random W3C-compatible trace ID (32 lowercase hex chars).
//
// Panics on crypto/rand failure. This is an accepted deviation from the
// "no panic in library code" rule: crypto/rand failure indicates a catastrophic
// OS-level entropy issue where no cryptographic operation in the system can be
// trusted (AES-GCM encryption, key generation, token signing). Crashing
// immediately is safer than continuing with broken cryptography.
func NewTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("protocol.NewTraceID: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(buf[:])
}

// ValidTraceID returns true if s is a valid W3C trace ID (32 lowercase hex chars, non-zero).
func ValidTraceID(s string) bool {
	if len(s) != 32 {
		return false
	}
	allZero := true
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
		if c != '0' {
			allZero = false
		}
	}
	return !allZero
}
