package message

import (
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/reaction"
)

type Message struct {
	CreatedOn  offsetdatetime.OffsetDateTime `json:"createdOn"`
	ModifiedOn offsetdatetime.OffsetDateTime `json:"modifiedOn,omitempty"`

	Id     string `json:"id"`
	RoomId string `json:"roomId"`
	SentBy string `json:"sentBy"`

	Message   string              `json:"message"`
	Members   []string            `json:"-"`
	Reactions []reaction.Reaction `json:"reactions,omitempty"`
	IsEdited  bool                `json:"isEdited"`
}
