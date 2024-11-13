package sendmessage

import "github.com/beglaryh/messenger/domain/message"

type SendMessageItem struct {
	Message message.Message `json:"message"`
}
