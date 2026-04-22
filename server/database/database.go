package database

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/certhelper"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/migrations"
	"github.com/pkg/errors"
)

func GetDatabase(
	log *slog.Logger,
	configuration config.Configuration,
	stats *metrics.Stats,
) (*pg.DB, error) {
	pg.SetLogger(logging.NewPGLogger(log))
	pgOptions := &pg.Options{
		Addr: net.JoinHostPort(
			configuration.PostgreSQL.Address,
			strconv.Itoa(configuration.PostgreSQL.Port),
		),
		User:            configuration.PostgreSQL.Username,
		Password:        configuration.PostgreSQL.Password,
		Database:        configuration.PostgreSQL.Database,
		ApplicationName: "monetr",
		MaxConnAge:      9 * time.Minute,
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

		pgOptions.TLSConfig = tlsConfiguration
		pgOptions.OnConnect = func(ctx context.Context, cn *pg.Conn) error {
			if tlsConfiguration != nil {
				log.Log(ctx, logging.LevelTrace, "new connection with cert")
			}

			return nil
		}
	}

	db := pg.Connect(pgOptions)
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
			func(path string) error {
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

				db.Options().TLSConfig = tlsConfig

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

	if err := db.Ping(context.Background()); err != nil {
		return db, errors.Wrap(err, "failed to ping postgresql")
	}

	if configuration.PostgreSQL.Migrate {
		log.InfoContext(context.Background(), "automatic migrations are enabled")
		migrations.RunMigrations(log, db)
	}

	return db, nil
}
