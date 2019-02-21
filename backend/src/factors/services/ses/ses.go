package ses

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type SESDriver struct {
	ses *ses.SES
}

const (
	charSet = "UTF-8"
)

func New(accesKeyId, secretAccessKey, region string) *SESDriver {
	session := session.New()
	credentials := credentials.NewStaticCredentials(accesKeyId, secretAccessKey, "")
	ses := ses.New(session, aws.NewConfig().WithCredentials(credentials).WithRegion(region))
	return &SESDriver{
		ses: ses,
	}
}

func (sesD *SESDriver) SendMail(to, from, subject, htmlBody, textBody string) error {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(from),
	}

	if err := input.Validate(); err != nil {
		return err
	}

	_, err := sesD.ses.SendEmail(input)

	return err
}
