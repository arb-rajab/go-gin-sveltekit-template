// Command starterapp runs the starterapp backend API server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/starterapp/backend/internal/operatorapi"
	"github.com/example/starterapp/backend/internal/operatorauth"
	"github.com/example/starterapp/backend/internal/optional/livefeed"
)

const shutdownTimeout = 15 * time.Second

// setupRouter wires the health endpoints and the operator-facing REST surface.
// pool may be nil only for tests that exercise /health alone; a route
// registered against a nil pool fails when invoked, never at registration.
func setupRouter(pool *pgxpool.Pool, sessionSecret []byte, hub *livefeed.Hub) *gin.Engine {
	r := gin.Default()
	r.GET("/health", healthHandler)
	r.GET("/ready", readyHandler(pool))

	routes := operatorapi.RegisterRoutes(r, pool, sessionSecret)

	// Optional module: live events over Server-Sent Events. To remove it,
	// delete internal/optional/livefeed, the hub argument, and this call.
	livefeed.Register(routes.API, routes.Auth, operatorapi.OperatorIDFrom, hub)

	// Register your domain routes on routes.API here:
	//   routes.API.GET("/things", routes.Auth, ListThings(pool))
	return r
}

// healthHandler is liveness only: the process is up and can answer.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyHandler is readiness: it checks the dependency the service cannot work
// without (Postgres).
func readyHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if pool == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			slog.Error("readiness check failed: database unreachable", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	sessionSecret, err := operatorauth.SigningSecretFromEnv()
	if err != nil {
		return fmt.Errorf("operator session config: %w", err)
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           setupRouter(pool, sessionSecret, livefeed.NewHub()),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpErrCh := make(chan error, 1)
	go func() {
		if serveErr := srv.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			httpErrCh <- fmt.Errorf("http server: %w", serveErr)
			return
		}
		httpErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case serveErr := <-httpErrCh:
		if serveErr != nil {
			slog.Error("http server failed", "error", serveErr)
			return serveErr
		}
	}

	// A fresh bound for shutdown itself: the root ctx is already canceled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout) //nolint:contextcheck // intentional fresh bound, root ctx is already canceled
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown", "error", err)
	}
	return nil
}
