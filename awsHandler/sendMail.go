package awsHandler

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ses"
    "errors"
)

// EmailRecipient represents a single recipient and their template data
type EmailRecipient struct {
    Email        string
    TemplateData map[string]interface{}
}

// BulkEmailSender handles sending bulk emails
type BulkEmailSender struct {
    svc          *ses.SES
    templateName string
    sender       string
    maxWorkers   int
    rateLimit    time.Duration // Time to wait between sends to avoid throttling
}

// NewBulkEmailSender creates a new BulkEmailSender instance
func NewBulkEmailSender( templateName, sender string) (*BulkEmailSender, error) {
    return &BulkEmailSender{
        svc:          svc,
        templateName: templateName,
        sender:       sender,
        maxWorkers:   3,                  // Concurrent email sends
        rateLimit:    time.Millisecond * 100, // 4 emails per second (adjust based on your SES limits)
    }, nil
}

// SendBulkEmails sends emails to multiple recipients
func (s *BulkEmailSender) SendBulkEmails(recipients []EmailRecipient) []error {
    var wg sync.WaitGroup
    errorsChan := make(chan error, len(recipients))
    semaphore := make(chan struct{}, s.maxWorkers)
    
    for _, recipient := range recipients {
        wg.Add(1)
        go func(r EmailRecipient) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() {
                // Release semaphore
                <-semaphore
                // Rate limiting
                time.Sleep(s.rateLimit)
            }()

            if err := s.SendSingleEmail(r); err != nil {
                errorsChan <- fmt.Errorf("failed to send email to %s: %v", r.Email, err)
            }
        }(recipient)
    }

    // Wait for all goroutines to complete
    wg.Wait()
    close(errorsChan)

    // Collect errors
    var errors []error
    for err := range errorsChan {
        errors = append(errors, err)
    }

    return errors
}

// sendSingleEmail sends an email to a single recipient
func (s *BulkEmailSender) SendSingleEmail(recipient EmailRecipient) error {
    // Convert template data to JSON
    templateData, err := json.Marshal(recipient.TemplateData)
    if err != nil {
        return fmt.Errorf("failed to marshal template data: %v", err)
    }

    input := &ses.SendTemplatedEmailInput{
        Destination: &ses.Destination{
            ToAddresses: []*string{aws.String(recipient.Email)},
        },
        Source:       aws.String(s.sender),
        Template:     aws.String(s.templateName),
        TemplateData: aws.String(string(templateData)),
    }

    _, err = s.svc.SendTemplatedEmail(input)
    return err
}

func SendBulk(templateName string , senderEmail string , recipients []EmailRecipient)[]error {
    // Initialize the bulk email sender
    sender, err := NewBulkEmailSender(
        templateName,                // Template name
        senderEmail,               // Sender email
    )
    if err != nil {
        return []error{errors.New(fmt.Sprintf("Failed to create bulk email sender: %v\n", err))}
    }

    // Send bulk emails
    if errors := sender.SendBulkEmails(recipients); len(errors) > 0 {
        return errors 
    } 

    return nil
}
