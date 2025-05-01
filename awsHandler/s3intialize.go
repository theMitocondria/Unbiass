package awsHandler

import(
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
)

var s3Client *s3.S3
func InitializeAWS() error {
	sess , err := session.NewSessionWithOptions(session.Options{
		Profile : "default",
		Config : aws.Config {
			Region : aws.String("eu-north-1"),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	s3Client = s3.New(sess)
	return  nil
}