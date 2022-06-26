package main

import (
	"log"

	"github.com/roshanlc/soe-backend/internal/data"
	mail "github.com/wneessen/go-mail"
)

type MailingContent struct {
	from    string
	to      string
	subject string
	content string
}
type MailingContainer struct {
	client *mail.Client
}

func (m *MailingContainer) Authenticate(config *data.Config) error {

	server, err := mail.NewClient(config.Mail.Host,
		mail.WithPort(config.Mail.Port),
		mail.WithUsername(config.Mail.Username),
		mail.WithPassword(config.Mail.Password), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(mail.TLSMandatory), mail.WithTimeout(mail.DefaultTimeout))

	if err != nil {
		log.Println(err)
		return err

	}

	// set the smtp client
	m.client = server
	return nil

}

func (m *MailingContainer) SendMail(obj *MailingContent) {

	email := mail.NewMsg()

	email.From(obj.from)
	email.To(obj.to)
	email.Subject(obj.subject)
	email.SetBodyString(mail.TypeTextPlain, obj.content)

	err := m.client.DialAndSend(email)
	// log the error
	if err != nil {
		log.Println(err)
	}

}

func NewMailer() *MailingContainer {
	return &MailingContainer{}
}
