package awsHandler

import(
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"fmt"
)

func DownloadS3Object(bucketName , key string) ([]byte, error) {

    // Use GetObject to fetch the content directly into memory
    result, err := s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to download object: %v", err)
    }
    defer result.Body.Close()

    // Read all the data from the response body into a byte slice
    data, err := io.ReadAll(result.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read object content: %v", err)
    }

    return data, nil
}