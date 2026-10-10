package operatorapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/starterapp/backend/internal/operatorauth"
)

// defaultTestDatabaseURL is the local Postgres from docker-compose.yml.
const defaultTestDatabaseURL = "postgres://starterapp:starterapp@localhost:5432/starterapp?sslmode=disable"

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultTestDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("skipping: could not create postgres pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping: postgres not reachable at %s: %v", url, err)
	}

	t.Cleanup(pool.Close)
	return pool
}

// testSessionSecret is a fixed, non-secret HMAC key for these tests only.
var testSessionSecret = []byte("test-only-fixed-32-byte-secret-k")

func testRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, pool, testSessionSecret)
	return r
}

// insertTestOperator creates a real operator via the real provisioning path
// (operatorauth.CreateOperator), never a hand-rolled row. email must be unique
// per test (the column is UNIQUE).
func insertTestOperator(t *testing.T, pool *pgxpool.Pool, email, password string) string {
	t.Helper()
	created, err := operatorauth.CreateOperator(t.Context(), pool, email, password)
	if err != nil {
		t.Fatalf("create test operator: %v", err)
	}
	t.Cleanup(func() {
		// A fresh context: t.Context() is already canceled when Cleanup runs.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, `DELETE FROM operators WHERE id = $1::uuid`, created.ID)
	})
	return created.ID
}

// realSessionCookie issues a valid session token directly (bypassing the login
// flow, which has its own tests).
func realSessionCookie(t *testing.T, operatorID string) string {
	t.Helper()
	token, _, err := operatorauth.IssueSession(testSessionSecret, operatorID, time.Now())
	if err != nil {
		t.Fatalf("issue test session: %v", err)
	}
	return token
}

// doRequest builds a request (optional session cookie and JSON body), runs it
// through the router, and returns the recorder.
func doRequest(t *testing.T, r *gin.Engine, method, path, sessionCookie string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: operatorauth.SessionCookieName, Value: sessionCookie})
	}
	// A real client sets this explicitly on every mutating request, body or
	// not; RequireJSONContentType depends on exactly that.
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	return recordRequest(r, req)
}

func mustRequest(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	return req
}

func recordRequest(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
