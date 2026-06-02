package tools

type artifactSummaryPayload struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}
