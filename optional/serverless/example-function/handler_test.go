package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

// handleRequest is invoked exactly as the Lambda runtime would: an
// events.APIGatewayV2HTTPRequest in, an events.APIGatewayV2HTTPResponse out.
func TestHandleRequest_Greets(t *testing.T) {
	resp, err := handleRequest(events.APIGatewayV2HTTPRequest{Body: `{"name":"Ada"}`})
	if err != nil {
		t.Fatalf("handleRequest returned an error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	var got greetResponse
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		t.Fatalf("decode response body: %v (body=%s)", err, resp.Body)
	}
	if got.Message != "Hello, Ada" {
		t.Fatalf("unexpected message %q", got.Message)
	}
}

func TestHandleRequest_MalformedBody(t *testing.T) {
	resp, err := handleRequest(events.APIGatewayV2HTTPRequest{Body: "not json"})
	if err != nil {
		t.Fatalf("handleRequest returned an error: %v", err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHandleRequest_MissingName(t *testing.T) {
	resp, err := handleRequest(events.APIGatewayV2HTTPRequest{Body: `{"name":"  "}`})
	if err != nil {
		t.Fatalf("handleRequest returned an error: %v", err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}
