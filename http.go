package rcbotframework

import (
	"fmt"
	"net/http"
)

type authTransport struct {
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Cookie", fmt.Sprintf("auth-token=%s", t.token))
	return http.DefaultTransport.RoundTrip(req)
}

func newHTTPClient(token string) *http.Client {
	return &http.Client{
		Transport: &authTransport{token: token},
	}
}
