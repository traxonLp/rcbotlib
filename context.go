package rcbotlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Context provides all actions a bot can perform.
type Context struct {
	UserInfo   DnsInfo
	conn       *websocket.Conn
	httpClient *http.Client
	baseURL    string
	connMu     sync.Mutex
	dnsMu      sync.RWMutex
	dns        map[int]DnsInfo
}

func newContext(conn *websocket.Conn, host, token string, userInfo DnsInfo) *Context {
	return &Context{
		conn:       conn,
		httpClient: newHTTPClient(token),
		baseURL:    fmt.Sprintf("https://%s", host),
		UserInfo:   userInfo,
		dns:        make(map[int]DnsInfo),
	}
}

// SendMsg sends a message to a chat, optionally quoting another message.
func (c *Context) SendMsg(chatID float64, content string, quoting ...uuid.UUID) error {
	msg := map[string]interface{}{
		"type":        "new_msg",
		"id":          uuid.New(),
		"chat_id":     chatID,
		"msg":         content,
		"attachments": []interface{}{},
	}
	if len(quoting) > 0 {
		msg["quoting"] = quoting[0]
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn.WriteJSON(msg)
}

// EditMsg edits an existing message. If append is true, content is appended instead of replaced.
func (c *Context) EditMsg(id uuid.UUID, append bool, content string) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn.WriteJSON(map[string]interface{}{
		"type":    "edit_msg",
		"id":      id,
		"append":  append,
		"content": content,
	})
}

// StartTyping sends a typing indicator to a chat.
func (c *Context) StartTyping(chatID float64) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn.WriteJSON(map[string]interface{}{
		"type":    "start_typing",
		"chat_id": chatID,
	})
}

// LookupUser returns cached user info by ID.
func (c *Context) LookupUser(id int) (DnsInfo, bool) {
	c.dnsMu.RLock()
	defer c.dnsMu.RUnlock()
	info, ok := c.dns[id]
	return info, ok
}

// DMWith creates or retrieves a DM chat with the given username.
func (c *Context) DMWith(username string) (map[string]interface{}, error) {
	body, _ := json.Marshal(map[string]string{"with": username})
	resp, err := c.httpClient.Post(c.baseURL+"/api/chats/create_dm", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// UploadFile uploads a file and returns the server response.
func (c *Context) UploadFile(data []byte, filename string) (map[string]interface{}, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, err
	}
	part.Write(data)
	w.Close()

	resp, err := c.httpClient.Post(c.baseURL+"/api/upload", w.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, b)
	}

	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return result, nil
}

// GetAttachment downloads raw attachment bytes from a URL.
func (c *Context) GetAttachment(url string) ([]byte, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Context) updateDNS(infos []DnsInfo) {
	c.dnsMu.Lock()
	defer c.dnsMu.Unlock()
	for _, info := range infos {
		c.dns[info.ID] = info
	}
}

// SetProfilePicture uploads a profile picture from your ahh pc
func (c *Context) SetProfilePicture(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	result, err := c.UploadFile(data, filepath.Base(path))
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"upload_id": result["id"].(string)})
	resp, err := c.httpClient.Post(c.baseURL+"/api/register_user_icon", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}
