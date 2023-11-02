package graph

import (
	"time"

	"github.com/nais/console-backend/internal/graph/scalar"
)

func (r *queryResolver) getStartAndEnd(from *scalar.Date, to *scalar.Date) (start time.Time, end time.Time, err error) {
	end = time.Now()
	start = end.Add(-24 * time.Hour * 6)

	if to != nil {
		end, err = time.Parse(scalar.DateFormatYYYYMMDD, to.String())
		if err != nil {
			return
		}
	}

	if from != nil {
		start, err = time.Parse(scalar.DateFormatYYYYMMDD, from.String())
		if err != nil {
			return
		}
	}

	return
}
