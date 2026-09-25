package database

import (
	"context"
	"database/sql"
	"net"
	"runtime"
	"strconv"
	"time"

	"log/slog"

	"github.com/monetr/monetr/server/certs"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/migrations"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func GetDatabase(
	log *slog.Logger,
	configuration config.Configuration,
	stats *metrics.Stats,
) (*bun.DB, error) {
	options := []pgdriver.Option{
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(net.JoinHostPort(
			configuration.PostgreSQL.Address,
			strconv.Itoa(configuration.PostgreSQL.Port),
		)),
		pgdriver.WithUser(configuration.PostgreSQL.Username),
		pgdriver.WithPassword(configuration.PostgreSQL.Password),
		pgdriver.WithDatabase(configuration.PostgreSQL.Database),
		pgdriver.WithApplicationName("monetr"),
		// go-pg did not apply socket timeouts by default; pgdriver does. Keep
		// the old behavior, long running things like migrations rely on it.
		pgdriver.WithReadTimeout(0),
		pgdriver.WithWriteTimeout(0),
		// No TLS unless it is enabled below.
		pgdriver.WithInsecure(true),
	}

	// If no CA certificate is provided then the server certificate is verified
	// against the system certificate authorities instead.
	if configuration.PostgreSQL.TLS ||
		configuration.PostgreSQL.CACertificatePath != "" ||
		configuration.PostgreSQL.CertificatePath != "" ||
		configuration.PostgreSQL.KeyPath != "" {
		certificates, err := certs.NewFileSource(log, certs.Options{
			CACertificatePath:  configuration.PostgreSQL.CACertificatePath,
			CertificatePath:    configuration.PostgreSQL.CertificatePath,
			KeyPath:            configuration.PostgreSQL.KeyPath,
			ServerName:         configuration.PostgreSQL.Address,
			InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
		})
		if err != nil {
			log.ErrorContext(context.Background(), "failed to load TLS certificates", "err", err)
			return nil, errors.Wrap(err, "failed to load TLS certificates")
		}
		// The certificates are watched for the lifetime of the process, new
		// connections will pick up rotated certificates without swapping the TLS
		// config.
		certificates.Start()

		options = append(options, pgdriver.WithTLSConfig(certificates.ClientConfig()))
	}

	connector := pgdriver.NewConnector(options...)
	sqldb := sql.OpenDB(connector)
	// Mirror go-pg's connection pool defaults.
	sqldb.SetMaxOpenConns(10 * runtime.NumCPU())
	sqldb.SetMaxIdleConns(10 * runtime.NumCPU())
	sqldb.SetConnMaxIdleTime(5 * time.Minute)
	sqldb.SetConnMaxLifetime(9 * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(logging.NewPostgresHooks(log, stats))

	if err := db.PingContext(context.Background()); err != nil {
		return db, errors.Wrap(err, "failed to ping postgresql")
	}

	if configuration.PostgreSQL.Migrate {
		log.InfoContext(context.Background(), "automatic migrations are enabled")
		migrations.RunMigrations(context.Background(), log, db)
	}

	return db, nil
}
