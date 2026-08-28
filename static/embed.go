package static

import "embed"

// Vendored frontend assets so the app does not depend on a CDN at runtime.
//
//	HTMX 2.0.10  — htmx.min.js
//	Pico CSS 2.1.1 — pico.min.css
//
//go:embed htmx.min.js pico.min.css
var FS embed.FS
