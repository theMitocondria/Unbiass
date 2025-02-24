package awsHandler 

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

func DeleteContentFromS3(bucketName , key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(key),
	}

	_, err := s3Client.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}
