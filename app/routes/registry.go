package routes

import "github.com/gofiber/fiber/v2"

// registrars holds functions that can mount additional routes onto the app.
var registrars []func(*fiber.App)

// RegisterRoute allows scaffolds and features to register route-mounting functions
// without editing the central routes.go.
func RegisterRoute(fn func(*fiber.App)) {
    registrars = append(registrars, fn)
}

// applyRegistrars mounts all registered route functions. Kept unexported to
// maintain a single public entry point (Register) for the app.
func applyRegistrars(app *fiber.App) {
    for _, fn := range registrars {
        fn(app)
    }
}
