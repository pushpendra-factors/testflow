package s3

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

const (
	separator = "/"
)

type S3Driver struct {
	s3         *s3.S3
	BucketName string
	Region     string
}

func New(bucketName, region string) *S3Driver {
	session := session.New()
	s3 := s3.New(session, aws.NewConfig().WithRegion(region))
	return &S3Driver{s3: s3, BucketName: bucketName, Region: region}
}

func (sd *S3Driver) Create(dir, fileName string, reader io.ReadSeeker) error {

	log.WithFields(log.Fields{
		"Dir":        dir,
		"BucketName": sd.BucketName,
		"Region":     sd.Region,
	}).Debug("S3Driver Creating file")

	// add
	// SSE
	// content type
	// any key value metadata if needed
	input := &s3.PutObjectInput{
		Bucket: aws.String(sd.BucketName),
		Body:   reader,
		Key:    aws.String(dir + separator + fileName),
	}
	_, err := sd.s3.PutObject(input)
	return err
}

func (sd *S3Driver) Get(dir, fileName string) (io.ReadCloser, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(sd.BucketName),
		Key:    aws.String(dir + separator + fileName),
	}
	op, err := sd.s3.GetObject(&input)
	return op.Body, err
}

func (sd *S3Driver) GetProjectModelDir(projectId, modelId uint64) string {
	return fmt.Sprintf("projects/%d/models/%d/", projectId, modelId)
}

func (sd *S3Driver) GetModelEventInfoFilePathAndName(projectId, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("event_info_%d.txt", modelId)
}

func (sd *S3Driver) GetModelPatternsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("patterns_%d.txt", modelId)
}

func (sd *S3Driver) GetModelEventsFilePathAndName(projectId, modelId uint64) (string, string) {
	path := sd.GetProjectModelDir(projectId, modelId)
	return path, fmt.Sprintf("events_%d.txt", modelId)
}

func (sd *S3Driver) GetProjectsDataFilePathAndName(version string) (string, string) {
	return "metadata/", fmt.Sprintf("%s.txt", version)
}
