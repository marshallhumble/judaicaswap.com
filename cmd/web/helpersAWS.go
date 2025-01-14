package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

func (app *application) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	return nil
}

func (app *application) putPresignURL(objectName, objectType string) string {

	presignedUrl, err := app.PresignClient.PresignPutObject(context.Background(),
		&s3.PutObjectInput{
			Bucket:      aws.String(app.S3Bucket + "/images/"),
			Key:         aws.String(objectName),
			ContentType: aws.String(objectType),
		},
		s3.WithPresignExpires(time.Minute*15))

	if err != nil {
		log.Fatal(err)
	}

	return presignedUrl.URL
}

func (app *application) updateObjectACL(bucketName, objectKey, acl string) error {

	/*
		If we try to do this too soon it won't work/will throw a CORS error or permission denied, this is because
		the file isn't actually fully written to s3 and replicated so we need to wait a few seconds until we try to add
		permissions/set the ACL
	*/
	time.Sleep(2 * time.Second)

	_, err := app.S3Client.PutObjectAcl(context.TODO(), &s3.PutObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("images/" + objectKey),
		ACL:    types.ObjectCannedACL(acl),
	})

	if err != nil {
		return err
	}

	return nil
}
