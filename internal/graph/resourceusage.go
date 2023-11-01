package graph

import (
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
)

func (r *queryResolver) getStartEndAndStep(from *scalar.Date, to *scalar.Date, resolution *model.Resolution) (start time.Time, end time.Time, step time.Duration, err error) {
	end = time.Now()
	start = end.Add(-24 * time.Hour * 6)
	step = 24 * time.Hour

	if resolution != nil && *resolution == model.ResolutionHourly {
		step = time.Hour
	}

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
