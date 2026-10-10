package operatorapi

import "github.com/gin-gonic/gin"

// errorEnvelope is the one JSON error shape every endpoint returns:
// {"error": {"code", "message", "field"}}; field names the offending request
// field for validation errors and is null otherwise. Keep any other API package's error
// writer identical so clients handle one contract.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Field   *string `json:"field"`
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message, Field: nil}})
}

// writeFieldError is writeError for a validation failure tied to one request
// field or query parameter.
func writeFieldError(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, errorEnvelope{Error: errorBody{Code: code, Message: message, Field: &field}})
}
