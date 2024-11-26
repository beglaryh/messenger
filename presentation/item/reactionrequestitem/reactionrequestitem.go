package reactionrequestitem

import "github.com/beglaryh/messenger/domain/reaction"

type ReactionRequestItem struct {
	MID      string                `json:"id"`
	Reaction reaction.ReactionType `json:"reaction"`
}
