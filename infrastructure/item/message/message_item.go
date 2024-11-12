package messageitem

import (
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/infrastructure/item"
)

type MessageItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	Message    string          `dynamodbav:"message"`
	EntityType item.EntityType `dynamodbav:"entityType"`
	CreatedOn  string          `dynamodbav:"createdOn"`
	CreatedBy  string          `dynamodbav:"createdBy"`
}

func From(m message.Message) MessageItem {
	return MessageItem{
		PK:         m.RoomId.String(),
		SK:         "M#" + m.Id.String(),
		EntityType: item.Message,
		Message:    m.Message,
		CreatedOn:  m.CreatedOn.String(),
		CreatedBy:  m.UserId.String(),
	}
}
