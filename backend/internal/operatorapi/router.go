// Package operatorapi is the human-facing REST surface: session login/logout
// and the routes behind them. Every route except POST /auth/login is gated by
// RequireOperator.
package operatorapi

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Routes exposes the pieces other packages need to mount their own routes on
// the same authenticated group.
type Routes struct {
	// API is the /api/v1 group.
	API *gin.RouterGroup
	// Auth is the session middleware (RequireOperator).
	Auth gin.HandlerFunc
	// CSRF is the JSON-content-type check for state-changing routes.
	CSRF gin.HandlerFunc
}

// RegisterRoutes wires the auth surface onto r under /api/v1 and returns the
// group and middleware so callers can add domain routes:
//
//	routes := operatorapi.RegisterRoutes(r, pool, sessionSecret)
//	routes.API.GET("/things", routes.Auth, ListThings(pool))
//	routes.API.POST("/things", routes.Auth, routes.CSRF, CreateThing(pool))
//
// sessionSecret is operatorauth's HMAC signing key
// (operatorauth.SigningSecretFromEnv).
func RegisterRoutes(r *gin.Engine, pool *pgxpool.Pool, sessionSecret []byte) Routes {
	limiter := newLoginRateLimiter(loginFailureLimit, loginRateWindow)
	csrf := RequireJSONContentType()
	auth := RequireOperator(sessionSecret)

	api := r.Group("/api/v1")

	api.POST("/auth/login", csrf, LoginHandler(pool, sessionSecret, limiter))
	api.POST("/auth/logout", auth, LogoutHandler())
	api.GET("/auth/me", auth, MeHandler(pool))

	return Routes{API: api, Auth: auth, CSRF: csrf}
}
