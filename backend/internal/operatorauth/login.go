package operatorauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInvalidCredential is returned by Login for any presented email/password
// pair that does not verify: unknown email or wrong password. Deliberately one
// error for both, since telling a caller "that email doesn't exist" versus
// "wrong password" hands an attacker an account-enumeration oracle.
var ErrInvalidCredential = errors.New("invalid email or password")

// dummyHash is compared against on an unknown-email login attempt so that
// Login takes roughly the same time (one bcrypt comparison) whether or not the
// email exists; otherwise response latency would reveal which emails are
// registered. It is a fixed, non-secret bcrypt hash of an unguessable
// placeholder: it protects no real credential, it only burns the same CPU.
const dummyHash = "$2a$10$C6UzMDM.H6dfI/f/IKcEeOgOo8ZzP6NgTBH5AtNBjqM2i.vLxJ2Eq"

// Login verifies an email/password against operators.password_hash and returns
// the operator's identity. The caller (operatorapi.LoginHandler) issues a
// session token from the returned Identity.
func Login(ctx context.Context, pool *pgxpool.Pool, email, password string) (Identity, error) {
	const stmt = `SELECT id::text, password_hash FROM operators WHERE email = $1`
	var id, hash string
	err := pool.QueryRow(ctx, stmt, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			VerifyPassword(dummyHash, password) // burn the same bcrypt cost as a real lookup
			return Identity{}, ErrInvalidCredential
		}
		return Identity{}, fmt.Errorf("look up operator by email: %w", err)
	}

	if !VerifyPassword(hash, password) {
		return Identity{}, ErrInvalidCredential
	}
	return Identity{OperatorID: id}, nil
}
