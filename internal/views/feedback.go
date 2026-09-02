package views

import "strconv"

type Feedback struct {
	ID           string
	ProjectID    string
	AuthorName   string
	AuthorEmail  string
	Message      string
	Rating       string
	Source       string
	ReceivedAt   string
	ReceivedUnix string
}

type FeedbackRatingStats struct {
	Average string
	Count   int
}

var FeedbackRatings = []string{"1", "2", "3", "4", "5"}

var FeedbackSorts = []struct {
	Value string
	Label string
}{
	{"newest", "Newest"},
	{"oldest", "Oldest"},
	{"highest", "Highest"},
	{"lowest", "Lowest"},
}

func FeedbackRatingStatsFrom(items []Feedback) FeedbackRatingStats {
	sum := 0
	n := 0
	for _, f := range items {
		v, err := strconv.Atoi(f.Rating)
		if err != nil || v < 1 || v > 5 {
			continue
		}
		sum += v
		n++
	}
	if n == 0 {
		return FeedbackRatingStats{}
	}
	avg := float64(sum) / float64(n)
	return FeedbackRatingStats{
		Average: strconv.FormatFloat(avg, 'f', 1, 64),
		Count:   n,
	}
}

func (s FeedbackRatingStats) CountLabel() string {
	if s.Count == 1 {
		return "1 rating"
	}
	return strconv.Itoa(s.Count) + " ratings"
}
