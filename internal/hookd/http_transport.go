package hookd

import (
	"net/http"
)

type Transport struct {
	PSK string
}

func (t Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-PSK", t.PSK)
	return http.DefaultTransport.RoundTrip(req)
}
