package application

import (
	"github.com/beglaryh/messenger/domain/message"
	"github.com/beglaryh/messenger/infrastructure/database"
)

type Application struct {
	db *database.DB
}

func (a *Application) SaveMessage(m message.Message) error {
	db.SaveMessage(m)
}
