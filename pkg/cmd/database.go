package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/certhelper"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/migrations"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func getDatabase(log *logrus.Entry, configuration config.Configuration, stats *metrics.Stats) (*pg.DB, error) {
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

	if configuration.PostgreSQL.CACertificatePath != "" {
		pgOptions.MaxConnAge = 9 * time.Minute
		{
			caCert, err := ioutil.ReadFile(configuration.PostgreSQL.CACertificatePath)
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
		}

		pgOptions.TLSConfig = tlsConfiguration
	}

	var db *pg.DB
	db = pg.Connect(pgOptions)
	db.AddQueryHook(logging.NewPostgresHooks(log, stats))
	pgOptions.OnConnect = func(ctx context.Context, cn *pg.Conn) error {
		log.Debugf("new connection with cert")

		return nil
	}

	if configuration.PostgreSQL.CACertificatePath != "" {
		paths := make([]string, 0, 1)
		for _, path := range []string{
			configuration.PostgreSQL.CACertificatePath,
			configuration.PostgreSQL.KeyPath,
			configuration.PostgreSQL.CertificatePath,
		} {
			directory := filepath.Dir(path)
			if !myownsanity.SliceContains(paths, directory) {
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
					caCert, err := ioutil.ReadFile(configuration.PostgreSQL.CACertificatePath)
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

	if MigrateDatabaseFlag {
		migrations.RunMigrations(log, db)
	} else {
		log.Info("automatic migrations are disabled")
	}

	return db, nil
}
