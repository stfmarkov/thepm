package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/auth"
	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/AdventurousNerd/thepm/static"
	"github.com/gin-gonic/gin"
)

type Server struct {
	auth    *auth.Client
	cookies auth.CookieConfig
	limit   *ipLimiter
	q       *db.Queries
}

func New() (*gin.Engine, error) {
	baseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	anonKey := strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	if baseURL == "" || anonKey == "" {
		return nil, errMissingAuthEnv
	}

	pool, err := db.Open(context.Background())
	if err != nil {
		return nil, err
	}

	s := &Server{
		auth:    auth.NewClient(baseURL, anonKey),
		cookies: auth.CookieConfig{Secure: os.Getenv("COOKIE_SECURE") == "true"},
		limit:   newIPLimiter(),
		q:       db.New(pool),
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	r.StaticFS("/static", http.FS(static.FS))
	r.GET("/health", health)

	public := r.Group("/")
	public.Use(s.ensureCSRF, s.requireCSRF)
	public.GET("/login", s.showLogin)
	public.POST("/login", s.limitAuth, s.postLogin)
	public.GET("/register", s.showRegister)
	public.POST("/register", s.limitAuth, s.postRegister)
	public.POST("/logout", s.postLogout)

	authed := r.Group("/")
	authed.Use(s.ensureCSRF, s.requireCSRF, s.requireUser)
	authed.GET("/", s.listProjects)
	authed.GET("/projects/new", s.newProject)
	authed.POST("/projects", s.createProject)
	authed.GET("/projects/:id", s.showProject)
	authed.GET("/projects/:id/edit", s.editProject)
	authed.POST("/projects/:id", s.updateProject)
	authed.POST("/projects/:id/delete", s.deleteProject)
	authed.POST("/projects/:id/links", s.createLink)
	authed.POST("/projects/:id/links/:link_id", s.updateLink)
	authed.POST("/projects/:id/links/:link_id/delete", s.deleteLink)

	return r, nil
}

var errMissingAuthEnv = errString("SUPABASE_URL and SUPABASE_ANON_KEY are required")

type errString string

func (e errString) Error() string { return string(e) }

func health(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("ok\n"))
}

func (s *Server) showLogin(c *gin.Context) {
	if _, err := c.Cookie(auth.AccessCookie); err == nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	Render(c, http.StatusOK, views.Login(csrfFrom(c), ""))
}

func (s *Server) postLogin(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	if email == "" || password == "" {
		Render(c, http.StatusBadRequest, views.Login(csrfFrom(c), "Could not sign in."))
		return
	}

	tok, err := s.auth.SignIn(email, password)
	if err != nil {
		log.Printf("login failed: %v", err)
		Render(c, http.StatusUnauthorized, views.Login(csrfFrom(c), "Could not sign in."))
		return
	}
	s.cookies.SetToken(c.Writer, tok)
	c.Redirect(http.StatusSeeOther, "/")
}

func (s *Server) showRegister(c *gin.Context) {
	if _, err := c.Cookie(auth.AccessCookie); err == nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	Render(c, http.StatusOK, views.Register(csrfFrom(c), ""))
}

func (s *Server) postRegister(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirm := c.PostForm("password_confirm")
	if email == "" || password == "" {
		Render(c, http.StatusBadRequest, views.Register(csrfFrom(c), "Email and password are required."))
		return
	}
	if password != confirm {
		Render(c, http.StatusBadRequest, views.Register(csrfFrom(c), "Passwords do not match."))
		return
	}
	if len(password) < 6 {
		Render(c, http.StatusBadRequest, views.Register(csrfFrom(c), "Password must be at least 6 characters."))
		return
	}

	tok, err := s.auth.SignUp(email, password)
	if err != nil {
		msg := "Could not create the account."
		if ae, ok := err.(*auth.Error); ok && ae.Status == http.StatusUnprocessableEntity {
			msg = "That email is already registered."
		}
		log.Printf("register failed: %v", err)
		Render(c, http.StatusBadRequest, views.Register(csrfFrom(c), msg))
		return
	}
	s.cookies.SetToken(c.Writer, tok)
	c.Redirect(http.StatusSeeOther, "/")
}

func (s *Server) postLogout(c *gin.Context) {
	if access, err := c.Cookie(auth.AccessCookie); err == nil && access != "" {
		_ = s.auth.Logout(access)
	}
	s.cookies.ClearToken(c.Writer)
	c.Redirect(http.StatusSeeOther, "/login")
}

func (s *Server) limitAuth(c *gin.Context) {
	if !s.limit.allow(c.ClientIP(), 20, 15*60) {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}
	c.Next()
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
