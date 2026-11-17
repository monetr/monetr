package encryption

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type awsWrapper struct {
	log    *slog.Logger
	client *kms.Client
	keyArn string
}

func newAWSWrapper(
	ctx context.Context,
	log *slog.Logger,
	config config.AWSKMS,
) (*awsWrapper, error) {
	var configOptions []func(*awsconfig.LoadOptions) error

	configOptions = append(configOptions,
		awsconfig.WithLogger(logging.NewAWSLogger(log)),
		awsconfig.WithClientLogMode(aws.LogRetries|aws.LogDeprecatedUsage|aws.LogSigning),
	)

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
			o.BaseEndpoint = new(*config.Endpoint)
		}
	})

	// Config may hold an alias or a bare key ID. The wrapper stores the ARN
	// returned by Encrypt, so resolve the config to an ARN before comparing.
	described, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(config.KeyID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to describe configured KMS key")
	}
	return &awsWrapper{
		log:    log,
		client: client,
		keyArn: *described.KeyMetadata.Arn,
	}, nil
}

func (a *awsWrapper) Kind() models.RootKeyKind {
	return models.RootKeyKindAWS
}

func (a *awsWrapper) Matches(w models.RootKeyWrapper) bool {
	return w.Kind == models.RootKeyKindAWS && w.Reference == a.keyArn
}

func (a *awsWrapper) Wrap(
	ctx context.Context,
	rootKey, aad []byte,
) (models.RootKeyWrapper, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	a.log.DebugContext(span.Context(),
		"wrapping root key using AWS KMS",
		"awsKeyId", a.keyArn,
	)
	result, err := a.client.Encrypt(span.Context(), &kms.EncryptInput{
		KeyId:               aws.String(a.keyArn),
		Plaintext:           rootKey,
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
		EncryptionContext: map[string]string{
			"aad": string(aad),
		},
	})
	if err != nil {
		a.log.ErrorContext(span.Context(),
			"failed to wrap root key using AWS KMS",
			"awsKeyId", a.keyArn,
			"err", err,
		)
		return models.RootKeyWrapper{}, errors.Wrap(err, "KMS failed to wrap root key")
	}

	return models.RootKeyWrapper{
		Kind: models.RootKeyKindAWS,
		// Matches compares against the ARN, so store the ARN.
		Reference: a.keyArn,
		Data:      models.EncryptionEncoding.EncodeToString(result.CiphertextBlob),
	}, nil
}

func (a *awsWrapper) Unwrap(
	ctx context.Context,
	wrapper models.RootKeyWrapper,
	aad []byte,
) ([]byte, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	blob, err := models.EncryptionEncoding.DecodeString(wrapper.Data)
	if err != nil {
		return nil, errors.Wrap(err, "KMS wrapper is not valid base32")
	}
	a.log.DebugContext(span.Context(),
		"unwrapping root key using AWS KMS",
		"awsKeyId", a.keyArn,
	)
	result, err := a.client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob:      blob,
		KeyId:               aws.String(a.keyArn), // KMS rejects it if a different key made it
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecSymmetricDefault,
		EncryptionContext: map[string]string{
			"aad": string(aad),
		},
	})
	if err != nil {
		a.log.ErrorContext(span.Context(),
			"failed to unwrap root key using AWS KMS",
			"awsKeyId", a.keyArn,
			"err", err,
		)
		return nil, errors.Wrap(err, "KMS failed to unwrap root key")
	}
	return result.Plaintext, nil
}
