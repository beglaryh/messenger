package room

import (
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/google/uuid"
)

type Room struct {
	CreatedOn offsetdatetime.OffsetDateTime `json:"createdOn"`
	CreatedBy string                        `json:"createdBy"`
	Name      string                        `json:"name,omitempty"`
	Members   []string                      `json:"members"`
	Id        uuid.UUID                     `json:"id"`
}
