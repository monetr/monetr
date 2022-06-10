package secrets

import (
	"context"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type GoogleKMSConfig struct {
	Log             *logrus.Entry
	KeyName         string
	URL             *string
	APIKey          *string
	CredentialsFile *string
}

type GoogleKMS struct {
	log    *logrus.Entry
	config GoogleKMSConfig
	client *kms.KeyManagementClient
}

func NewGoogleKMS(ctx context.Context, config GoogleKMSConfig) (KeyManagement, error) {
	options := make([]option.ClientOption, 0)
	if config.URL != nil && *config.URL != "" {
		options = append(options, option.WithEndpoint(*config.URL))
	}

	if config.APIKey != nil && *config.APIKey != "" {
		options = append(options, option.WithAPIKey(*config.APIKey))
	}

	if config.CredentialsFile != nil && *config.CredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(*config.CredentialsFile))
	}

	client, err := kms.NewKeyManagementClient(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create google key management client")
	}

	return &GoogleKMS{
		log:    config.Log,
		config: config,
		client: client,
	}, nil
}

func (g *GoogleKMS) Encrypt(input []byte) (keyID, version string, result []byte, _ error) {
	//TODO implement me
	panic("implement me")
}

func (g *GoogleKMS) Decrypt(keyID, version string, input []byte) (result []byte, _ error) {
	//TODO implement me
	panic("implement me")
}

