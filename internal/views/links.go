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

type LinkGroup struct {
	Kind  string
	Links []Link
}

func GroupLinks(links []Link) []LinkGroup {
	buckets := make(map[string][]Link, len(LinkKinds))
	for _, l := range links {
		buckets[l.Kind] = append(buckets[l.Kind], l)
	}
	groups := make([]LinkGroup, 0, len(LinkKinds))
	for _, k := range LinkKinds {
		if items := buckets[k]; len(items) > 0 {
			groups = append(groups, LinkGroup{Kind: k, Links: items})
		}
	}
	return groups
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

func LinkTitle(l Link) string {
	if s := strings.TrimSpace(l.Label); s != "" {
		return s
	}
	return l.URL
}
