package operatorapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/starterapp/backend/internal/operatorauth"
)

// RequireOperator is the session security scheme as Gin middleware. A missing,
// malformed, or expired cookie is always 401: there is no lesser-privilege
// outcome to fall back to. The verified operator id is stashed in the Gin
// context; read it back with OperatorIDFrom.
func RequireOperator(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(operatorauth.SessionCookieName)
		if err != nil || cookie == "" {
			writeError(c, http.StatusUnauthorized, "unauthorized", "missing or invalid operator session")
			c.Abort()
			return
		}
		identity, err := operatorauth.VerifySession(secret, cookie, time.Now())
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized", "missing or invalid operator session")
			c.Abort()
			return
		}
		c.Set(operatorIDContextKey, identity.OperatorID)
		c.Next()
	}
}

// operatorIDContextKey is unexported so a handler reads the id only through
// OperatorIDFrom, never by string, and a typo can never become a silent
// empty-string identity.
const operatorIDContextKey = "starterapp_operator_id"

// OperatorIDFrom returns the operator id RequireOperator verified for this
// request. The boolean is false only if it is called from a route that is not
// behind RequireOperator, which is a wiring bug; handlers turn it into a 401
// rather than proceeding with an empty identity.
func OperatorIDFrom(c *gin.Context) (string, bool) {
	value, exists := c.Get(operatorIDContextKey)
	if !exists {
		return "", false
	}
	operatorID, ok := value.(string)
	if !ok || operatorID == "" {
		return "", false
	}
	return operatorID, true
}

// RequireJSONContentType is the CSRF mitigation for state-changing requests.
// It is not a synchronizer or double-submit token; it relies on a combination
// that makes the classic attack impossible:
//
//  1. The session cookie is SameSite=Strict, so a browser never attaches it to
//     any cross-site request.
//  2. As defence in depth, an HTML form (the only way a browser sends a
//     cookie-bearing cross-origin request without JavaScript) cannot set
//     Content-Type: application/json, and every mutating handler here only
//     accepts JSON.
//  3. This server enables no CORS policy, so cross-origin JavaScript cannot
//     attach credentials to a request against this API at all.
//
// If you add a CORS policy, or a non-browser client that posts non-JSON
// bodies, revisit this and add a real CSRF token.
func RequireJSONContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
			c.Next()
			return
		}
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			writeError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "request body must be application/json")
			c.Abort()
			return
		}
		c.Next()
	}
}
