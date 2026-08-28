package httpx

import (
	"log"
	"net/http"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) createNote(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	body := strings.TrimSpace(c.PostForm("body"))
	if body == "" {
		s.renderProjectNotes(c, p, http.StatusBadRequest, "Note cannot be empty.")
		return
	}
	_, err := s.q.CreateNote(c.Request.Context(), db.CreateNoteParams{
		UserID:    p.UserID,
		ProjectID: p.ID,
		Body:      body,
	})
	if err != nil {
		log.Printf("create note: %v", err)
		s.renderProjectNotes(c, p, http.StatusInternalServerError, "Could not add the note.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectNotes(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) deleteNote(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	noteID, err := parseUUID(c.Param("note_id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	n, err := s.q.DeleteNote(c.Request.Context(), db.DeleteNoteParams{
		ID:        noteID,
		ProjectID: p.ID,
		UserID:    p.UserID,
	})
	if err != nil {
		log.Printf("delete note: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectNotes(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) renderProjectNotes(c *gin.Context, p db.Project, status int, errMsg string) {
	notes, err := s.loadProjectNotes(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	Render(c, status, views.ProjectNotes(csrfFrom(c), uuidStr(p.ID), notes, errMsg))
}

func (s *Server) loadProjectNotes(c *gin.Context, projectID, userID pgtype.UUID) ([]views.Note, error) {
	rows, err := s.q.ListProjectNotes(c.Request.Context(), db.ListProjectNotesParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		log.Printf("list project notes: %v", err)
		return nil, err
	}
	out := make([]views.Note, 0, len(rows))
	for _, n := range rows {
		out = append(out, toNoteView(n))
	}
	return out, nil
}

func toNoteView(n db.Note) views.Note {
	return views.Note{
		ID:        uuidStr(n.ID),
		ProjectID: uuidStr(n.ProjectID),
		Body:      n.Body,
		CreatedAt: formatTime(n.CreatedAt),
	}
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Local().Format("2006.01.02 15:04")
}
