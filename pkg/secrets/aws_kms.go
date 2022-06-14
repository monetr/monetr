package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type AWSKMSConfig struct {
	Log       *logrus.Entry
	KeyID     string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  *string
}

type AWSKMS struct {
	log    *logrus.Entry
	config AWSKMSConfig
	client *kms.KMS
}

func NewAWSKMS(ctx context.Context, config AWSKMSConfig) (KeyManagement, error) {
	options := session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
				AccessKeyID:     config.AccessKey,
				SecretAccessKey: config.SecretKey,
			}),
			Region:     aws.String(config.Region),
			DisableSSL: nil,
			HTTPClient: nil,
			LogLevel:   nil,
			Logger:     nil,
			MaxRetries: nil,
			Retryer:    nil,
		},
	}

	if config.Endpoint != nil && *config.Endpoint != "" {
		options.Config.Endpoint = config.Endpoint
	}

	awsSession, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	client := kms.New(awsSession)
	return &AWSKMS{
		log:    config.Log,
		config: config,
		client: client,
	}, nil
}

func (a *AWSKMS) Encrypt(ctx context.Context, input []byte) (keyID string, version string, result []byte, _ error) {
	span := sentry.StartSpan(ctx, "Encrypt KMS")
	defer span.Finish()
	span.SetTag("kms", "aws")

	span.Data = map[string]interface{}{
		"resource": keyID,
	}

	request := &kms.EncryptInput{
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"),
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               aws.String(a.config.KeyID),
		Plaintext:           input,
	}

	response, err := a.client.EncryptWithContext(span.Context(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return "", "", nil, errors.Wrap(err, "failed to encrypt data using AWS KMS")
	}

	span.Status = sentry.SpanStatusOK

	return *response.KeyId, "", response.CiphertextBlob, nil
}

func (a *AWSKMS) Decrypt(ctx context.Context, keyID string, version string, input []byte) (result []byte, _ error) {
	span := sentry.StartSpan(ctx, "Decrypt KMS")
	defer span.Finish()
	span.SetTag("kms", "aws")

	span.Data = map[string]interface{}{
		"resource": keyID,
	}

	request := &kms.DecryptInput{
		CiphertextBlob:      input,
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"), // TODO Maybe make this a config thing?
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               aws.String(keyID),
	}

	response, err := a.client.DecryptWithContext(span.Context(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to decrypt data using AWS KMS")
	}

	span.Status = sentry.SpanStatusOK

	return response.Plaintext, nil
}
