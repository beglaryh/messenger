package editrequestitem

import (
	"github.com/beglaryh/gocommon/time/offsetdatetime"
	"github.com/beglaryh/messenger/domain/editrequest"
)

type EditRequest struct {
	Message EditRequestItem `json:"message"`
}

type EditRequestItem struct {
	MID     string `json:"messageId"`
	Message string `json:"message"`
}

func (e *EditRequestItem) To() editrequest.EditRequest {
	return editrequest.EditRequest{
		MID:        e.MID,
		Message:    e.Message,
		ModifiedOn: offsetdatetime.Now(),
	}
}
