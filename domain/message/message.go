package message

import (
	"github.com/beglaryh/gocommon/time/localdate"
	"github.com/google/uuid"
)

type Message struct {
	CreatedOn  localdate.LocalDate
	ModifiedOn localdate.LocalDate

	Message string

	Id     uuid.UUID
	RoomId uuid.UUID
	UserId uuid.UUID

	IsEdited bool
}
