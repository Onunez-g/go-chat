package chat

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/onunez-g/go-chat/utils"
)

type Server struct {
	rooms    map[string]*Room
	commands chan Command
	clients  map[string]*Client
}

func NewServer() *Server {
	return &Server{
		rooms:    make(map[string]*Room),
		commands: make(chan Command),
		clients:  make(map[string]*Client),
	}
}

func (s *Server) Run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_ADD:
			s.add(cmd.client, cmd.args)
		case CMD_REJECT:
			s.reject(cmd.client, cmd.args)
		case CMD_ROOM:
			s.room(cmd.client, cmd.args)
		case CMD_ROOMLIST:
			s.roomlist(cmd.client)
		case CMD_CHAT:
			s.chat(cmd.client, cmd.args)
		case CMD_USERS:
			s.users(cmd.client)
		case CMD_LOG:
			s.log(cmd.client)
		case CMD_REQUESTS:
			s.requestlist(cmd.client, cmd.args)
		case CMD_INVITES:
			s.invitelist(cmd.client)
		case CMD_QUIT:
			s.quitCurrentRoom(cmd.client, cmd.args)
		case CMD_CLOSE:
			s.quit(cmd.client)
		}
	}
}

func (s *Server) NewClient(conn net.Conn) {
	log.Printf("A new client has connected: %s", conn.RemoteAddr().String())
	c := &Client{
		conn:        conn,
		nick:        "anonymous" + strconv.Itoa(rand.Intn(1000)),
		commands:    s.commands,
		invitations: make([]string, 0),
	}
	s.clients[c.nick] = c
	c.ReadInput()
}

func (s *Server) nick(c *Client, args []string) {
	nickname := args[1]
	_, ok := s.clients[nickname]
	_, ok2 := s.rooms[nickname]
	if strcase.ToCamel(nickname) != nickname {
		c.msg("NotValid")
		return
	}
	if !ok && !ok2 {
		delete(s.clients, c.nick)
		s.clients[nickname] = c
		c.nick = nickname
		c.msg(fmt.Sprintln("Ok"))
		return
	}
	c.msg("Taken")
}

func (s *Server) join(c *Client, args []string) {
	roomName := args[1]

	r, ok := s.rooms[roomName]
	if !ok {
		c.msg("NotFound")
		return
	}
	if _, ok := r.members[c.conn.RemoteAddr()]; ok {
		c.msg("Already")
		return
	}
	for k, v := range c.invitations {
		if v == roomName {
			r.members[c.conn.RemoteAddr()] = c
			c.invitations = append(c.invitations[:k], c.invitations[k+1:]...)
			c.msg("Ok")
			r.broadcast(c, fmt.Sprintf("/ROOMJOIN %s joined %s", c.nick, roomName))
			return
		}
	}
	var hasRequested bool
	for _, v := range r.requests {
		if v == c.nick {
			hasRequested = true
		}
	}
	if !hasRequested {
		r.requests = append(r.requests, c.nick)
	}

	c.msg("Ok")
	s.clients[r.owner].msg(fmt.Sprintf("/ROOMJOIN %s request-to-join %s", c.nick, roomName))
}

func (s *Server) reject(c *Client, args []string) {
	roomName := args[1]
	for k, v := range c.invitations {
		if roomName == v {
			c.invitations = append(c.invitations[:k], c.invitations[k+1:]...)
			owner := s.rooms[roomName].owner
			s.clients[owner].msg(fmt.Sprintf("/ROOMREJECT %s reject", c.nick))
			c.msg("Ok")
			return
		}
	}
}
func (s *Server) requestlist(c *Client, args []string) {
	roomName := args[1]

	r, ok := s.rooms[roomName]
	if !ok {
		c.msg("NotFound")
		return
	}
	if r.owner != c.nick {
		c.msg("NoOwner")
		return
	}

	c.msg(fmt.Sprint(r.requests))
}

func (s *Server) invitelist(c *Client) {
	c.msg(fmt.Sprint(c.invitations))
}

func (s *Server) roomlist(c *Client) {
	var rooms []string
	if len(s.rooms) == 0 {
		c.msg("Empty")
		return
	}
	for name := range s.rooms {
		rooms = append(rooms, name)
	}
	c.msg(fmt.Sprint(rooms))
}

func (s *Server) room(c *Client, args []string) {
	roomName := args[1]
	_, ok := s.rooms[roomName]
	_, ok2 := s.clients[roomName]
	if !ok && !ok2 {
		r := &Room{
			name:  roomName,
			owner: c.nick,
			members: map[net.Addr]*Client{
				c.conn.RemoteAddr(): c,
			},
			requests: make([]string, 0),
		}
		s.rooms[roomName] = r
		c.msg("Ok")
	} else {
		c.msg("Taken")
	}
}

func (s *Server) chat(c *Client, args []string) {
	var name string
	flagMessage := utils.FindIndex(args, "-m")
	if flagMessage == -1 {
		c.err(errors.New("Bad Syntax"))
	}
	var msg string = strings.Join(args[flagMessage+1:], " ")
	var chatType int
	if args[1] == "-u" {
		name = args[2]
		chatType = 1
	} else if args[1] == "-g" {
		name = args[2]
		chatType = 2
	}

	if chatType == 1 {
		client, ok := s.clients[name]
		if !ok {
			c.msg("NotFound")
			return
		}
		client.msg(fmt.Sprintf("/MESSAGE %s %s", c.nick, msg), true)
		c.self("Ok")
	} else if chatType == 2 {
		room, ok := s.rooms[name]
		if !ok {
			c.msg("NotFound")
			return
		}
		if _, ok := room.members[c.conn.RemoteAddr()]; !ok {
			c.msg("NotInRoom")
			return
		}
		room.broadcast(c, fmt.Sprintf("/MESSAGE %s_%s %s", name, c.nick, msg))
		c.self("Ok")
	} else {
		for k, v := range s.clients {
			if c.nick != k {
				v.msg(fmt.Sprintf("/MESSAGE %s %s", c.nick, msg))
				c.self("Ok")
			}
		}
	}
}

func (s *Server) add(c *Client, args []string) {
	var roomName string
	var users []string
	var isForced bool
	if args[1] == "-f" {
		roomName = args[2]
		users = args[3:]
		isForced = true
	} else {
		roomName = args[1]
		users = args[2:]
	}
	r, ok := s.rooms[roomName]
	if !ok {
		c.msg("NotFound")
		return
	}
	for _, v := range users {
		client, ok := s.clients[v]
		if !ok {
			c.msg(fmt.Sprintf("NotFound %s", v))
			continue
		}
		if _, ok := r.members[client.conn.RemoteAddr()]; ok {
			c.msg(fmt.Sprintf("Already %s", v))
			continue
		}
		if isForced {
			r.members[client.conn.RemoteAddr()] = client
			client.msg(fmt.Sprintf("/ADDED %s", roomName))
		} else {
			if index := utils.FindIndex(r.requests, client.nick); index != -1 {
				r.members[client.conn.RemoteAddr()] = client
				r.requests = append(r.requests[:index], r.requests[index+1:]...)
				client.msg(fmt.Sprintf("/ADDED %s", roomName))
			} else if index := utils.FindIndex(client.invitations, roomName); index == -1 {
				client.invitations = append(client.invitations, roomName)
				client.msg(fmt.Sprintf("/INVITED %s", roomName))
			}
		}
	}
	c.msg("Ok")
}

func (s *Server) users(c *Client) {
	var users []string
	if len(s.clients) == 0 {
		c.msg(fmt.Sprintln("Empty"))
		return
	}
	for nick := range s.clients {
		users = append(users, nick)
	}
	c.msg(fmt.Sprint(users))
}

func (s *Server) log(c *Client) {
	c.msg(fmt.Sprintf("Your message log:\n %s", strings.Join(c.messages, "\n")))
}

func (s *Server) quit(c *Client) {
	log.Printf("Client has disconnected: %s", c.conn.RemoteAddr().String())

	for _, r := range s.rooms {
		if m, ok := r.members[c.conn.RemoteAddr()]; ok {
			s.quitCurrentRoom(m, []string{"", r.name})
		}
	}
	s.quitServer(c)

	c.msg("Ok")

	c.conn.Close()
}

func (s *Server) quitCurrentRoom(c *Client, args []string) {
	roomName := args[1]
	room, ok := s.rooms[roomName]
	if ok {
		if room.owner == c.nick {
			delete(s.rooms, roomName)
			room.broadcast(c, fmt.Sprintf("/ROOMQUIT %s deleted %s", room.owner, room.name))
		} else {
			delete(room.members, c.conn.RemoteAddr())
			room.broadcast(c, fmt.Sprintf("/ROOMQUIT %s left %s", c.nick, roomName))
		}
	} else {
		c.msg("NotInRoom")
	}
}

func (s *Server) quitServer(c *Client) {
	delete(s.clients, c.nick)
}
