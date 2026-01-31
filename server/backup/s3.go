package backup

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type s3BackupDestination struct {
	log       *logrus.Entry
	chunkSize int
	bucket    string
	path      string
	session   *s3.S3
	uploadId  *string
	completed []*s3.CompletedPart
}

// Start is a blocking function that takes a context and a reader. It will
// consumme from the reader until the reader is finished and returns nothing or
// an end of file error. It will read chunks from the reader and upload them to
// an S3 bucket as part of a multipart upload.
func (d *s3BackupDestination) Start(ctx context.Context, reader io.Reader) error {
	log := d.log.WithContext(ctx)

	log.Info("starting multipart upload to S3 store")
	multiPartUpload, err := d.session.CreateMultipartUploadWithContext(
		ctx,
		&s3.CreateMultipartUploadInput{
			Bucket: aws.String(d.bucket),
			// TODO Content will always be tar.gz
			ContentEncoding: nil,
			// TODO Content will always be tar.gz
			ContentType: nil,
			Key:         aws.String(d.path),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create multipart upload")
	}

	d.uploadId = multiPartUpload.UploadId
	log = log.WithField("uploadId", *d.uploadId)

	partNumber := 1
	chunk := make([]byte, d.chunkSize)
	for {
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "failed to read chunk from backup manager")
		}
		if n == 0 {
			break
		}

		log.WithFields(logrus.Fields{
			"bytes": n,
			"part":  partNumber,
		}).Debug("uploading part of backup to S3 store")
		part, err := d.session.UploadPartWithContext(
			ctx,
			&s3.UploadPartInput{
				Body:          bytes.NewReader(chunk[:n]),
				Bucket:        aws.String(d.bucket),
				ContentLength: aws.Int64(int64(n)),
				Key:           aws.String(d.path),
				PartNumber:    aws.Int64(int64(partNumber)),
				UploadId:      multiPartUpload.UploadId,
			},
		)
		if err != nil {
			return errors.Wrap(err, "failed to upload part of backup")
		}

		log.WithFields(logrus.Fields{
			"bytes": n,
			"part":  partNumber,
		}).Info("successfully uploaded part of backup to S3 store")

		d.completed = append(d.completed, &s3.CompletedPart{
			ChecksumCRC32:  part.ChecksumCRC32,
			ChecksumCRC32C: part.ChecksumCRC32C,
			ChecksumSHA1:   part.ChecksumSHA1,
			ChecksumSHA256: part.ChecksumSHA256,
			ETag:           part.ETag,
			PartNumber:     aws.Int64(int64(partNumber)),
		})
		partNumber++
	}

	log.Info("finished uploading all parts of backup to S3 store")

	return nil
}

// Close completes the multipart upload to S3.
func (d *s3BackupDestination) Close() error {
	_, err := d.session.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(d.path),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: d.completed,
		},
		UploadId: d.uploadId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to complete multipart upload")
	}

	d.log.Info("completed multipart upload")

	return nil
}
