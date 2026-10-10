package operatorapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/starterapp/backend/internal/operatorauth"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	OperatorID string `json:"operator_id"`
	Email      string `json:"email"`
}

// LoginHandler is POST /auth/login. It is rate limited by both the attempted
// email and the caller's remote IP before any password verification happens,
// so a blocked caller never burns a bcrypt comparison.
func LoginHandler(pool *pgxpool.Pool, secret []byte, limiter *loginRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			writeError(c, http.StatusBadRequest, "malformed_request", "email and password are required")
			return
		}

		now := time.Now()
		ip := c.ClientIP()
		if !limiter.allow(req.Email, now) || !limiter.allow("ip:"+ip, now) {
			c.Header("Retry-After", loginRateWindow.String())
			writeError(c, http.StatusTooManyRequests, "too_many_requests", "too many failed login attempts, try again later")
			return
		}

		identity, err := operatorauth.Login(c.Request.Context(), pool, req.Email, req.Password)
		if err != nil {
			if errors.Is(err, operatorauth.ErrInvalidCredential) {
				limiter.recordFailure(req.Email, now)
				limiter.recordFailure("ip:"+ip, now)
				writeError(c, http.StatusUnauthorized, "unauthorized", "invalid email or password")
				return
			}
			writeError(c, http.StatusServiceUnavailable, "database_unavailable", "could not verify credentials")
			return
		}

		limiter.reset(req.Email)
		limiter.reset("ip:" + ip)

		token, expiresAt, err := operatorauth.IssueSession(secret, identity.OperatorID, now)
		if err != nil {
			writeError(c, http.StatusServiceUnavailable, "session_issuance_failed", "could not issue session")
			return
		}
		setSessionCookie(c, token, expiresAt)

		c.JSON(http.StatusOK, loginResponse{OperatorID: identity.OperatorID, Email: req.Email})
	}
}

// LogoutHandler is POST /auth/logout. It clears the cookie client-side, which
// is the full scope of "logout" a stateless token supports: an already-issued
// token stays valid until it expires or the signing secret is rotated.
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		clearSessionCookie(c)
		c.Status(http.StatusNoContent)
	}
}

// MeHandler is GET /auth/me: the signed-in operator's identity, used by the
// frontend to check whether a session cookie is still good.
func MeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		operatorID, ok := OperatorIDFrom(c)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "missing or invalid operator session")
			return
		}
		var email string
		err := pool.QueryRow(c.Request.Context(), `SELECT email FROM operators WHERE id = $1::uuid`, operatorID).Scan(&email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(c, http.StatusUnauthorized, "unauthorized", "missing or invalid operator session")
				return
			}
			writeError(c, http.StatusServiceUnavailable, "database_unavailable", "could not load operator")
			return
		}
		c.JSON(http.StatusOK, loginResponse{OperatorID: operatorID, Email: email})
	}
}

// Secure is always true, so browsers only send this cookie over HTTPS (or to
// localhost, which browsers treat as a secure context). Terminate TLS in front
// of the backend in any real deployment.
func setSessionCookie(c *gin.Context, token string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(operatorauth.SessionCookieName, token, int(time.Until(expiresAt).Seconds()), "/", "", true, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(operatorauth.SessionCookieName, "", -1, "/", "", true, true)
}
