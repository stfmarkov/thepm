package httpx

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

// Render writes a templ component. Pages go through here, not c.HTML.
// c.JSON is reserved for the public feedback ingest route.
func Render(c *gin.Context, status int, component templ.Component) {
	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(err)
		if !c.Writer.Written() {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
