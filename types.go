package rcbotframework

import (
	"fmt"

	"github.com/google/uuid"
)

// DnsInfo holds user information cached from the server.
type DnsInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	Bot      bool   `json:"bot"`
	State    string `json:"state"`
}

// Message represents an incoming chat message.
type Message struct {
	ID          uuid.UUID    `json:"id"`
	ChatID      float64      `json:"chat_id"`
	SenderID    int          `json:"sender_id"`
	Content     string       `json:"content"`
	Quoting     *uuid.UUID   `json:"quoting"`
	Attachments []Attachment `json:"attachments"`
	ctx         *Context
}

// Reply sends a reply quoting this message.
func (m *Message) Reply(content string) error {
	return m.ctx.SendMsg(m.ChatID, content, m.ID)
}

// Sender returns the DnsInfo of the message sender.
func (m *Message) Sender() (DnsInfo, error) {
	info, ok := m.ctx.LookupUser(m.SenderID)
	if !ok {
		return DnsInfo{}, fmt.Errorf("[server bug]: user %d not in DNS", m.SenderID)
	}
	return info, nil
}

// MsgEdit contains information about an edited message.
type MsgEdit struct {
	MsgID   uuid.UUID `json:"id"`
	Append  bool      `json:"append"`
	Content string    `json:"content"`
	ChatID  float64   `json:"chat_id"`
}

// Typing contains information about a typing event.
type Typing struct {
	UserID int     `json:"user"`
	ChatID float64 `json:"chat_id"`
}

// Attachment represents a file attached to a message.
type Attachment struct {
	ID          uuid.UUID `json:"id"`
	Path        string    `json:"path"`
	ContentType string    `json:"media_type"`
	host        string
}

// FullURL returns the absolute URL of the attachment.
func (a *Attachment) FullURL() string {
	return a.host + a.Path
}
