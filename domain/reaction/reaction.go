package reaction

type Reaction struct {
	Type ReactionType `json:"type"`
	By   string       `json:"by"`
}
