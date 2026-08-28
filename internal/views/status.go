package views

import "strings"

func StatusLabel(status string) string {
	if status == "" {
		return ""
	}
	return strings.ToUpper(status[:1]) + status[1:]
}

func StatusChipClass(status string) string {
	switch status {
	case "active":
		return "chip chip-active"
	case "paused":
		return "chip chip-paused"
	case "shipped":
		return "chip chip-shipped"
	case "archived":
		return "chip chip-archived"
	default:
		return "chip"
	}
}

func TitleClass(status string) string {
	if status == "active" {
		return "row-title"
	}
	return "row-title is-quiet"
}

func FeedbackSourceClass(source string) string {
	if source == "ingest" {
		return "chip chip-ingest"
	}
	return "chip"
}
