package awsHandler

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"fmt"
)

var svc *ses.SES
func InitializeSESService() error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"), // Replace with your region
	})
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}
	svc = ses.New(sess)
	return nil
}
