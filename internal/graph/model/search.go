package model

type SearchEdge struct {
	Node   SearchNode `json:"node"`
	Cursor string     `json:"cursor"`
	Rank   int        `json:"-"`
}
