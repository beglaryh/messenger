package messageuseritem

import (
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/infrastructure/item"
	"github.com/beglaryh/messenger/infrastructure/item/reactionitem"
)

type MessageUserItem struct {
	PK         string                      `dynamodbav:"pk"`
	SK         string                      `dynamodbav:"sk"`
	GSI1PK     string                      `dynamodbav:"gsi1pk"`
	GSI1SK     string                      `dynamodbav:"gsi1sk"`
	RoomId     string                      `dynamodbav:"roomId"`
	Message    string                      `dynamodbav:"message"`
	EntityType item.EntityType             `dynamodbav:"entityType"`
	CreatedOn  string                      `dynamodbav:"createdOn"`
	CreatedBy  string                      `dynamodbav:"createdBy"`
	ModifiedOn string                      `dynamodbav:"modifiedOn"`
	Reactions  []reactionitem.ReactionItem `dynamodbav:"reactions"`
	IsEdited   bool                        `dyanmodbav:"isEdited"`
}

func From(m message.Message) *[]MessageUserItem {
	items := make([]MessageUserItem, len(m.Members))

	for i, user := range m.Members {
		items[i] = MessageUserItem{
			PK:         m.Id,
			SK:         user,
			GSI1PK:     user,
			GSI1SK:     "M#" + m.Id,
			RoomId:     m.RoomId,
			Message:    m.Message,
			EntityType: item.UserMessage,
			IsEdited:   m.IsEdited,
			CreatedOn:  m.CreatedOn.String(),
			CreatedBy:  m.SentBy,
		}
	}
	return &items
}

func To(items *[]MessageUserItem) []message.Message {
	messages := make([]message.Message, len(*items))
	for i, e := range *items {
		messages[i] = e.To()
	}
	return messages
}

func (item *MessageUserItem) To() message.Message {
	createdOn, _ := offsetdatetime.Parse(item.CreatedOn)
	modifiedOn, _ := offsetdatetime.Parse(item.ModifiedOn)
	return message.Message{
		Id:         item.PK,
		RoomId:     item.RoomId,
		SentBy:     item.CreatedBy,
		Message:    item.Message,
		IsEdited:   item.IsEdited,
		Reactions:  reactionitem.BatchTo(item.Reactions),
		CreatedOn:  createdOn,
		ModifiedOn: modifiedOn,
	}
}
