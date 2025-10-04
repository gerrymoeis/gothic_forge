package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/gologin"
	ghlogin "github.com/dghubble/gologin/github"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	ghoauth "golang.org/x/oauth2/github"

	"gothicforge3/internal/auth"
	"gothicforge3/internal/env"
)

func init() {
	RegisterRoute(registerOAuthGitHub)
}

func registerOAuthGitHub(r chi.Router) {
	clientID := strings.TrimSpace(env.Get("GITHUB_CLIENT_ID", ""))
	secret := strings.TrimSpace(env.Get("GITHUB_CLIENT_SECRET", ""))
	if clientID == "" || secret == "" {
		return // OAuth not configured
	}
	base := strings.TrimSpace(env.Get("OAUTH_BASE_URL", ""))
	if base == "" {
		base = strings.TrimSpace(env.Get("SITE_BASE_URL", ""))
	}
	if base == "" {
		base = "/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	cb := base + "auth/github/callback"

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  cb,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     ghoauth.Endpoint,
	}

	stateCfg := gologin.DebugOnlyCookieConfig

	r.Method(http.MethodGet, "/auth/github/login", ghlogin.StateHandler(stateCfg, ghlogin.LoginHandler(conf, nil)))
	r.Method(http.MethodGet, "/auth/github/callback", ghlogin.StateHandler(stateCfg, ghlogin.CallbackHandler(conf, http.HandlerFunc(githubSuccess), http.HandlerFunc(oauthFailure))))
	r.Get("/auth/logout", http.HandlerFunc(logout))
	r.Get("/api/me", http.HandlerFunc(apiMe))
}

func githubSuccess(w http.ResponseWriter, r *http.Request) {
	// Extract GitHub user from context
	u, err := ghlogin.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "user context error", http.StatusInternalServerError)
		return
	}
	var uid int64
	var login, name string
	if u != nil {
		uid = u.GetID()
		login = u.GetLogin()
		name = u.GetName()
	}
	claims := map[string]any{
		"sub":      uid,
		"login":    login,
		"name":     name,
		"provider": "github",
	}
	tok, exp, err := auth.Issue(7*24*time.Hour, claims)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	auth.SetJWTCookie(w, "gf_jwt", tok, exp)
	http.Redirect(w, r, "/", http.StatusFound)
}

func oauthFailure(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "oauth failure", http.StatusUnauthorized)
}

func logout(w http.ResponseWriter, r *http.Request) {
	// Expire cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "gf_jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/", http.StatusFound)
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
