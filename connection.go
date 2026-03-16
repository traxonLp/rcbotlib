package rcbotlib

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

const protocolVersion = "11.0.0"

// Connect establishes an authenticated connection to an rcrs server.
// If this fails, you're fucked
func Connect(host, user, pass, token string) (*Connection, error) {
	conn, err := dial(host, user, pass, token)
	if err != nil {
		return nil, err
	}

	ctx, err := handshake(conn, host, token)
	if err != nil {
		return nil, err
	}

	return &Connection{ctx: ctx}, nil
}

func dial(host, user, pass, token string) (*websocket.Conn, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))

	conn, resp, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("wss://%s/api/ws", host),
		http.Header{
			"Cookie":        {fmt.Sprintf("auth-token=%s", token)},
			"Authorization": {"Basic " + auth},
		},
	)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("connection failed (HTTP %d): %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
}

func handshake(conn *websocket.Conn, host, token string) (*Context, error) {
	var hs map[string]interface{}
	if err := conn.ReadJSON(&hs); err != nil {
		return nil, fmt.Errorf("handshake read failed: %w", err)
	}

	serverVersion, ok := hs["proto_version"].(string)
	if !ok {
		return nil, fmt.Errorf("handshake missing proto_version")
	}
	if serverVersion != protocolVersion {
		return nil, fmt.Errorf("protocol version mismatch — server: %s, client: %s", serverVersion, protocolVersion)
	}

	if err := conn.WriteJSON(map[string]interface{}{
		"type":    "version",
		"version": protocolVersion,
	}); err != nil {
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	var hello map[string]interface{}
	if err := conn.ReadJSON(&hello); err != nil {
		return nil, fmt.Errorf("hello read failed: %w", err)
	}

	userInfo, err := parseDnsInfo(hello["user_info"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return newContext(conn, host, token, userInfo), nil
}
