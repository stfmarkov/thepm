package httpx

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) createLink(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	form, errMsg := linkFromForm(c)
	if errMsg != "" {
		s.renderProjectLinks(c, p, http.StatusBadRequest, errMsg)
		return
	}
	_, err := s.q.CreateLink(c.Request.Context(), db.CreateLinkParams{
		UserID:    p.UserID,
		ProjectID: p.ID,
		Kind:      form.Kind,
		Url:       form.URL,
		Label:     form.Label,
		Notes:     form.Notes,
	})
	if err != nil {
		log.Printf("create link: %v", err)
		s.renderProjectLinks(c, p, http.StatusInternalServerError, "Could not add the link.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectLinks(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) updateLink(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	linkID, err := parseUUID(c.Param("link_id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	form, errMsg := linkFromForm(c)
	if errMsg != "" {
		s.renderProjectLinks(c, p, http.StatusBadRequest, errMsg)
		return
	}
	_, err = s.q.UpdateLink(c.Request.Context(), db.UpdateLinkParams{
		ID:        linkID,
		ProjectID: p.ID,
		UserID:    p.UserID,
		Kind:      form.Kind,
		Url:       form.URL,
		Label:     form.Label,
		Notes:     form.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		log.Printf("update link: %v", err)
		s.renderProjectLinks(c, p, http.StatusInternalServerError, "Could not save the link.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectLinks(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) deleteLink(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	linkID, err := parseUUID(c.Param("link_id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	n, err := s.q.DeleteLink(c.Request.Context(), db.DeleteLinkParams{
		ID:        linkID,
		ProjectID: p.ID,
		UserID:    p.UserID,
	})
	if err != nil {
		log.Printf("delete link: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectLinks(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) renderProjectLinks(c *gin.Context, p db.Project, status int, errMsg string) {
	links, err := s.loadProjectLinks(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	Render(c, status, views.ProjectLinks(csrfFrom(c), uuidStr(p.ID), links, errMsg))
}

func (s *Server) loadProjectLinks(c *gin.Context, projectID, userID pgtype.UUID) ([]views.Link, error) {
	rows, err := s.q.ListProjectLinks(c.Request.Context(), db.ListProjectLinksParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		log.Printf("list project links: %v", err)
		return nil, err
	}
	out := make([]views.Link, 0, len(rows))
	for _, l := range rows {
		out = append(out, toLinkView(l))
	}
	return out, nil
}

func linkFromForm(c *gin.Context) (views.Link, string) {
	url := strings.TrimSpace(c.PostForm("url"))
	kind := allowedKind(c.PostForm("kind"))
	if url == "" {
		return views.Link{}, "URL is required."
	}
	if kind == "" {
		return views.Link{}, "Pick a valid kind."
	}
	return views.Link{
		Kind:  kind,
		URL:   url,
		Label: strings.TrimSpace(c.PostForm("label")),
		Notes: strings.TrimSpace(c.PostForm("notes")),
	}, ""
}

func toLinkView(l db.Link) views.Link {
	return views.Link{
		ID:        uuidStr(l.ID),
		ProjectID: uuidStr(l.ProjectID),
		Kind:      l.Kind,
		URL:       l.Url,
		Label:     l.Label,
		Notes:     l.Notes,
	}
}

func allowedKind(s string) string {
	for _, k := range views.LinkKinds {
		if s == k {
			return s
		}
	}
	return ""
}
