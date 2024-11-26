package reactionitem

import (
	"github.com/beglaryh/messenger/domain/reaction"
)

type ReactionItem struct {
	Type reaction.ReactionType `dynamodbav:"type"`
	By   string                `dynamodbav:"by"`
}

func From(r reaction.Reaction) ReactionItem {
	return ReactionItem{r.Type, r.By}
}

func BatchTo(items []ReactionItem) []reaction.Reaction {
	reactions := make([]reaction.Reaction, len(items))
	for i, item := range items {
		reactions[i] = item.To()
	}
	return reactions
}

func (item ReactionItem) To() reaction.Reaction {
	return reaction.Reaction{
		Type: item.Type,
		By:   item.By,
	}
}
