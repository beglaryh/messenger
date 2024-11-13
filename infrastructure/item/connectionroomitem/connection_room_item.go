package connectionroomitem

import "github.com/beglaryh/messenger/infrastructure/item"

type ConnectionRoomItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	GSI1PK     string          `dynamodbav:"gsi1pk"`
	GSI1SK     string          `dynamodbav:"gsi1sk"`
	UserId     string          `dynamodbav:"userId"`
	EntityType item.EntityType `dynamodbav:"entityType"`
}

func New(connectionId, roomId, userId string) ConnectionRoomItem {
	return ConnectionRoomItem{
		PK:         connectionId,
		SK:         roomId,
		GSI1PK:     roomId,
		GSI1SK:     "U#" + userId,
		UserId:     userId,
		EntityType: item.Connection,
	}
}
