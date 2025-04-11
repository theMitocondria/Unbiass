package awsHandler

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
)

func CreateSESEmailTemplate(templateName string , subjectPart string , textPart string , htmlPart string) error {
	// Define the template
	template := &ses.CreateTemplateInput{
		Template: &ses.Template{
			TemplateName: aws.String(templateName),
			SubjectPart:  aws.String(subjectPart),
			TextPart:     aws.String(textPart),
			HtmlPart:     aws.String(htmlPart),
		},
	}

	// Create the template
	_, err := svc.CreateTemplate(template)
	if err != nil {
		return fmt.Errorf("error creating template: %v", err)
	}

	fmt.Println("Email template created successfully!")
	return nil
}
