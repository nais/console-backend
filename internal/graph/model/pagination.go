package model

import (
	"fmt"

	"github.com/nais/console-backend/internal/graph/scalar"
)

type Pagination struct {
	first  *int
	last   *int
	after  *scalar.Cursor
	before *scalar.Cursor
}

func NewPagination(first, last *int, after, before *scalar.Cursor) (*Pagination, error) {
	if first != nil && last != nil {
		return nil, fmt.Errorf("using both `first` and `last` with pagination is not supported")
	}
	return &Pagination{
		first:  first,
		last:   last,
		after:  after,
		before: before,
	}, nil
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

func (p *Pagination) After() *scalar.Cursor {
	if p.after == nil {
		return &scalar.Cursor{Offset: -1}
	}
	return p.after
}

func (p *Pagination) Before() *scalar.Cursor {
	return p.before
}
