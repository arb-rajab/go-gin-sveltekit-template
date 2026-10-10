// Package operatorauth implements a stateless session mechanism: a
// server-HMAC-signed cookie (operator ID + expiry) with no server-side session
// store. Revocation is by signing-secret rotation, not per session; a short
// SessionTTL bounds the exposure of an already-issued token. If you need
// per-session revocation, add a session table and check it in VerifySession.
package operatorauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SessionCookieName is the name of the session cookie. The SvelteKit frontend
// (frontend/src/lib/server/backend.ts) forwards this cookie by name.
const SessionCookieName = "starterapp_session"

// SessionTTL bounds how long an issued session is valid. It is the one
// mitigation for the stateless design's lack of per-session revocation: a
// shorter TTL shrinks the window an already-captured token stays usable.
const SessionTTL = 24 * time.Hour

// minSigningSecretBytes is a floor on the HMAC key length: a short key weakens
// HMAC-SHA256's security margin.
const minSigningSecretBytes = 32

// ErrInvalidSession is returned by VerifySession for any token that does not
// verify: malformed, badly signed, or expired. Deliberately one error for all
// three, since an unauthenticated caller has no legitimate use for knowing
// which.
var ErrInvalidSession = errors.New("invalid or expired operator session")

// Identity is what a verified session resolves to.
type Identity struct {
	OperatorID string
}

// SigningSecretFromEnv loads the HMAC key used to sign and verify session
// tokens, from SESSION_SIGNING_SECRET (base64-encoded, must decode to at
// least 32 bytes). Required, with no hardcoded fallback: an unset or too-short
// secret is a startup-time error, not a silently-weak default. Generate one
// with `openssl rand -base64 32`.
func SigningSecretFromEnv() ([]byte, error) {
	encoded := os.Getenv("SESSION_SIGNING_SECRET")
	if encoded == "" {
		return nil, errors.New("SESSION_SIGNING_SECRET not set")
	}
	secret, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode SESSION_SIGNING_SECRET: %w", err)
	}
	if len(secret) < minSigningSecretBytes {
		return nil, fmt.Errorf("SESSION_SIGNING_SECRET must decode to at least %d bytes, got %d", minSigningSecretBytes, len(secret))
	}
	return secret, nil
}

// IssueSession mints a new session token for operatorID, valid for
// SessionTTL from now. The payload (operator ID + expiry) is signed, never
// encrypted — nothing in it is secret, only its integrity matters (the same
// property a JWT's own signature protects), so HMAC-SHA256 over a plain
// payload is the complete mechanism, not a partial one.
func IssueSession(secret []byte, operatorID string, now time.Time) (token string, expiresAt time.Time, err error) {
	expiresAt = now.Add(SessionTTL)
	payload := encodePayload(operatorID, expiresAt)
	sig := sign(secret, payload)
	token = base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig)
	return token, expiresAt, nil
}

// VerifySession checks a presented token's signature and expiry, returning
// the operator identity it names. now is passed in explicitly so expiry is
// deterministically testable.
func VerifySession(secret []byte, token string, now time.Time) (Identity, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return Identity{}, ErrInvalidSession
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Identity{}, ErrInvalidSession
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, ErrInvalidSession
	}
	if !hmac.Equal(sig, sign(secret, payload)) {
		return Identity{}, ErrInvalidSession
	}

	operatorID, expiresAt, err := decodePayload(payload)
	if err != nil {
		return Identity{}, ErrInvalidSession
	}
	if now.After(expiresAt) {
		return Identity{}, ErrInvalidSession
	}
	return Identity{OperatorID: operatorID}, nil
}

func sign(secret, payload []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return mac.Sum(nil)
}

// encodePayload/decodePayload use a delimiter (":") that can never appear in
// either field: operatorID is always a UUID (hyphens and hex digits only)
// and expiresAt is always a decimal Unix-seconds integer.
func encodePayload(operatorID string, expiresAt time.Time) []byte {
	return []byte(operatorID + ":" + strconv.FormatInt(expiresAt.Unix(), 10))
}

func decodePayload(payload []byte) (operatorID string, expiresAt time.Time, err error) {
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return "", time.Time{}, errors.New("malformed session payload")
	}
	unixSeconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("malformed session expiry: %w", err)
	}
	return parts[0], time.Unix(unixSeconds, 0), nil
}
