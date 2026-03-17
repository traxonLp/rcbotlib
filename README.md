
# rcbotlib — Documentation

**Module:** `github.com/traxonLp/rcbotlib`
**Go Version:** `1.22`
**Protocol Version:** `11.0.0`

A Go library for building bots for RCRS servers via WebSocket.

## Dependencies

- `github.com/google/uuid v1.6.0`
- `github.com/gorilla/websocket v1.5.1`

---
## Credentials
Make a `.env` file in your bots root. Fill in your credentials.
```
HOST_URL=<url>
BOT_USER=<auth-user>
BOT_PASS=<auth-password>
BOT_TOKEN=<bot-token>
```


--- 

## Quick Start
```
type MyBot struct{}

func (b *MyBot) OnMessage(msg *rcbotlib.Message, ctx *rcbotlib.Context) {
    msg.Reply("Hello!")
}
func (b *MyBot) OnEditMsg(edit rcbotlib.MsgEdit)                          {}
func (b *MyBot) OnDeleteMsg(id uuid.UUID)                                 {}
func (b *MyBot) OnTyping(info rcbotlib.Typing, ctx *rcbotlib.Context)     {}

func main() {
    conn, err := rcbotlib.Connect("rcrs.example.com", "user", "pass", "token")
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(conn.Run(&MyBot{}))
}
```
---

## connection.go

### Constant

go
const protocolVersion = "11.0.0"


The expected protocol version. If the server version does not match, the handshake will abort with an error.

---

### `Connect`

go
func Connect(host, user, pass, token string) (*Connection, error)


**Public entry point.** Establishes a fully authenticated WebSocket connection to the RCRS server. Internally calls `dial()` and `handshake()`.

| Parameter | Type     | Description                                                        |
|-----------|----------|--------------------------------------------------------------------|
| `host`    | `string` | Server hostname (without protocol, e.g. `rcrs.example.com`)       |
| `user`    | `string` | Username for Basic Auth                                            |
| `pass`    | `string` | Password for Basic Auth                                            |
| `token`   | `string` | Auth token, sent as cookie `auth-token=`                           |

**Returns:** `*Connection` on success, `error` on failure.

**Example:**
go
conn, err := rcbotlib.Connect("rcrs.example.com", "myUser", "myPass", "abc123")
if err != nil {
    log.Fatal(err)
}


---

### `dial` *(internal)*

go
func dial(host, user, pass, token string) (*websocket.Conn, error)


Opens the raw WebSocket connection to `wss://<host>/api/ws`. Sends the following headers:
- `Authorization: Basic <base64(user:pass)>`
- `Cookie: auth-token=<token>`

If the server responds with an HTTP error, the status code is included in the returned error.

---

### `handshake` *(internal)*

go
func handshake(conn *websocket.Conn, host, token string) (*Context, error)


Performs the protocol handshake over the established WebSocket connection.

**Flow:**
1. Reads the initial server message and checks `proto_version`
2. Compares with `protocolVersion` — mismatch → error
3. Sends `{"type": "version", "version": "11.0.0"}`
4. Reads the `hello` message and parses `user_info` via `parseDnsInfo()`
5. Returns a new `*Context` via `newContext()`

**Returns:** `*Context` on success, `error` at any failed step.

---

## handler.go

### `Handler` Interface

go
type Handler interface {
    OnMessage(msg *Message, ctx *Context)
    OnEditMsg(edit MsgEdit)
    OnDeleteMsg(id uuid.UUID)
    OnTyping(info Typing, ctx *Context)
}


Must be implemented by your bot. Each method is invoked in its own goroutine.

| Method         | Trigger                              |
|----------------|--------------------------------------|
| `OnMessage`    | A new message was received           |
| `OnEditMsg`    | An existing message was edited       |
| `OnDeleteMsg`  | A message was deleted                |
| `OnTyping`     | A user sent a typing indicator       |

---

### `Connection.Context`

go
func (c *Connection) Context() *Context


Returns the internal `*Context` for direct API access — useful outside of event callbacks.

---

### `Connection.Run`

go
func (c *Connection) Run(handler Handler) error


Starts the event loop. **Blocks** until the connection is closed or an error occurs. Dispatches incoming events to the provided `Handler`.

| `type` value   | Action                                              |
|----------------|-----------------------------------------------------|
| `new_msg`      | `handler.OnMessage()`                               |
| `edit_msg`     | `handler.OnEditMsg()`                               |
| `delete_msg`   | `handler.OnDeleteMsg()`                             |
| `start_typing` | `handler.OnTyping()`                                |
| `update_dns`   | DNS cache updated internally via `ctx.updateDNS()`  |

---

## context.go

`Context` provides all actions a bot can perform. It is thread-safe (mutex on WebSocket writes and DNS cache reads/writes).

### `newContext` *(internal)*

go
func newContext(conn *websocket.Conn, host, token string, userInfo DnsInfo) *Context


Creates a new `Context` with an authenticated HTTP client, base URL, and initial `UserInfo`. Only called by `handshake()`.

---

### `Context.SendMsg`

go
func (c *Context) SendMsg(chatID float64, content string, quoting ...uuid.UUID) error


Sends a message to a chat. The optional `quoting` argument quotes an existing message.

| Parameter | Type        | Description                                   |
|-----------|-------------|-----------------------------------------------|
| `chatID`  | `float64`   | Target chat ID                                |
| `content` | `string`    | Message content                               |
| `quoting` | `uuid.UUID` | *(optional)* ID of the message to quote       |

**Example:**
go
// Plain message
ctx.SendMsg(12345, "Hello!")

// With quote
ctx.SendMsg(12345, "Got it!", msgID)


---

### `Context.EditMsg`

go
func (c *Context) EditMsg(id uuid.UUID, append bool, content string) error


Edits a previously sent message. If `append` is `true`, `content` is appended instead of replacing the existing text.

| Parameter | Type        | Description                                 |
|-----------|-------------|---------------------------------------------|
| `id`      | `uuid.UUID` | ID of the message to edit                   |
| `append`  | `bool`      | `true` = append, `false` = replace          |
| `content` | `string`    | New or appended content                     |

---

### `Context.StartTyping`

go
func (c *Context) StartTyping(chatID float64) error


Sends a typing indicator to the specified chat.

---

### `Context.LookupUser`

go
func (c *Context) LookupUser(id int) (DnsInfo, bool)


Returns cached user information by user ID. The cache is automatically updated on `update_dns` events. Returns `false` if the user is not found in the cache.

---

### `Context.DMWith`

go
func (c *Context) DMWith(username string) (map[string]interface{}, error)


Creates or retrieves an existing DM chat with the given username (`POST /api/chats/create_dm`).

---

### `Context.UploadFile`

go
func (c *Context) UploadFile(data []byte, filename string) (map[string]interface{}, error)


Uploads a file as `multipart/form-data` (`POST /api/upload`) and returns the server response.

**Example:**
go
data, _ := os.ReadFile("image.png")
result, err := ctx.UploadFile(data, "image.png")


---

### `Context.GetAttachment`

go
func (c *Context) GetAttachment(url string) ([]byte, error)


Downloads raw attachment bytes from a URL using the authenticated HTTP client.

---

### `updateDNS` *(internal)*

go
func (c *Context) updateDNS(infos []DnsInfo)


Updates the internal user cache with a list of `DnsInfo` objects. Automatically called by `Run()` on `update_dns` events.

---

## http.go

### `authTransport.RoundTrip` *(internal)*

go
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error)


Implements `http.RoundTripper`. Automatically appends `Cookie: auth-token=<token>` to every outgoing HTTP request.

---

### `newHTTPClient` *(internal)*

go
func newHTTPClient(token string) *http.Client


Creates an `*http.Client` that automatically signs all requests with the auth token.

---

## parse.go

### `parseMessage` *(internal)*

go
func parseMessage(cmd map[string]interface{}, ctx *Context) (*Message, error)


Converts a raw WebSocket message (`map[string]interface{}`) into a typed `*Message`. Returns an error if the UUID is invalid.

---

### `parseDnsInfo` *(internal)*

go
func parseDnsInfo(raw interface{}) (DnsInfo, error)


Parses a `user_info` object from the handshake or a DNS update into a `DnsInfo` struct.

---

## types.go

### Types

| Type         | Description                                                              |
|--------------|--------------------------------------------------------------------------|
| `DnsInfo`    | User information: `ID`, `Username`, `Admin`, `Bot`, `State`              |
| `Message`    | Incoming message: `ID`, `ChatID`, `SenderID`, `Content`, `Attachments`   |
| `MsgEdit`    | Edited message info: `MsgID`, `Append`, `Content`, `ChatID`              |
| `Typing`     | Typing event: `UserID`, `ChatID`                                         |
| `Attachment` | File attachment: `ID`, `Path`, `ContentType`                             |

---

### `Message.Reply`

go
func (m *Message) Reply(content string) error


Replies to the message with an automatic quote. Shorthand for:
go
ctx.SendMsg(m.ChatID, content, m.ID)


---

### `Message.Sender`

go
func (m *Message) Sender() (DnsInfo, error)


Returns the `DnsInfo` of the message sender from the DNS cache. Returns an error if the user is not found in the cache.

---

### `Attachment.FullURL`

go
func (a *Attachment) FullURL() string


Returns the full absolute URL of the attachment (`host + path`).
