package model

type Pagination struct {
	first      *int
	last       *int
	after      *Cursor
	before     *Cursor
	totalCount int
}

func NewPagination(first, last *int, after, before *Cursor, totalCount int) *Pagination {
	p := &Pagination{
		first:      first,
		last:       last,
		after:      after,
		before:     before,
		totalCount: totalCount,
	}
	return p
}

func (p *Pagination) Limit() int {
	if p.before != nil {
		return p.before.Offset
	}

	if p.Limit() > p.totalCount {
		return p.totalCount
	}

	return p.First() + p.After().Offset + 1
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
