package model

type SearchEdge struct {
	Node   SearchNode `json:"node"`
	Cursor Cursor     `json:"cursor"`
	Rank   int        `json:"-"`
}
