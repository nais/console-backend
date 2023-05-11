package model

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type Ident struct {
	ID   string
	Type string
}

func (i Ident) MarshalGQL(w io.Writer) {
	v := url.Values{}
	v.Set("id", i.ID)
	v.Set("type", i.Type)
	_, err := w.Write([]byte(strconv.Quote(base64.URLEncoding.EncodeToString([]byte(v.Encode())))))
	if err != nil {
		panic(err)
	}
}

func (i *Ident) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("id must be a string")
	}

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	m, err := url.ParseQuery(string(b))
	if err != nil {
		return err
	}

	i.ID = m.Get("id")
	i.Type = m.Get("type")

	return nil
}
