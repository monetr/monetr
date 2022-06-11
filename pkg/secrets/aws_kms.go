package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type AWSKMSConfig struct {
	Log     *logrus.Entry
	KeyName string
	URL     *string
}

type AWSKMS struct {
	log    *logrus.Entry
	config AWSKMSConfig
	client *kms.KMS
}

func NewAWSKMS(ctx context.Context, config AWSKMSConfig) (KeyManagement, error) {
	return nil, nil
}

func (a *AWSKMS) Encrypt(ctx context.Context, input []byte) (keyID string, version string, result []byte, _ error) {
	request := &kms.EncryptInput{
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"),
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               aws.String(a.config.KeyName),
		Plaintext:           input,
	}

	response, err := a.client.EncryptWithContext(context.Background(), request)
	if err != nil {
		return "", "", nil, errors.Wrap(err, "failed to encrypt data using AWS KMS")
	}

	return *response.KeyId, "", response.CiphertextBlob, nil
}

func (a *AWSKMS) Decrypt(ctx context.Context, keyID string, version string, input []byte) (result []byte, _ error) {
	request := &kms.DecryptInput{
		CiphertextBlob:      input,
		EncryptionAlgorithm: aws.String("SYMMETRIC_DEFAULT"), // TODO Maybe make this a config thing?
		EncryptionContext:   nil,
		GrantTokens:         []*string{},
		KeyId:               aws.String(a.config.KeyName),
	}

	response, err := a.client.DecryptWithContext(context.Background(), request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data using AWS KMS")
	}

	return response.Plaintext, nil
}
