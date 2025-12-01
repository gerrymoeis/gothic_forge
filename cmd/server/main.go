package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "gothicforge3/app/routes"
    "gothicforge3/internal/db"
    "gothicforge3/internal/env"
    "gothicforge3/internal/server"
)

func main() {
	_ = env.Load()
	r := server.New()

	// Mount application routes
	routes.Register(r)

    // Prefer PORT from the platform (Leapcell/Heroku/etc.). Fallback to .env values.
    port := os.Getenv("PORT")
    if port == "" { port = env.Get("HTTP_PORT", "8080") }
    host := env.Get("HTTP_HOST", "")
    if host == "" {
        if os.Getenv("PORT") != "" {
            host = "0.0.0.0"
        } else {
            host = "127.0.0.1"
        }
    }
    addr := fmt.Sprintf("%s:%s", host, port)

    // Create HTTP server with timeouts
    srv := &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server in goroutine
    go func() {
        log.Printf("Gothic Forge v3 listening at http://%s", addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    sig := <-quit

    log.Printf("Shutting down server (received %v)...", sig)
    shutdownStart := time.Now()

    // Graceful shutdown with 30s timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Close database connections
    db.Close()

    // Shutdown HTTP server
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }

    shutdownDuration := time.Since(shutdownStart)
    log.Printf("Server exited (shutdown took %v)", shutdownDuration)
}
