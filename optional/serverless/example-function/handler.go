package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type greetRequest struct {
	Name string `json:"name"`
}

type greetResponse struct {
	Message string `json:"message"`
}

// handleRequest is the API Gateway (HTTP API, payload format 2.0) entry point:
// POST {"name": "..."} and get {"message": "Hello, ..."} back. A malformed
// body is a 422 response, not a Lambda invocation failure, so API Gateway
// returns it to the caller instead of retrying or logging a function error.
// Replace the body of this function with your own logic.
func handleRequest(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var req greetRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return jsonResponse(http.StatusUnprocessableEntity, map[string]string{
			"error": `request body is not valid JSON for {"name": "..."}`,
		})
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return jsonResponse(http.StatusUnprocessableEntity, map[string]string{
			"error": "name is required",
		})
	}

	return jsonResponse(http.StatusOK, greetResponse{Message: "Hello, " + name})
}

func jsonResponse(status int, body any) (events.APIGatewayV2HTTPResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}
