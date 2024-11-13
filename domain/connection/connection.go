package connection

type Connection struct {
	ID  string
	UID string
}

func New(id, uid string) Connection {
	return Connection{id, uid}
}
