// Package agentauth provides machine-credential primitives for non-human
// callers (agents, services, CI jobs): a bearer token generated once, stored
// only as a hash, and compared in constant time. It is a distinct identity
// space from the human session cookie in operatorauth; never accept one where
// the other is expected.
//
// This package has no database code. Add an `agents` (or `api_keys`) table,
// store HashToken(token) in it, and look the row up by hash on each request.
package agentauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// tokenBytes is the random credential's entropy: 256 bits, comfortably beyond
// what a bearer token needs to resist brute force.
const tokenBytes = 32

// GenerateToken mints a new, high-entropy bearer credential (crypto/rand,
// never math/rand; this is a real secret, not a display id). Returned
// base64url-encoded (RawURLEncoding: no padding, URL/header-safe) so it can
// be dropped directly into an Authorization: Bearer header.
func GenerateToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate credential: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken is the one-way transform to store and to recompute on every
// verification attempt.
//
// Plain SHA-256, not bcrypt/argon2: a slow password KDF earns its cost
// defending a human-chosen secret from a small keyspace, but GenerateToken's
// output is already 256 bits of uniform randomness, so a deliberately
// expensive hash would add cost to every request without shrinking an
// attacker's search space. A fast cryptographic hash is the standard choice
// for high-entropy API tokens.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// TokensEqual compares two hex-encoded hashes in constant time. Both sides are
// already-hashed values, so this is cheap hygiene for any secret-derived
// comparison rather than a defence against recovering the raw token.
func TokensEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
