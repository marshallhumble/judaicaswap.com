package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	//aws
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (app *application) uploadFileToS3(fileName string) error {
	if fileName == "" {
		return nil
	}

	fullName := "./ui/static/SharePics/" + fileName
	file, err := os.Open(fullName)

	if err != nil {
		log.Printf("Couldn't open file %v to upload. Here's why: %v\n", fullName, err)
	} else {
		defer file.Close()
		_, err := app.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(app.S3Bucket + "/images/"),
			Key:    aws.String(fileName),
			Body:   file,
			ACL:    types.ObjectCannedACLPublicRead, // Optional: Set ACL for the file
		})

		if err != nil {
			return fmt.Errorf("failed to upload file to S3: %v", err)
		}

		_ = os.Remove(fullName)

		return nil

	}

	return nil
}

func (app *application) ListObjects(ctx context.Context, bucketName string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	var objects []types.Object
	objectPaginator := s3.NewListObjectsV2Paginator(app.S3Client, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", bucketName)
				err = noBucket
			}
			break
		} else {
			objects = append(objects, output.Contents...)
		}
	}
	return objects, err
}
