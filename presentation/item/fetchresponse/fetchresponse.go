package fetchresponse

import "github.com/beglaryh/messenger/domain/message"

type FetchResponse struct {
	Messages []message.Message `json:"messages"`
}
