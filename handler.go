package rcbotlib

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Handler is the interface your bot must implement.
type Handler interface {
	OnMessage(msg *Message, ctx *Context)
	OnEditMsg(edit MsgEdit)
	OnDeleteMsg(id uuid.UUID)
	OnTyping(info Typing, ctx *Context)
}

// Connection represents an active connection to an rcrs server.
type Connection struct {
	ctx *Context
}

// Context returns the connection's context for direct API access.
func (c *Connection) Context() *Context {
	return c.ctx
}

// Run starts the event loop, calling the handler for each incoming event.
// This function blocks until the connection is closed or an error occurs.
func (c *Connection) Run(handler Handler) error {
	for {
		var cmd map[string]interface{}
		if err := c.ctx.conn.ReadJSON(&cmd); err != nil {
			return fmt.Errorf("connection closed: %w", err)
		}

		switch cmd["type"] {
		case "new_msg":
			msg, err := parseMessage(cmd, c.ctx)
			if err != nil {
				continue
			}
			go handler.OnMessage(msg, c.ctx)

		case "edit_msg":
			go handler.OnEditMsg(MsgEdit{
				MsgID:   uuid.MustParse(cmd["id"].(string)),
				Append:  cmd["append"].(bool),
				Content: cmd["content"].(string),
				ChatID:  cmd["chat_id"].(float64),
			})

		case "delete_msg":
			go handler.OnDeleteMsg(uuid.MustParse(cmd["msg_id"].(string)))

		case "start_typing":
			go handler.OnTyping(Typing{
				UserID: int(cmd["user"].(float64)),
				ChatID: cmd["chat_id"].(float64),
			}, c.ctx)

		case "update_dns":
			raw, _ := json.Marshal(cmd["infos"])
			var infos []DnsInfo
			json.Unmarshal(raw, &infos)
			c.ctx.updateDNS(infos)
		}
	}
}
