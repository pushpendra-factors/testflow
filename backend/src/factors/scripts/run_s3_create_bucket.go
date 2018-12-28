package main

import (
	"fmt"
	"os"

	c "factors/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {

	bucketName := "factors-dev"
	ed := "http://localhost:4572"
	region := endpoints.UsEast1RegionID

	session := session.Must(session.NewSession(&aws.Config{
		S3ForcePathStyle: aws.Bool(c.IsDevelopment()),
		Region:           aws.String(region),
		Endpoint:         aws.String(ed),
	}))

	svc := s3.New(session)

	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(endpoints.UsEast1RegionID),
		},
	}

	result, err := svc.CreateBucket(createBucketInput)
	if err != nil {
		fmt.Printf("Err: %v", err)
		return
	}

	fmt.Printf("Result %+v\n", *result)

	putBEncInput := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: []*s3.ServerSideEncryptionRule{
				&s3.ServerSideEncryptionRule{
					ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
						SSEAlgorithm: aws.String(s3.ServerSideEncryptionAes256),
					},
				},
			},
		},
	}
	_, err = svc.PutBucketEncryption(putBEncInput)
	if err != nil {
		fmt.Println("Got an error adding default KMS encryption to bucket", bucketName)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Bucket: " + bucketName + " now has KMS encryption by default")

}
