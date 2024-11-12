package member

import (
	"github.com/beglaryh/gocommon/stream"
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/item"
	"github.com/google/uuid"
)

type MemberItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	GSI1PK     string          `dynamodbav:"gsi1pk"`
	GSI1SK     string          `dynamodbav:"gsi1sk"`
	EntityType item.EntityType `dynamodbav:"entityType"`
	CreatedOn  string          `dynamodbav:"createdOn"`
	UserId     uuid.UUID       `dynamodbav:"userId"`
}

func From(room room.Room) []MemberItem {
	return stream.Map(room.Members, func(user uuid.UUID) MemberItem {
		return MemberItem{
			PK:         room.Id.String(),
			SK:         "MB#" + user.String(),
			GSI1PK:     user.String(),
			GSI1SK:     "R#" + room.Id.String(),
			UserId:     user,
			EntityType: item.Member,
			CreatedOn:  room.CreatedOn.String(),
		}
	}).Slice()
}
