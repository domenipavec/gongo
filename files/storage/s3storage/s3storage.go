package s3storage

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

type S3Storage struct {
	uploader *s3manager.Uploader
	client   *s3.S3

	bucket     string
	signedUrls bool
}

func New(awsSession *session.Session, bucket string, signedUrls bool) (*S3Storage, error) {
	region, err := s3manager.GetBucketRegion(context.Background(), awsSession, bucket, "us-east-1")
	if err != nil {
		return nil, errors.Wrapf(err, "could not locate bucket %s region", bucket)
	}
	awsSession = awsSession.Copy(aws.NewConfig().WithRegion(region))

	s := &S3Storage{
		bucket:     bucket,
		uploader:   s3manager.NewUploader(awsSession),
		client:     s3.New(awsSession),
		signedUrls: signedUrls,
	}

	return s, nil
}

func (s *S3Storage) URL(name string) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &name,
	})
	if s.signedUrls {
		url, err := req.Presign(time.Hour)
		if err != nil {
			return "", errors.Wrap(err, "could not presign url")
		}
		return url, nil
	}

	err := req.Build()
	if err != nil {
		return "", errors.Wrap(err, "could not build url")
	}

	return req.HTTPRequest.URL.String(), nil
}

func (s *S3Storage) Save(name string, input io.Reader) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: &s.bucket,
		Key:    &name,
		Body:   input,
	})
	if err != nil {
		return errors.Wrapf(err, "could not upload file %s", name)
	}

	return nil
}

func (s *S3Storage) Delete(name string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &name,
	})
	if err != nil {
		return errors.Wrapf(err, "could not delete file %s", name)
	}

	return nil
}
