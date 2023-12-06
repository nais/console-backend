package teams

import (
	"net/http"

	"github.com/nais/console-backend/internal/auth"
)

type httpClient struct {
	client     *http.Client
	apiToken   string
	onBehalfOf bool
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if c.onBehalfOf {
		ctx := req.Context()
		iapAssertion, err := auth.GetIAPAssertion(ctx)
		if err != nil {
			return nil, err
		}
		iapEmail, err := auth.GetIAPEmail(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Goog-IAP-JWT-Assertion", iapAssertion)
		req.Header.Set("X-Goog-Authenticated-User-Email", iapEmail)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiToken)
	}

	return c.client.Do(req)
}
