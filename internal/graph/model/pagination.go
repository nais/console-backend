package model

import "fmt"

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *Cursor `json:"startCursor,omitempty"`
	EndCursor       *Cursor `json:"endCursor,omitempty"`
}

type Pagination struct {
	first  *int
	last   *int
	after  *Cursor
	before *Cursor
}

func NewPagination(first, last *int, after, before *Cursor) *Pagination {
	p := &Pagination{
		first:  first,
		last:   last,
		after:  after,
		before: before,
	}
	fmt.Printf("first %#v\n", *first)
	return p
}

func (p *Pagination) ForSlice(length int) (start, end int) {
	length -= 1
	if p.before != nil {
		start = p.before.Offset - p.Last()
		end = p.before.Offset
	} else {
		start = p.After().Offset + 1
		end = start + p.First()
	}

	if start > length {
		start = length
	}

	if start < 0 {
		start = 0
	}

	if end < 0 {
		end = 0
	}
	if end > length {
		end = length + 1
	}
	return start, end
}

func (p *Pagination) First() int {
	if p.first == nil {
		return 10
	}
	return *p.first
}

func (p *Pagination) Last() int {
	if p.last == nil {
		return 10
	}
	return *p.last
}

func (p *Pagination) After() *Cursor {
	if p.after == nil {
		return &Cursor{Offset: -1}
	}
	return p.after
}

func (p *Pagination) Before() *Cursor {
	return p.before
}
