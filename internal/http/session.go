package httpx

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/auth"
	"github.com/gin-gonic/gin"
)

func (s *Server) requireUser(c *gin.Context) {
	u, err := s.loadUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	c.Set("user", u)
	c.Next()
}

func (s *Server) loadUser(c *gin.Context) (auth.User, error) {
	access, _ := c.Cookie(auth.AccessCookie)
	if access != "" {
		u, err := s.auth.User(access)
		if err == nil {
			return u, nil
		}
	}
	refresh, err := c.Cookie(auth.RefreshCookie)
	if err != nil || refresh == "" {
		return auth.User{}, errNoSession
	}
	tok, err := s.auth.Refresh(refresh)
	if err != nil {
		s.cookies.ClearToken(c.Writer)
		return auth.User{}, err
	}
	s.cookies.SetToken(c.Writer, tok)
	return s.auth.User(tok.Access)
}

var errNoSession = errString("no session")

func currentUser(c *gin.Context) auth.User {
	v, ok := c.Get("user")
	if !ok {
		return auth.User{}
	}
	u, _ := v.(auth.User)
	return u
}

func (s *Server) ensureCSRF(c *gin.Context) {
	token, err := c.Cookie(auth.CSRFCookie)
	if err != nil || token == "" {
		token = randomToken()
		s.cookies.SetCSRF(c.Writer, token)
		c.Request.AddCookie(&http.Cookie{Name: auth.CSRFCookie, Value: token})
	}
	c.Set("csrf", token)
	c.Next()
}

func csrfFrom(c *gin.Context) string {
	v, _ := c.Get("csrf")
	s, _ := v.(string)
	return s
}

func (s *Server) requireCSRF(c *gin.Context) {
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
		c.Next()
		return
	}
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.Next()
		return
	}
	form := c.PostForm("csrf_token")
	cookie, _ := c.Cookie(auth.CSRFCookie)
	if form == "" || cookie == "" || subtle.ConstantTimeCompare([]byte(form), []byte(cookie)) != 1 {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}
