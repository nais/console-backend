package apierror

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	ErrInternal = Errorf("The server errored out while processing your request, and we didn't write a suitable error message. You might consider that a bug on our side. Please try again, and if the error persists, contact the NAIS team.")
	ErrDatabase = Errorf("The database system encountered an error while processing your request. This is probably a transient error, please try again. If the error persists, contact the NAIS team.")
)

// Error is an error that can be presented to end-users
type Error struct {
	err error
}

func (e Error) Error() string {
	return e.err.Error()
}

// Errorf formats an error message for end-users. Remember not to leak sensitive information in error messages
func Errorf(format string, args ...any) Error {
	return Error{
		err: fmt.Errorf(format, args...),
	}
}

// GetErrorPresenter returns a GraphQL error presenter that filters out error messages not intended for end users.
// Filtered errors will be logged with the original error attached.
func GetErrorPresenter(log logrus.FieldLogger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		gqlError := graphql.DefaultErrorPresenter(ctx, err)
		unwrappedError := errors.Unwrap(err)

		switch unwrappedError.(type) {
		default:
			break
		case Error:
			return gqlError // err is already formatted for end-user
		case *pgconn.PgError:
			gqlError.Message = ErrDatabase.Error()
			log.WithError(err).Errorf("database error")
			return gqlError
		}

		switch {
		default:
			log.WithError(err).Errorf("unhandled error in the GraphQL error presenter")
			gqlError.Message = ErrInternal.Error()
		case errors.Is(unwrappedError, sql.ErrNoRows):
			gqlError.Message = "Object was not found in the database."
		case errors.Is(unwrappedError, context.Canceled):
			gqlError.Message = "Request canceled."
		}

		return gqlError
	}
}
