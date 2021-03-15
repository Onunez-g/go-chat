package chat

type commandID int

const (
	CMD_NICK commandID = iota
	CMD_JOIN
	CMD_ADD
	CMD_REJECT
	CMD_ROOM
	CMD_ROOMS
	CMD_CHAT
	CMD_USERS
	CMD_LOG
	CMD_REQUESTS
	CMD_INVITES
	CMD_QUIT
	CMD_CLOSE
)

type Command struct {
	id     commandID
	client *Client
	args   []string
}
