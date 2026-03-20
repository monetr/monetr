package secrets

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

type AWSKMSConfig struct {
	Log       *slog.Logger
	KeyID     string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  *string
}

type AWSKMS struct {
	log    *slog.Logger
	config AWSKMSConfig
	client *kms.Client
}

func NewAWSKMS(ctx context.Context, config AWSKMSConfig) (KeyManagement, error) {
	var configOptions []func(*awsconfig.LoadOptions) error

	if config.Region != "" {
		configOptions = append(configOptions, awsconfig.WithRegion(config.Region))
	}

	if config.AccessKey != "" {
		configOptions = append(configOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws kms session")
	}

	client := kms.NewFromConfig(cfg, func(o *kms.Options) {
		if config.Endpoint != nil && *config.Endpoint != "" {
			o.BaseEndpoint = aws.String(*config.Endpoint)
		}
	})

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
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
		EncryptionContext:   nil,
		GrantTokens:         []string{},
		KeyId:               aws.String(a.config.KeyID),
		Plaintext:           []byte(input),
	}

	response, err := a.client.Encrypt(span.Context(), request)
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
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault, // TODO Maybe make this a config thing?
		EncryptionContext:   nil,
		GrantTokens:         []string{},
		KeyId:               keyId,
	}

	response, err := a.client.Decrypt(span.Context(), request)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return "", errors.Wrap(err, "failed to decrypt data using AWS KMS")
	}

	span.Status = sentry.SpanStatusOK

	return string(response.Plaintext), nil
}
