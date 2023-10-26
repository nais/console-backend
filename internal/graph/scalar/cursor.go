package scalar

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type Cursor struct {
	Offset int `json:"offset"`
}

func (c Cursor) MarshalGQLContext(_ context.Context, w io.Writer) error {
	v := url.Values{}
	v.Set("offset", strconv.Itoa(c.Offset))

	_, err := w.Write([]byte(strconv.Quote(base64.URLEncoding.EncodeToString([]byte(v.Encode())))))
	return err
}

func (c *Cursor) UnmarshalGQLContext(_ context.Context, v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("cursor must be a string")
	}

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	m, err := url.ParseQuery(string(b))
	if err != nil {
		return err
	}

	offset, err := strconv.Atoi(m.Get("offset"))
	if err != nil {
		return err
	}

	c.Offset = offset

	return nil
}
