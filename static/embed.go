package static

import "embed"

// Vendored frontend assets so the app does not depend on a CDN at runtime
// for CSS, HTMX, or small helpers. Fonts are loaded from Google Fonts.
//
//	HTMX 2.0.10  — htmx.min.js
//
//go:embed htmx.min.js app.css app.js
var FS embed.FS
