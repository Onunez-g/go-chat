package chat

import (
	"fmt"
	"net"
)

type Room struct {
	name     string
	owner    string
	requests []string
	members  map[net.Addr]*Client
}

func (r *Room) broadcast(sender *Client, msg string) {
	for addr, m := range r.members {
		if sender.conn.RemoteAddr() != addr {
			m.msg(msg, true, fmt.Sprintf("from %s to room %s: ", sender.nick, r.name))
		}
	}
}
