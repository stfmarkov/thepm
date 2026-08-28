package httpx

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) listProjects(c *gin.Context) {
	uid, err := s.userID(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	rows, err := s.q.ListProjects(c.Request.Context(), uid)
	if err != nil {
		log.Printf("list projects: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	out := make([]views.Project, 0, len(rows))
	for _, p := range rows {
		out = append(out, toView(p))
	}
	u := currentUser(c)
	Render(c, http.StatusOK, views.ProjectList(u.Email, csrfFrom(c), out))
}

func (s *Server) newProject(c *gin.Context) {
	u := currentUser(c)
	Render(c, http.StatusOK, views.ProjectNew(u.Email, csrfFrom(c), "", views.Project{Status: "active"}))
}

func (s *Server) createProject(c *gin.Context) {
	uid, err := s.userID(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	form := projectFromForm(c)
	if form.Name == "" {
		Render(c, http.StatusBadRequest, views.ProjectNew(currentUser(c).Email, csrfFrom(c), "Name is required.", form))
		return
	}
	p, err := s.q.CreateProject(c.Request.Context(), db.CreateProjectParams{
		UserID:  uid,
		Name:    form.Name,
		Slug:    form.Slug,
		Status:  form.Status,
		Stack:   form.Stack,
		Summary: form.Summary,
	})
	if err != nil {
		msg := "Could not create the project."
		if isUniqueViolation(err) {
			msg = "That slug is already used."
		} else {
			log.Printf("create project: %v", err)
		}
		Render(c, http.StatusBadRequest, views.ProjectNew(currentUser(c).Email, csrfFrom(c), msg, form))
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) showProject(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	links, err := s.loadProjectLinks(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	notes, err := s.loadProjectNotes(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	feedback, err := s.loadProjectFeedback(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	u := currentUser(c)
	Render(c, http.StatusOK, views.ProjectDetail(u.Email, csrfFrom(c), toDetailView(c, p), links, notes, feedback))
}

func (s *Server) editProject(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	u := currentUser(c)
	Render(c, http.StatusOK, views.ProjectEdit(u.Email, csrfFrom(c), "", toView(p)))
}

func (s *Server) updateProject(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	form := projectFromForm(c)
	form.ID = uuidStr(p.ID)
	if form.Name == "" {
		Render(c, http.StatusBadRequest, views.ProjectEdit(currentUser(c).Email, csrfFrom(c), "Name is required.", form))
		return
	}
	updated, err := s.q.UpdateProject(c.Request.Context(), db.UpdateProjectParams{
		ID:      p.ID,
		UserID:  p.UserID,
		Name:    form.Name,
		Slug:    form.Slug,
		Status:  form.Status,
		Stack:   form.Stack,
		Summary: form.Summary,
	})
	if err != nil {
		msg := "Could not save the project."
		if isUniqueViolation(err) {
			msg = "That slug is already used."
		} else {
			log.Printf("update project: %v", err)
		}
		Render(c, http.StatusBadRequest, views.ProjectEdit(currentUser(c).Email, csrfFrom(c), msg, form))
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(updated.ID))
}

func (s *Server) deleteProject(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	n, err := s.q.DeleteProject(c.Request.Context(), db.DeleteProjectParams{ID: p.ID, UserID: p.UserID})
	if err != nil {
		log.Printf("delete project: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	if c.GetHeader("HX-Request") != "" {
		c.Status(http.StatusOK)
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func (s *Server) loadOwnedProject(c *gin.Context) (db.Project, bool) {
	uid, err := s.userID(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return db.Project{}, false
	}
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return db.Project{}, false
	}
	p, err := s.q.GetProject(c.Request.Context(), db.GetProjectParams{ID: id, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return db.Project{}, false
		}
		log.Printf("get project: %v", err)
		c.Status(http.StatusInternalServerError)
		return db.Project{}, false
	}
	return p, true
}

func (s *Server) userID(c *gin.Context) (pgtype.UUID, error) {
	return parseUUID(currentUser(c).ID)
}

func projectFromForm(c *gin.Context) views.Project {
	name := strings.TrimSpace(c.PostForm("name"))
	return views.Project{
		Name:    name,
		Slug:    slugify(name, c.PostForm("slug")),
		Status:  allowedStatus(c.PostForm("status")),
		Stack:   strings.TrimSpace(c.PostForm("stack")),
		Summary: strings.TrimSpace(c.PostForm("summary")),
	}
}

func toView(p db.Project) views.Project {
	return views.Project{
		ID:                uuidStr(p.ID),
		Name:              p.Name,
		Slug:              p.Slug,
		Status:            p.Status,
		Stack:             p.Stack,
		Summary:           p.Summary,
		FeedbackIngestKey: p.FeedbackIngestKey,
		FeedbackOrigin:    p.FeedbackOrigin,
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func uuidStr(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

func allowedStatus(s string) string {
	for _, st := range views.ProjectStatuses {
		if s == st {
			return s
		}
	}
	return "active"
}

var slugJunk = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name, slug string) string {
	s := strings.ToLower(strings.TrimSpace(slug))
	if s == "" {
		s = strings.ToLower(strings.TrimSpace(name))
	}
	s = slugJunk.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "project"
	}
	if len(s) > 80 {
		s = strings.Trim(s[:80], "-")
	}
	return s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
