package rcbotlib

import (
	"fmt"
	"net/http"
)

type authTransport struct {
	token string
	user  string
	pass  string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Cookie", fmt.Sprintf("auth-token=%s", t.token))
	if t.user != "" && t.pass != "" {
		req.SetBasicAuth(t.user, t.pass)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func newHTTPClient(token, user, pass string) *http.Client {
	return &http.Client{
		Transport: &authTransport{token: token, user: user, pass: pass},
	}
}
