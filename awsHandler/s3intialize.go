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
		Config : aws.Config {
			Region : aws.String("eu-north-1"),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	/ Debug: Check if credentials are loaded
	creds := sess.Config.Credentials
	val, err := creds.Get()
	if err != nil {
		return fmt.Errorf("unable to load AWS credentials: %v", err)
	}
	fmt.Printf("✅ AWS credentials loaded: %s\n", val.AccessKeyID)


	s3Client = s3.New(sess)
	return  nil
}