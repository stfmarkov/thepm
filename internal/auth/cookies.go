package auth

import (
	"net/http"
	"time"
)

const (
	AccessCookie  = "thepm_access"
	RefreshCookie = "thepm_refresh"
	CSRFCookie    = "thepm_csrf"
)

type CookieConfig struct {
	Secure bool
}

func (cfg CookieConfig) SetToken(w http.ResponseWriter, t Token) {
	http.SetCookie(w, cfg.cookie(AccessCookie, t.Access, t.ExpiresIn))
	http.SetCookie(w, cfg.cookie(RefreshCookie, t.Refresh, int((30 * 24 * time.Hour).Seconds())))
}

func (cfg CookieConfig) ClearToken(w http.ResponseWriter) {
	http.SetCookie(w, cfg.cookie(AccessCookie, "", -1))
	http.SetCookie(w, cfg.cookie(RefreshCookie, "", -1))
}

func (cfg CookieConfig) SetCSRF(w http.ResponseWriter, token string) {
	http.SetCookie(w, cfg.cookie(CSRFCookie, token, int((12 * time.Hour).Seconds())))
}

func (cfg CookieConfig) cookie(name, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}
