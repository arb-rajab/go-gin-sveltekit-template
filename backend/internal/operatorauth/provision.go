package operatorauth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// provisionDB is the minimal surface CreateOperator/CountOperators need,
// satisfied by *pgxpool.Pool.
type provisionDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Created is the result of provisioning a new operator account. The operator
// authenticates with the password they chose, not a generated credential.
type Created struct {
	ID    string
	Email string
}

// CountOperators reports how many operator rows exist, so the bootstrapping
// tool (cmd/provision-operator) can refuse to create a second account unless
// told to.
func CountOperators(ctx context.Context, db provisionDB) (int, error) {
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM operators`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count operators: %w", err)
	}
	return count, nil
}

// CreateOperator hashes the chosen password and inserts the row. It is the
// answer to the first-account chicken-and-egg problem: a real CLI
// (cmd/provision-operator) calling a real function, never a hardcoded default
// credential shipped in code or a migration.
func CreateOperator(ctx context.Context, db provisionDB, email, password string) (Created, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return Created{}, fmt.Errorf("hash operator password: %w", err)
	}

	const stmt = `INSERT INTO operators (email, password_hash) VALUES ($1, $2) RETURNING id::text`
	var id string
	if err := db.QueryRow(ctx, stmt, email, hash).Scan(&id); err != nil {
		return Created{}, fmt.Errorf("insert operator %q: %w", email, err)
	}
	return Created{ID: id, Email: email}, nil
}
