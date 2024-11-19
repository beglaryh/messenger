package editrequest

import "github.com/beglaryh/gocommon/time/offsetdatetime"

type EditRequest struct {
	ModifiedOn offsetdatetime.OffsetDateTime
	MID        string
	Message    string
}
