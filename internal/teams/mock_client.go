package teams

import (
	"bytes"
	"net/http"
)

func (c *Client) UseMockClient() {
	c.httpClient = &http.Client{
		Transport: &mockClient{},
	}
}

type mockClient struct{}

func (m *mockClient) RoundTrip(req *http.Request) (*http.Response, error) {
	buf := mockBuffer{
		Buffer: bytes.NewBufferString(`{"data":{"userByEmail":{"name":"test user","id":"06c3fc83-b789-439a-a045-7028c4ec5c88","teams":[{"team":{"slug":"thomas"}}, {"team":{"slug":"test"}}]}}}`),
	}

	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       &buf,
	}, nil
}

type mockBuffer struct {
	*bytes.Buffer
}

func (m *mockBuffer) Close() error {
	return nil
}
