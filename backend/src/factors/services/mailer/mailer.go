package mailer

import (
	log "github.com/sirupsen/logrus"
)

type Mailer interface {
	SendMail(to, from, subject, html, text string) error
}

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) SendMail(to, from, subject, html, text string) error {
	log.WithFields(log.Fields{
		"To":      to,
		"From":    from,
		"Subject": subject,
		"html":    html,
		"text":    text,
	}).Debug("Sending Email")
	return nil
}
