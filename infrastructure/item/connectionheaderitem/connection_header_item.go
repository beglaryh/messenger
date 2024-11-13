package connectionheaderitem

import (
	"github.com/beglaryh/messenger/domain/connection"
	"github.com/beglaryh/messenger/infrastructure/item"
)

type ConnectionHeaderItem struct {
	PK         string          `dynamodbav:"pk"`
	SK         string          `dynamodbav:"sk"`
	GSI1PK     string          `dynamodbav:"gsi1pk"`
	GSI1SK     string          `dynamodbav:"gsi1sk"`
	EntityType item.EntityType `dynamodbav:"entityType"`
}

const (
	CH_SK            = "CH"
	CH_GSI1SK_PREXIF = "CH#"
)

func New(connectionId, userId string) ConnectionHeaderItem {
	return ConnectionHeaderItem{
		PK:         connectionId,
		SK:         CH_SK,
		GSI1PK:     userId,
		GSI1SK:     CH_GSI1SK_PREXIF + connectionId,
		EntityType: item.ConnectionHeader,
	}
}

func From(connection connection.Connection) ConnectionHeaderItem {
	return New(connection.ID, connection.UID)
}

func (ch *ConnectionHeaderItem) To() connection.Connection {
	return connection.New(ch.PK, ch.GSI1PK)
}
