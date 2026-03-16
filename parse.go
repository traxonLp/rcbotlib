package rcbotframework

import (
	"fmt"

	"github.com/google/uuid"
)

func parseMessage(cmd map[string]interface{}, ctx *Context) (*Message, error) {
	id, err := uuid.Parse(cmd["id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid message id: %w", err)
	}

	return &Message{
		ID:       id,
		ChatID:   cmd["chat_id"].(float64),
		SenderID: int(cmd["sender_id"].(float64)),
		Content:  cmd["content"].(string),
		ctx:      ctx,
	}, nil
}

func parseDnsInfo(raw interface{}) (DnsInfo, error) {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return DnsInfo{}, fmt.Errorf("invalid user_info format")
	}

	return DnsInfo{
		ID:       int(m["id"].(float64)),
		Username: m["username"].(string),
		Admin:    m["admin"].(bool),
		Bot:      m["bot"].(bool),
		State:    m["state"].(string),
	}, nil
}
