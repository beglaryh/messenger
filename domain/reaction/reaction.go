package reaction

type Reaction struct {
	By   string       `json:"by"`
	Type ReactionType `json:"type"`
}
