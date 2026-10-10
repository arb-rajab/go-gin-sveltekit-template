package ssrf

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Literal IPs only, so the guard itself is proven without depending on
// network access or DNS.
func TestValidateURL_BlocksPrivateAndLoopbackLiterals(t *testing.T) {
	blocked := []string{
		"http://127.0.0.1/hook",
		"http://127.0.0.1:8080/hook",
		"https://169.254.169.254/latest/meta-data/",          // AWS/GCP metadata
		"http://169.254.170.2/v2/credentials/some-task-role", // ECS task metadata
		"http://10.0.0.5/hook",
		"http://172.16.0.5/hook",
		"http://192.168.1.5/hook",
		"http://0.0.0.0/hook",
		"http://[::1]/hook",
		"http://[fe80::1]/hook",
		"http://[fc00::1]/hook",
		"http://224.0.0.1/hook",
	}
	for _, dest := range blocked {
		t.Run(dest, func(t *testing.T) {
			if err := ValidateURL(t.Context(), dest); !errors.Is(err, ErrDestinationBlocked) {
				t.Fatalf("expected %q to be blocked, got %v", dest, err)
			}
		})
	}
}

// A syntactically valid public-looking destination, and one whose hostname
// does not resolve (reserved .invalid TLD), are accepted at configuration
// time; the dial-time guard enforces the real check.
func TestValidateURL_AllowsPublicLooking(t *testing.T) {
	allowed := []string{
		"https://hooks.example.invalid/T00/B00/a-fake-webhook-token",
		"https://hooks.slack.com/services/T00/B00/token",
		"http://8.8.8.8/hook",
	}
	for _, dest := range allowed {
		t.Run(dest, func(t *testing.T) {
			if err := ValidateURL(t.Context(), dest); err != nil {
				t.Fatalf("expected %q to be allowed, got %v", dest, err)
			}
		})
	}
}

func TestValidateURL_RejectsMalformedOrNonHTTP(t *testing.T) {
	cases := []string{
		"not a url at all: %zz",
		"file:///etc/passwd",
		"gopher://127.0.0.1:25/",
		"ftp://example.invalid/hook",
		"",
	}
	for _, dest := range cases {
		t.Run(dest, func(t *testing.T) {
			if err := ValidateURL(t.Context(), dest); !errors.Is(err, ErrDestinationBlocked) {
				t.Fatalf("expected %q to be rejected, got %v", dest, err)
			}
		})
	}
}

// The dial-time half: a request to a loopback server must be refused by the
// guarded client even though nothing validated the URL beforehand.
func TestGuardedClient_BlocksLoopbackAtDialTime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := NewGuardedClient(5 * time.Second).Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected the dial to a loopback address to be refused")
	}
	if !errors.Is(err, ErrDestinationBlocked) {
		t.Fatalf("expected ErrDestinationBlocked, got %v", err)
	}
}
