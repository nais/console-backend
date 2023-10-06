package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
	"net/url"
	"runtime"
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

// NewDB creates a new database connection and runs migrations
func NewDB(ctx context.Context, dsn string, log *logrus.Entry) (gensql.Querier, closeFuncs, error) {
	dbDriver := "pgx"
	isUrl := strings.Contains(dsn, "://")
	if !isUrl {
		dbDriver = "cloudsql-postgres"
	}

	if runtime.NumCPU() < 5 {
		if isUrl {
			dsn += "&pool_max_conns=5"
		} else {
			dsn += " pool_max_conns=5"
		}
	}

	cloudsql := dbDriver != "pgx"
	cloudsqlHost := ""
	if cloudsql {
		vals, err := url.ParseQuery(strings.ReplaceAll(dsn, " ", "&"))
		if err != nil {
			return nil, nil, err
		}
		cloudsqlHost = vals.Get("host")
		delete(vals, "host")
		dsn = strings.ReplaceAll(vals.Encode(), "&", " ")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	closers := closeFuncs{}

	if cloudsql {
		dialer, err := cloudsqlconn.NewDialer(ctx)
		if err != nil {
			return nil, closers, fmt.Errorf("failed to initialize dialer: %w", err)
		}
		closers = append(closers, dialer.Close)
		config.ConnConfig.DialFunc = func(ctx context.Context, _, instance string) (net.Conn, error) {
			return dialer.Dial(ctx, cloudsqlHost)
		}

		cleanup, err := cloudsqlpgx.RegisterDriver("cloudsql-postgres", cloudsqlconn.WithIAMAuthN())
		if err != nil {
			return nil, closers, err
		}
		closers = append(closers, cleanup)
	}

	if err := migrateDatabaseSchema(dbDriver, dsn, log); err != nil {
		return nil, closers, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, closers, fmt.Errorf("failed to connect: %w", err)
	}

	return gensql.New(conn), closers, nil
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
