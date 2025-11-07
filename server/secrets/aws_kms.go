package secrets

import (
	"context"
	"encoding/hex"

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

func (a *AWSKMS) Encrypt(ctx context.Context, input string) (keyId, version *string, result string, _ error) {
	span := sentry.StartSpan(ctx, "Encrypt KMS")
	defer span.Finish()
	span.SetTag("kms", "aws")

	span.Data = map[string]any{
		"resource": keyId,
	}

	request := &kms.EncryptInput{
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"),
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               aws.String(a.config.KeyID),
		Plaintext:           []byte(input),
	}

	response, err := a.client.EncryptWithContext(span.Context(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, nil, "", errors.Wrap(err, "failed to encrypt data using AWS KMS")
	}

	span.Status = sentry.SpanStatusOK

	encrypted := hex.EncodeToString(response.CiphertextBlob)
	return response.KeyId, nil, encrypted, nil
}

func (a *AWSKMS) Decrypt(ctx context.Context, keyId, version *string, encrypted string) (result string, _ error) {
	span := sentry.StartSpan(ctx, "Decrypt KMS")
	defer span.Finish()
	span.SetTag("kms", "aws")

	span.Data = map[string]any{
		"resource": keyId,
	}

	input, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode encrypted secret")
	}

	request := &kms.DecryptInput{
		CiphertextBlob:      input,
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"), // TODO Maybe make this a config thing?
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               keyId,
	}

	response, err := a.client.DecryptWithContext(span.Context(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return "", errors.Wrap(err, "failed to decrypt data using AWS KMS")
	}

	span.Status = sentry.SpanStatusOK

	return string(response.Plaintext), nil
}
