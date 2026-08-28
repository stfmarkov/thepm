package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const (
	ingestIPMax      = 30
	ingestProjectMax = 60
	ingestWindowSec  = 15 * 60
	ingestMaxBody    = 32 << 10
	ingestMaxName    = 200
	ingestMaxEmail   = 320
	ingestMaxMessage = 8000
)

type ingestRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Rating  *int   `json:"rating"`
}

func (s *Server) ingestCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, X-Feedback-Key")
	c.Header("Access-Control-Max-Age", "3600")
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusNoContent)
		c.Abort()
		return
	}
	c.Next()
}

func (s *Server) limitIngest(c *gin.Context) {
	if c.Request.Method == http.MethodOptions {
		c.Next()
		return
	}
	if !s.limit.allow("ingest-ip:"+c.ClientIP(), ingestIPMax, ingestWindowSec) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"ok": false, "error": "too many requests"})
		return
	}
	pid := c.Param("project_id")
	if pid != "" && !s.limit.allow("ingest-proj:"+pid, ingestProjectMax, ingestWindowSec) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"ok": false, "error": "too many requests"})
		return
	}
	c.Next()
}

func (s *Server) ingestFeedback(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ingestMaxBody)

	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "not found"})
		return
	}
	key := strings.TrimSpace(c.GetHeader("X-Feedback-Key"))
	if key == "" {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "not found"})
		return
	}

	var req ingestRequest
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	message := strings.TrimSpace(req.Message)
	if name == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "name and message are required"})
		return
	}
	if len(name) > ingestMaxName || len(email) > ingestMaxEmail || len(message) > ingestMaxMessage {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "field too long"})
		return
	}
	if email != "" && !strings.Contains(email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid email"})
		return
	}
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "rating must be 1 to 5"})
		return
	}

	p, err := s.q.GetProjectByIngest(c.Request.Context(), db.GetProjectByIngestParams{
		ID:                projectID,
		FeedbackIngestKey: key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "not found"})
			return
		}
		log.Printf("ingest lookup: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "could not save"})
		return
	}

	if origin := strings.TrimSpace(p.FeedbackOrigin); origin != "" {
		reqOrigin := strings.TrimSpace(c.GetHeader("Origin"))
		if reqOrigin != "" && !originsEqual(origin, reqOrigin) {
			c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": "origin not allowed"})
			return
		}
	}

	rating := optionalRating("")
	if req.Rating != nil {
		rating.Int32 = int32(*req.Rating)
		rating.Valid = true
	}

	_, err = s.q.CreateFeedback(c.Request.Context(), db.CreateFeedbackParams{
		UserID:      p.UserID,
		ProjectID:   p.ID,
		AuthorName:  name,
		AuthorEmail: optionalText(email),
		Message:     message,
		Rating:      rating,
		Source:      "ingest",
	})
	if err != nil {
		log.Printf("ingest insert: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "could not save"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) updateFeedbackOrigin(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	origin, err := normalizeOrigin(c.PostForm("origin"))
	if err != nil {
		s.renderProjectFeedback(c, p, http.StatusBadRequest, "Website must be an http(s) URL, for example https://myapp.com")
		return
	}
	updated, err := s.q.UpdateFeedbackOrigin(c.Request.Context(), db.UpdateFeedbackOriginParams{
		ID:             p.ID,
		UserID:         p.UserID,
		FeedbackOrigin: origin,
	})
	if err != nil {
		log.Printf("update feedback origin: %v", err)
		s.renderProjectFeedback(c, p, http.StatusInternalServerError, "Could not save the website.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectFeedback(c, updated, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(updated.ID))
}

func (s *Server) rotateFeedbackIngestKey(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	updated, err := s.q.RotateFeedbackIngestKey(c.Request.Context(), db.RotateFeedbackIngestKeyParams{
		ID:     p.ID,
		UserID: p.UserID,
	})
	if err != nil {
		log.Printf("rotate ingest key: %v", err)
		s.renderProjectFeedback(c, p, http.StatusInternalServerError, "Could not rotate the key.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectFeedback(c, updated, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(updated.ID))
}

func publicIngestURL(c *gin.Context, projectID string) string {
	scheme := "http"
	if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/api/v1/projects/" + projectID + "/feedback"
}

func toDetailView(c *gin.Context, p db.Project) views.Project {
	v := toView(p)
	v.IngestURL = publicIngestURL(c, v.ID)
	return v
}

func normalizeOrigin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	return parseOrigin(raw)
}

func originsEqual(a, b string) bool {
	oa, err1 := parseOrigin(a)
	ob, err2 := parseOrigin(b)
	return err1 == nil && err2 == nil && oa == ob
}

func parseOrigin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("invalid origin")
	}
	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}
