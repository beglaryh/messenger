package messageroomitem

import (
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/infrastructure/item"
)

type MessageRoomItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	GSI1PK     string          `dynamodbav:"gsi1pk"`
	GSI1SK     string          `dynamodbav:"gsi1sk"`
	Message    string          `dynamodbav:"message"`
	EntityType item.EntityType `dynamodbav:"entityType"`
	CreatedOn  string          `dynamodbav:"createdOn"`
	CreatedBy  string          `dynamodbav:"createdBy"`
}

const (
	SK            = "M"
	GSI1SK_PREFIX = "M#"
)

func From(m message.Message) MessageRoomItem {
	return MessageRoomItem{
		PK:         m.Id,
		SK:         SK,
		GSI1PK:     m.RoomId,
		GSI1SK:     "M#" + m.Id,
		EntityType: item.RoomMessage,
		Message:    m.Message,
		CreatedOn:  m.CreatedOn.String(),
		CreatedBy:  m.SentBy,
	}
}

func (item MessageRoomItem) To() message.Message {
	createdOn, _ := offsetdatetime.Parse(item.CreatedOn)
	return message.Message{
		Id:        item.SK[2:],
		RoomId:    item.PK,
		Message:   item.Message,
		SentBy:    item.CreatedBy,
		CreatedOn: createdOn,
	}
}
