package database

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/certhelper"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/migrations"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GetDatabase(
	log *logrus.Entry,
	configuration config.Configuration,
	stats *metrics.Stats,
) (*pg.DB, error) {
	pgOptions := &pg.Options{
		Addr: fmt.Sprintf("%s:%d",
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
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
			log.WithError(err).Errorf("failed to load ca certificate")
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
		}

		if configuration.PostgreSQL.KeyPath != "" {
			tlsCert, err := tls.LoadX509KeyPair(
				configuration.PostgreSQL.CertificatePath,
				configuration.PostgreSQL.KeyPath,
			)
			if err != nil {
				log.WithError(err).Errorf("failed to load client certificate")
				return nil, errors.Wrap(err, "failed to load client certificate")
			}
			tlsConfiguration.Certificates = []tls.Certificate{
				tlsCert,
			}
		}

		pgOptions.TLSConfig = tlsConfiguration
		pgOptions.OnConnect = func(ctx context.Context, cn *pg.Conn) error {
			if tlsConfiguration != nil {
				log.Trace("new connection with cert")
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
				log.Info("reloading TLS certificates")

				tlsConfig := &tls.Config{
					Rand:               rand.Reader,
					InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
					RootCAs:            nil,
					ServerName:         configuration.PostgreSQL.Address,
					Renegotiation:      tls.RenegotiateFreelyAsClient,
				}

				{
					caCert, err := os.ReadFile(configuration.PostgreSQL.CACertificatePath)
					if err != nil {
						log.WithError(err).Errorf("failed to load updated ca certificate")
						return errors.Wrap(err, "failed to load updated ca certificate")
					}

					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)

					log.Debugf("new ca certificate loaded, swapping")

					tlsConfig.RootCAs = caCertPool
				}

				{
					if configuration.PostgreSQL.KeyPath != "" {
						tlsCert, err := tls.LoadX509KeyPair(
							configuration.PostgreSQL.CertificatePath,
							configuration.PostgreSQL.KeyPath,
						)
						if err != nil {
							log.WithError(err).Errorf("failed to load client certificate")
							return errors.Wrap(err, "failed to load client certificate")
						}

						tlsConfig.Certificates = []tls.Certificate{
							tlsCert,
						}
					}
				}

				db.Options().TLSConfig = tlsConfig

				log.Debugf("successfully swapped ca certificate")

				return nil
			},
		)
		if err != nil {
			log.WithError(err).Errorf("failed to setup certificate watcher")
			return nil, errors.Wrap(err, "failed to setup certificate watcher")
		}
		watchCertificate.Start()

		defer watchCertificate.Stop()
	}

	if err := db.Ping(context.Background()); err != nil {
		return db, errors.Wrap(err, "failed to ping postgresql")
	}

	if configuration.PostgreSQL.Migrate {
		log.Info("automatic migrations are enabled")
		migrations.RunMigrations(log, db)
	}

	return db, nil
}
