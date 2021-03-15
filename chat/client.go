package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	conn        net.Conn
	nick        string
	commands    chan<- Command
	messages    []string
	invitations []string
}

func (c *Client) ReadInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.Trim(msg, "\r\n")
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])
		switch cmd {
		case "/ID":
			c.commands <- Command{
				id:     CMD_NICK,
				client: c,
				args:   args,
			}
		case "/JOIN":
			c.commands <- Command{
				id:     CMD_JOIN,
				client: c,
				args:   args,
			}
		case "/ADD":
			c.commands <- Command{
				id:     CMD_ADD,
				client: c,
				args:   args,
			}
		case "/REJECT":
			c.commands <- Command{
				id:     CMD_REJECT,
				client: c,
				args:   args,
			}
		case "/ROOM":
			c.commands <- Command{
				id:     CMD_ROOM,
				client: c,
				args:   args,
			}
		case "/ROOMLIST":
			c.commands <- Command{
				id:     CMD_ROOMLIST,
				client: c,
			}
		case "/CHAT":
			c.commands <- Command{
				id:     CMD_CHAT,
				client: c,
				args:   args,
			}
		case "/USERLIST":
			c.commands <- Command{
				id:     CMD_USERS,
				client: c,
			}
		case "/CHATLIST":
			c.commands <- Command{
				id:     CMD_LOG,
				client: c,
			}
		case "/REQUESTLIST":
			c.commands <- Command{
				id:     CMD_REQUESTS,
				client: c,
				args:   args,
			}
		case "/INVITELIST":
			c.commands <- Command{
				id:     CMD_INVITES,
				client: c,
			}
		case "/QUIT":
			c.commands <- Command{
				id:     CMD_QUIT,
				client: c,
				args:   args,
			}
		case "/CLOSE":
			c.commands <- Command{
				id:     CMD_CLOSE,
				client: c,
			}
		default:
			c.err(fmt.Errorf("Unknown command: %s", cmd))
		}
	}
}

func (c *Client) err(err error) {
	c.conn.Write([]byte("Error: " + err.Error() + "\n"))
}

func (c *Client) msg(msg string, args ...interface{}) {
	c.appendMessages(msg, args)
	c.conn.Write([]byte(msg + "\n"))
}

func (c *Client) self(msg string, args ...interface{}) {
	c.appendMessages(msg, args)
	c.conn.Write([]byte(msg + "\n"))
}

func (c *Client) appendMessages(msg string, args []interface{}) {
	var isLog bool
	var logMsg string
	for _, v := range args {
		switch v.(type) {
		case bool:
			isLog = v.(bool)
		case string:
			logMsg += v.(string)
		}
	}
	if isLog {
		m := logMsg + msg
		c.messages = append(c.messages, m)
	}
}
