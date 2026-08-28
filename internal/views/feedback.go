package views

type Feedback struct {
	ID          string
	ProjectID   string
	AuthorName  string
	AuthorEmail string
	Message     string
	Rating      string
	Source      string
	ReceivedAt  string
}

var FeedbackRatings = []string{"1", "2", "3", "4", "5"}
