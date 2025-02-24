package awsHandler

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
	"fmt"
)

func UploadContentToS3( key, content , bucketName string ) error { 
    input := &s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(key),
        Body:   strings.NewReader(content),
    }

    _, err := s3Client.PutObject(input)
    if err != nil {
        return fmt.Errorf("failed to upload content to S3: %v", err)
    }

    return nil
}