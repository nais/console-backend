package search

import (
	"context"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/nais/console-backend/internal/graph/model"
)

type Filters struct {
	Type string
}

type Searchable interface {
	Search(ctx context.Context, q string, filters Filters) []*model.SearchEdge
}

type Searcher struct {
	searchables []Searchable
}

func New(s ...Searchable) *Searcher {
	return &Searcher{searchables: s}
}

func (s *Searcher) Search(ctx context.Context, q string, filters Filters) []*model.SearchEdge {
	ret := []*model.SearchEdge{}
	for _, searchable := range s.searchables {
		edges := searchable.Search(ctx, q, filters)
		ret = append(ret, edges...)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Rank < ret[j].Rank
	})

	return ret
}

// Match returns the rank of a match between q and val. 0 means best match. -1 means no match.
func Match(q, val string) int {
	return fuzzy.RankMatchFold(q, val)
}
