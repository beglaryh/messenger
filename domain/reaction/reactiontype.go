package reaction

type ReactionType string

const (
	ThumbsUp    = ReactionType("ThumbsUp")
	ThumbsDown  = ReactionType("ThumbsDown")
	Heart       = ReactionType("Heart")
	HaHa        = ReactionType("HaHa")
	Question    = ReactionType("Question")
	Exclamation = ReactionType("Exclamation")
)
