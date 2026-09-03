package httpx

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func onRender() bool {
	return strings.TrimSpace(os.Getenv("RENDER")) != ""
}

func applyGinMode() {
	if strings.TrimSpace(os.Getenv("GIN_MODE")) != "" {
		return
	}
	if onRender() {
		gin.SetMode(gin.ReleaseMode)
	}
}

func cookieSecure() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE"))) {
	case "true", "1":
		return true
	case "false", "0":
		return false
	}
	return onRender()
}

func trustProxies(r *gin.Engine) {
	if !onRender() {
		_ = r.SetTrustedProxies(nil)
		return
	}
	// Render's proxy is the only hop. Trust private/CGNAT ranges so ClientIP()
	// uses X-Forwarded-For instead of the load balancer address.
	_ = r.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
		"fc00::/7",
	})
}
