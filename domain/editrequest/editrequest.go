package editrequest

import "github.com/beglaryh/gocommon/time/offsetdatetime"

type EditRequest struct {
	ModifiedOn offsetdatetime.OffsetDateTime `json:"-"`
	MID        string                        `json:"id"`
	Message    string                        `json:"message"`
}
