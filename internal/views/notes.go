package views

type Note struct {
	ID        string
	ProjectID string
	Body      string
	CreatedAt string
}

func ComposerNote(draft *Note) Note {
	if draft == nil {
		return Note{}
	}
	return *draft
}
