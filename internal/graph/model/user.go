package model

type User struct {
	ID    Ident  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (User) IsNode()        {}
func (u User) GetID() Ident { return u.ID }
