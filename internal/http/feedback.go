package httpx

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/AdventurousNerd/thepm/internal/db"
	"github.com/AdventurousNerd/thepm/internal/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) createFeedback(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	form, errMsg := feedbackFromForm(c)
	if errMsg != "" {
		s.renderProjectFeedback(c, p, http.StatusBadRequest, errMsg)
		return
	}
	_, err := s.q.CreateFeedback(c.Request.Context(), db.CreateFeedbackParams{
		UserID:      p.UserID,
		ProjectID:   p.ID,
		AuthorName:  form.AuthorName,
		AuthorEmail: optionalText(form.AuthorEmail),
		Message:     form.Message,
		Rating:      optionalRating(form.Rating),
		Source:      "manual",
	})
	if err != nil {
		log.Printf("create feedback: %v", err)
		s.renderProjectFeedback(c, p, http.StatusInternalServerError, "Could not add the feedback.")
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectFeedback(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) deleteFeedback(c *gin.Context) {
	p, ok := s.loadOwnedProject(c)
	if !ok {
		return
	}
	feedbackID, err := parseUUID(c.Param("feedback_id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	n, err := s.q.DeleteFeedback(c.Request.Context(), db.DeleteFeedbackParams{
		ID:        feedbackID,
		ProjectID: p.ID,
		UserID:    p.UserID,
	})
	if err != nil {
		log.Printf("delete feedback: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	if c.GetHeader("HX-Request") != "" {
		s.renderProjectFeedback(c, p, http.StatusOK, "")
		return
	}
	c.Redirect(http.StatusSeeOther, "/projects/"+uuidStr(p.ID))
}

func (s *Server) renderProjectFeedback(c *gin.Context, p db.Project, status int, errMsg string) {
	items, err := s.loadProjectFeedback(c, p.ID, p.UserID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	Render(c, status, views.ProjectFeedback(csrfFrom(c), toDetailView(c, p), items, errMsg))
}

func (s *Server) loadProjectFeedback(c *gin.Context, projectID, userID pgtype.UUID) ([]views.Feedback, error) {
	rows, err := s.q.ListProjectFeedback(c.Request.Context(), db.ListProjectFeedbackParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		log.Printf("list project feedback: %v", err)
		return nil, err
	}
	out := make([]views.Feedback, 0, len(rows))
	for _, f := range rows {
		out = append(out, toFeedbackView(f))
	}
	return out, nil
}

func feedbackFromForm(c *gin.Context) (views.Feedback, string) {
	name := strings.TrimSpace(c.PostForm("name"))
	message := strings.TrimSpace(c.PostForm("message"))
	email := strings.TrimSpace(c.PostForm("email"))
	rating := strings.TrimSpace(c.PostForm("rating"))
	if name == "" {
		return views.Feedback{}, "Name is required."
	}
	if message == "" {
		return views.Feedback{}, "Message is required."
	}
	if rating != "" {
		n, err := strconv.Atoi(rating)
		if err != nil || n < 1 || n > 5 {
			return views.Feedback{}, "Rating must be 1 to 5."
		}
	}
	return views.Feedback{
		AuthorName:  name,
		AuthorEmail: email,
		Message:     message,
		Rating:      rating,
	}, ""
}

func toFeedbackView(f db.Feedback) views.Feedback {
	rating := ""
	if f.Rating.Valid {
		rating = strconv.Itoa(int(f.Rating.Int32))
	}
	email := ""
	if f.AuthorEmail.Valid {
		email = f.AuthorEmail.String
	}
	return views.Feedback{
		ID:          uuidStr(f.ID),
		ProjectID:   uuidStr(f.ProjectID),
		AuthorName:  f.AuthorName,
		AuthorEmail: email,
		Message:     f.Message,
		Rating:      rating,
		Source:      f.Source,
		ReceivedAt:  formatTime(f.ReceivedAt),
	}
}

func optionalText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func optionalRating(s string) pgtype.Int4 {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Int4{}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(n), Valid: true}
}
