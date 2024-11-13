package item

type EntityType string

const (
	Room             EntityType = "R"
	Member           EntityType = "MBR"
	RoomMessage      EntityType = "RM"
	UserMessage      EntityType = "UM"
	Connection       EntityType = "C"
	ConnectionHeader EntityType = "CH"
)
