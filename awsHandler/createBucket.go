package awsHandler 

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"fmt"
)


func CreateBucket(bucketName string) error {
	_ , err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket : aws.String(bucketName),
	})
	
	if err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}
	return nil
}