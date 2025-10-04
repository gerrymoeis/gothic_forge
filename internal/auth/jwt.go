package auth

import (
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"gothicforge3/internal/env"
)

var tokenAuth *jwtauth.JWTAuth

// Init initializes the JWT signer using JWT_SECRET.
func Init() {
	secret := env.Get("JWT_SECRET", "devsecret-change-me")
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
}

// TokenAuth returns the configured JWT auth instance.
func TokenAuth() *jwtauth.JWTAuth {
	if tokenAuth == nil {
		Init()
	}
	return tokenAuth
}

// Issue creates a signed JWT string with standard claims and provided custom claims.
func Issue(ttl time.Duration, claims map[string]any) (string, time.Time, error) {
	ta := TokenAuth()
	exp := time.Now().Add(ttl)
	std := map[string]any{
		"exp": exp.Unix(),
		"iat": time.Now().Unix(),
	}
	for k, v := range claims {
		std[k] = v
	}
	_, t, err := ta.Encode(std)
	return t, exp, err
}

// SetJWTCookie writes a JWT as an HttpOnly cookie, SameSite=Lax, Secure in prod.
func SetJWTCookie(w http.ResponseWriter, name, token string, exp time.Time) {
	c := &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  exp,
	}
	if env.Get("APP_ENV", "development") == "production" {
		c.Secure = true
	}
	http.SetCookie(w, c)
}

// ReadAndVerifyCookie reads a cookie and verifies the JWT.
func ReadAndVerifyCookie(r *http.Request, name string) (map[string]any, error) {
	ck, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	ta := TokenAuth()
	tok, err := jwtauth.VerifyToken(ta, ck.Value)
	if err != nil {
		return nil, err
	}
	claims, err := tok.AsMap(r.Context())
	if err != nil {
		return nil, err
	}
	return claims, nil
}
