package commands

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"
	"path/filepath"

	"log/slog"

	"github.com/monetr/monetr/server/config"
	"github.com/pkg/errors"
)

func loadCertificates(
	configuration config.Configuration,
	log *slog.Logger,
	generateCertificates bool,
) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	log.Debug("loading certificates",
		"privateKeyPath", configuration.Security.PrivateKey,
		"generateCertificates", generateCertificates,
	)

	var publicKey ed25519.PublicKey
	var privateKey ed25519.PrivateKey
	var ok bool

	{ // Parse the private key or generate one in memory.
		keyBytes, err := os.ReadFile(configuration.Security.PrivateKey)
		// If the private key path is blank or it cannot be found then execute this
		// block. This will conditionally generate certificates and persist the to
		// the disk. If the generate certificates flag is not provided though then
		// instead certificates will just be generated and stored in memory.
		if configuration.Security.PrivateKey == "" || os.IsNotExist(err) {

			if configuration.Security.PrivateKey == "" {
				log.Info("private key path not specified, a random security key will be generated")
			} else if generateCertificates {
				log.Info("private key does not exist, will generate a new one and store it at the specified path")
			} else {
				log.Info("private key does not exist, a random security key will be generated for this runtime")
			}

			publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to generate certificate")
			}

			privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to marshal private key")
			}
			privateKeyBlock := &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: privateKeyBytes,
			}
			privateKeyPem := pem.EncodeToMemory(privateKeyBlock)

			// If we are generating certificates then we should have a path and we can
			// write the to disk.
			if generateCertificates {
				directory, err := filepath.Abs(path.Dir(configuration.Security.PrivateKey))
				if err != nil {
					return nil, nil, errors.Wrap(err, "public key directory is not valid")
				}
				if err := os.MkdirAll(directory, 0600); err != nil {
					return nil, nil, errors.Wrap(err, "failed to create directory for certificates")
				}
				if err := os.WriteFile(
					configuration.Security.PrivateKey,
					privateKeyPem,
					0644,
				); err != nil {
					return nil, nil, errors.Wrap(err, "failed to write private key")
				}
			}

			return publicKey, privateKey, nil
		} else if err != nil {
			return nil, nil, errors.Wrap(err, "unable to read private key")
		}

		keyBlock, _ := pem.Decode(keyBytes)
		if keyBlock == nil {
			return nil, nil, errors.New("failed to decode PEM block containing private key")
		}

		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse private key")
		}

		privateKey, ok = key.(ed25519.PrivateKey)
		if !ok {
			return nil, nil, errors.New("provided private key is not an ED25519 private key")
		}

		publicKey, ok = privateKey.Public().(ed25519.PublicKey)
		if !ok {
			return nil, nil, errors.New("provided key does not contain a ED25519 public key")
		}

		return publicKey, privateKey, nil
	}
}
