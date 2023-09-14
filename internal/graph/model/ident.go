package model

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/types"
)

type IdentType string

const (
	IdentTypeApp       IdentType = "app"
	IdentTypeDeployKey IdentType = "deployKey"
	IdentTypeEnv       IdentType = "env"
	IdentTypeJob       IdentType = "job"
	IdentTypePod       IdentType = "pod"
	IdentTypeTeam      IdentType = "team"
	IdentTypeUser      IdentType = "user"
)

type Ident struct {
	ID   string
	Type IdentType
}

func (i Ident) MarshalGQL(w io.Writer) {
	if i.ID == "" || i.Type == "" {
		panic(fmt.Errorf("id and type must be set"))
	}
	v := url.Values{}
	v.Set("id", i.ID)
	v.Set("type", string(i.Type))
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
	i.Type = IdentType(m.Get("type"))

	return nil
}

func AppIdent(id string) Ident {
	return newIdent(id, IdentTypeApp)
}

func DeployKeyIdent(id string) Ident {
	return newIdent(id, IdentTypeDeployKey)
}

func EnvIdent(id string) Ident {
	return newIdent(id, IdentTypeEnv)
}

func JobIdent(id string) Ident {
	return newIdent(id, IdentTypeJob)
}

func PodIdent(id types.UID) Ident {
	return newIdent(string(id), IdentTypePod)
}

func TeamIdent(id string) Ident {
	return newIdent(id, IdentTypeTeam)
}

func UserIdent(id string) Ident {
	return newIdent(id, IdentTypeUser)
}

func newIdent(id string, t IdentType) Ident {
	return Ident{
		ID:   id,
		Type: t,
	}
}
