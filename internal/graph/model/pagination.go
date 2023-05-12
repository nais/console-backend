package model

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
	return p
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
