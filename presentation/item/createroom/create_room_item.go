package createroom

import "github.com/beglaryh/messenger/domain/room"

type CreateRoomItem struct {
	Message room.Room `json:"message"`
}
