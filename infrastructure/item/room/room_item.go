package roomitem

import (
	"github.com/beglaryh/messenger/domain/room"
	"github.com/beglaryh/messenger/infrastructure/item"
	"github.com/google/uuid"
)

type RoomItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	EntityType item.EntityType `dynamodbav:"entityType"`

	CreatedOn string `dynamodbav:"createdOn"`
	CreatedBy string `dynamodbav:"CreatedBy"`

	Members []string `dynamodbav:"Members"`
}

const SK = "R"

func From(room room.Room) RoomItem {
	return RoomItem{
		PK:         room.Id.String(),
		SK:         SK,
		EntityType: item.Room,
		CreatedOn:  room.CreatedOn.String(),
		CreatedBy:  room.CreatedBy,
		Members:    room.Members,
	}
}

func (item RoomItem) To() room.Room {
	id, _ := uuid.Parse(item.PK)
	return room.Room{
		Id:      id,
		Members: item.Members,
	}
}
