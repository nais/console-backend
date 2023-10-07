package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
	"net/url"
	"strings"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlpgx "cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

// CloseConnectionFuncs is a list of database connection close functions
type CloseConnectionFuncs []func() error

// Close will run all registered close functions and return the last error
func (fns CloseConnectionFuncs) Close() error {
	var err error
	for _, fn := range fns {
		if e := fn(); e != nil {
			err = e
		}
	}
	return err
}

//go:embed migrations/0*.sql
var embedMigrations embed.FS

// NewDB runs migrations and returns a new database connection
func NewDB(ctx context.Context, dsn string, log *logrus.Entry) (*gensql.Queries, CloseConnectionFuncs, error) {
	var closeFuncs CloseConnectionFuncs
	databaseDriver := "pgx"

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse dsn config: %w", err)
	}

	if !strings.HasPrefix(dsn, "postgres://") {
		databaseDriver = "cloudsql-postgres"
		dialer, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize dialer: %w", err)
		}
		closeFuncs = append(closeFuncs, dialer.Close)

		var instanceConnectionName string
		dsn, instanceConnectionName, err = ExtractInstanceConnectionNameFromDsn(dsn)
		if err != nil {
			return nil, closeFuncs, err
		}
		config.ConnConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.Dial(ctx, instanceConnectionName)
		}

		cleanup, err := cloudsqlpgx.RegisterDriver("cloudsql-postgres", cloudsqlconn.WithIAMAuthN())
		if err != nil {
			return nil, closeFuncs, err
		}
		closeFuncs = append(closeFuncs, cleanup)
	}

	if err := migrateDatabaseSchema(databaseDriver, dsn, log); err != nil {
		return nil, closeFuncs, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, closeFuncs, fmt.Errorf("failed to connect: %w", err)
	}
	closeFuncs = append(closeFuncs, func() error {
		conn.Close()
		return nil
	})

	return gensql.New(conn), closeFuncs, nil
}

// ExtractInstanceConnectionNameFromDsn will extract the instance connection name from the dsn and return the modified
// dsn, the instance connection name and a potential error. On error, the first two return values are both empty strings.
func ExtractInstanceConnectionNameFromDsn(dsn string) (string, string, error) {
	parts, err := url.ParseQuery(strings.ReplaceAll(dsn, " ", "&"))
	if err != nil {
		return "", "", err
	}
	instanceConnectionName := parts.Get("host")
	if instanceConnectionName == "" {
		return "", "", fmt.Errorf("dsn does not have a host field: %q", dsn)
	}
	delete(parts, "host")
	s, err := url.PathUnescape(parts.Encode())
	if err != nil {
		return "", "", err
	}
	updatedDsn := strings.ReplaceAll(s, "&", " ")
	return updatedDsn, instanceConnectionName, nil
}

// migrateDatabaseSchema runs database migrations
func migrateDatabaseSchema(driver, dsn string, log *logrus.Entry) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.WithError(err).Error("closing database migration connection")
		}
	}()

	return goose.Up(db, "migrations")
}
