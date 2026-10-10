package operatorapi

import (
	"net/http"
	"testing"

	"github.com/example/starterapp/backend/internal/agentauth"
)

// Every route behind RequireOperator is listed here and proven to reject a
// request with no session cookie. Add a row whenever you register a gated
// route; a route missing from this table is a route nobody checked.
func TestEveryGatedEndpoint_RejectsUnauthenticatedRequests(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	cases := []struct {
		method string
		path   string
		body   []byte
	}{
		{http.MethodPost, "/api/v1/auth/logout", nil},
		{http.MethodGet, "/api/v1/auth/me", nil},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := doRequest(t, r, tc.method, tc.path, "", tc.body)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 with no session cookie, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// A syntactically present but bogus cookie must be rejected too, not just an
// absent one.
func TestGatedEndpoint_RejectsInvalidSessionCookie(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	const bogus = "this-was-never-issued.by-anything"
	w := doRequest(t, r, http.MethodGet, "/api/v1/auth/me", bogus, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a bogus session cookie, got %d: %s", w.Code, w.Body.String())
	}
}

// A machine bearer token presented as a session cookie value must fail the
// HMAC check: the two credential shapes are disjoint identity spaces, so this
// is a 401 like any other missing credential, never a 403.
func TestGatedEndpoint_RejectsAMachineToken(t *testing.T) {
	pool := testPool(t)
	r := testRouter(pool)

	token, err := agentauth.GenerateToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	w := doRequest(t, r, http.MethodGet, "/api/v1/auth/me", token, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
