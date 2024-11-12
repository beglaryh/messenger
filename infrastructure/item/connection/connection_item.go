package connection

import "github.com/beglaryh/messenger/infrastructure/item"

type ConnectionItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	GSI1PK     string          `dynamodbav:"gsi1pk"`
	GSI1SK     string          `dynamodbav:"gsi1sk"`
	EntityType item.EntityType `dynamodbav:"entityType"`
}

func New(connectionId, roomId, userId string) ConnectionItem {
	return ConnectionItem{
		PK:         connectionId,
		SK:         roomId,
		GSI1PK:     userId,
		GSI1SK:     "R#" + roomId,
		EntityType: item.Connection,
	}
}
