package routes

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "gothicforge3/internal/auth"
)

func init() { RegisterRoute(registerAuthAPI) }

func registerAuthAPI(r chi.Router) {
    r.Get("/api/me", http.HandlerFunc(apiMe))
    r.Get("/auth/logout", http.HandlerFunc(logout))
}

func apiMe(w http.ResponseWriter, r *http.Request) {
    claims, err := auth.ReadAndVerifyCookie(r, "gf_jwt")
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    _ = json.NewEncoder(w).Encode(claims)
}

func logout(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:     "gf_jwt",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Expires:  time.Unix(0, 0),
    })
    http.Redirect(w, r, "/", http.StatusFound)
}
