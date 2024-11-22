package requestitem

type EventItem struct {
	Message RequestItem `json:"message"`
}

type RequestItem struct {
	Action  string `json:"action"`
	Message any    `json:"message"`
}
