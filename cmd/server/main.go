package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "gothicforge3/app/routes"
    "gothicforge3/internal/env"
    "gothicforge3/internal/server"
)

func main() {
	_ = env.Load()
	r := server.New()

	// Mount application routes
	routes.Register(r)

    // Prefer PORT from the platform (Railway/Heroku/etc.). Fallback to .env values.
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
	log.Printf("Gothic Forge v3 listening at http://%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
