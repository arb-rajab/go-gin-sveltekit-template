// Command provision-operator creates the first operator account.
//
// There is no operator for /auth/login to authenticate until one exists, and
// every operator-facing endpoint requires a session, so nothing in the API
// could create the first account. This is a local, database-connected CLI that
// calls the real provisioning function (operatorauth.CreateOperator), never a
// hardcoded default email/password shipped in code or a migration.
//
// It refuses to run if an operator already exists. It does not offer a
// password-reset path; recovering a lost password today needs direct database
// access (UPDATE operators SET password_hash = ... computed with
// operatorauth.HashPassword). Delete the guard, or build a reset flow, if your
// product needs multiple accounts.
//
// Usage:
//
//	echo 'a-real-password' | provision-operator -email you@example.com
//
// The password is read from stdin, one line, deliberately not a -password flag,
// which would leave the plaintext in shell history and process listings.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/starterapp/backend/internal/operatorauth"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "provision-operator:", err)
		os.Exit(1)
	}
}

func run(args []string, in *os.File, out *os.File) error {
	fs := flag.NewFlagSet("provision-operator", flag.ContinueOnError)
	email := fs.String("email", "", "email address for the new operator account (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*email) == "" {
		return errors.New("-email is required")
	}

	password, err := readPassword(in)
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New("no password provided on stdin")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	existing, err := operatorauth.CountOperators(ctx, pool)
	if err != nil {
		return fmt.Errorf("check for existing operators: %w", err)
	}
	if existing > 0 {
		return fmt.Errorf("refusing to run: %d operator account(s) already exist; this tool only bootstraps the first operator (see the package doc for the lost-password gap)", existing)
	}

	created, err := operatorauth.CreateOperator(ctx, pool, *email, password)
	if err != nil {
		return fmt.Errorf("create operator: %w", err)
	}

	if _, err := fmt.Fprintf(out, "operator_id: %s\n", created.ID); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if _, err := fmt.Fprintf(out, "email:       %s\n", created.Email); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if _, err := fmt.Fprintln(out, "Operator account created. Log in at POST /api/v1/auth/login with this email and the password you provided."); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func readPassword(in *os.File) (string, error) {
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		return "", nil
	}
	return strings.TrimRight(scanner.Text(), "\r\n"), nil
}
