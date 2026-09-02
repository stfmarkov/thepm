package views

import (
	"strings"

	"github.com/a-h/templ"
)

var LinkKinds = []string{"repo", "site", "hosting", "domain", "docs", "other"}

var linkKindLabels = map[string]string{
	"repo":    "Repository",
	"site":    "Live site",
	"hosting": "Hosting",
	"domain":  "Domain",
	"docs":    "Docs",
	"other":   "Other",
}

type Link struct {
	ID        string
	ProjectID string
	Kind      string
	URL       string
	Label     string
	Notes     string
}

func LinkKindLabel(kind string) string {
	if s, ok := linkKindLabels[kind]; ok {
		return s
	}
	return kind
}

func LinkHref(raw string) templ.SafeURL {
	u := strings.TrimSpace(raw)
	if u == "" {
		return ""
	}
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return templ.SafeURL(u)
	}
	return templ.SafeURL("https://" + u)
}

func ComposerLink(draft *Link) Link {
	if draft == nil {
		return Link{Kind: LinkKinds[0]}
	}
	out := *draft
	if out.Kind == "" {
		out.Kind = LinkKinds[0]
	}
	return out
}
