package database

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"time"

	"log/slog"

	"github.com/monetr/monetr/server/certhelper"
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
		// No TLS unless a CA certificate is configured below, same as go-pg.
		pgdriver.WithInsecure(true),
	}

	var tlsConfiguration *tls.Config

	// TODO Make it so that the TLS config will work even when a CA certificate is
	// not being provided. This would be ideal for something where the PostgreSQL
	// TLS certificate is a well know certificate already included in the
	// certificate authority bundle on the OS.
	if configuration.PostgreSQL.CACertificatePath != "" {
		caCert, err := os.ReadFile(configuration.PostgreSQL.CACertificatePath)
		if err != nil {
			log.ErrorContext(context.Background(), "failed to load ca certificate", "err", err)
			return nil, errors.Wrap(err, "failed to load ca certificate")
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfiguration = &tls.Config{
			Rand:               rand.Reader,
			InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
			RootCAs:            caCertPool,
			ServerName:         configuration.PostgreSQL.Address,
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			MinVersion:         tls.VersionTLS12,
		}

		if configuration.PostgreSQL.KeyPath != "" {
			tlsCert, err := tls.LoadX509KeyPair(
				configuration.PostgreSQL.CertificatePath,
				configuration.PostgreSQL.KeyPath,
			)
			if err != nil {
				log.ErrorContext(context.Background(), "failed to load client certificate", "err", err)
				return nil, errors.Wrap(err, "failed to load client certificate")
			}
			tlsConfiguration.Certificates = []tls.Certificate{
				tlsCert,
			}
		}

		options = append(options, pgdriver.WithTLSConfig(tlsConfiguration))
	}

	connector := pgdriver.NewConnector(options...)
	sqldb := sql.OpenDB(connector)
	// Mirror go-pg's connection pool defaults.
	sqldb.SetMaxOpenConns(10 * runtime.NumCPU())
	sqldb.SetConnMaxIdleTime(5 * time.Minute)
	sqldb.SetConnMaxLifetime(9 * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(logging.NewPostgresHooks(log, stats))
	if configuration.PostgreSQL.CACertificatePath != "" {
		paths := make([]string, 0, 3)
		for _, path := range []string{
			configuration.PostgreSQL.CACertificatePath,
			configuration.PostgreSQL.KeyPath,
			configuration.PostgreSQL.CertificatePath,
		} {
			directory := filepath.Dir(path)
			if !slices.Contains(paths, directory) {
				paths = append(paths, directory)
			}
		}

		watchCertificate, err := certhelper.NewFileCertificateHelper(
			log,
			paths,
			func(_ string) error {
				log.InfoContext(context.Background(), "reloading TLS certificates")

				tlsConfig := &tls.Config{
					Rand:               rand.Reader,
					InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
					RootCAs:            nil,
					ServerName:         configuration.PostgreSQL.Address,
					Renegotiation:      tls.RenegotiateFreelyAsClient,
					MinVersion:         tls.VersionTLS12,
				}

				{
					caCert, err := os.ReadFile(configuration.PostgreSQL.CACertificatePath)
					if err != nil {
						log.ErrorContext(context.Background(), "failed to load updated ca certificate", "err", err)
						return errors.Wrap(err, "failed to load updated ca certificate")
					}

					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)

					log.DebugContext(context.Background(), "new ca certificate loaded, swapping")

					tlsConfig.RootCAs = caCertPool
				}

				{
					if configuration.PostgreSQL.KeyPath != "" {
						tlsCert, err := tls.LoadX509KeyPair(
							configuration.PostgreSQL.CertificatePath,
							configuration.PostgreSQL.KeyPath,
						)
						if err != nil {
							log.ErrorContext(context.Background(), "failed to load client certificate", "err", err)
							return errors.Wrap(err, "failed to load client certificate")
						}

						tlsConfig.Certificates = []tls.Certificate{
							tlsCert,
						}
					}
				}

				// Future connections will be established with the new TLS
				// configuration; existing connections are unaffected, matching the
				// old go-pg behavior of swapping Options().TLSConfig.
				connector.Config().TLSConfig = tlsConfig

				log.DebugContext(context.Background(), "successfully swapped ca certificate")

				return nil
			},
		)
		if err != nil {
			log.ErrorContext(context.Background(), "failed to setup certificate watcher", "err", err)
			return nil, errors.Wrap(err, "failed to setup certificate watcher")
		}
		watchCertificate.Start()

		defer watchCertificate.Stop()
	}

	if err := db.PingContext(context.Background()); err != nil {
		return db, errors.Wrap(err, "failed to ping postgresql")
	}

	if configuration.PostgreSQL.Migrate {
		log.InfoContext(context.Background(), "automatic migrations are enabled")
		migrations.RunMigrations(context.Background(), log, db)
	}

	return db, nil
}
