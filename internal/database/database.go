package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/nais/console-backend/internal/database/gensql"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlpgx "cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

type closeFuncs []func() error

func (c closeFuncs) Close() error {
	var err error
	for _, f := range c {
		if e := f(); e != nil {
			err = e
		}
	}
	return err
}

//go:embed migrations/0*.sql
var embedMigrations embed.FS

func NewDB(ctx context.Context, dbConnDSN string, cloudsql bool) (*gensql.Queries, closeFuncs, error) {
	cloudsqlHost := ""
	if cloudsql {
		vals, err := url.ParseQuery(strings.ReplaceAll(dbConnDSN, " ", "&"))
		if err != nil {
			return nil, nil, err
		}
		cloudsqlHost = vals.Get("host")
		delete(vals, "host")
		dbConnDSN = strings.ReplaceAll(vals.Encode(), "&", " ")
	}

	config, err := pgxpool.ParseConfig(dbConnDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	closers := closeFuncs{}

	if cloudsql {
		// Create a new dialer with any options
		d, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			return nil, closers, fmt.Errorf("failed to initialize dialer: %w", err)
		}
		closers = append(closers, d.Close)

		// Tell the driver to use the Cloud SQL Go Connector to create connections
		config.ConnConfig.DialFunc = func(ctx context.Context, _ string, instance string) (net.Conn, error) {
			return d.Dial(ctx, cloudsqlHost)
		}

		cleanup, err := cloudsqlpgx.RegisterDriver("cloudsql-postgres", cloudsqlconn.WithIAMAuthN())
		if err != nil {
			return nil, closers, err
		}
		closers = append(closers, cleanup)
	}

	// Interact with the dirver directly as you normally would
	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, closers, fmt.Errorf("failed to connect: %w", err)
	}

	return gensql.New(conn), closers, nil
}

func Migrate(driver, dsn string, log *logrus.Entry) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log)

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
