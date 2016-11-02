package utils

import (
	"net/smtp"
	"time"

	log "github.com/Sirupsen/logrus"
)

type EmailConfig interface {
	GetServiceEmail() string
	GetEmailPass() string
	GetSMTP() string
	GetSMTPPort() string
	GetAdminEmails() []string
}

type EmailUtils struct {
	config EmailConfig
}

func NewEmailUtils(config EmailConfig) *EmailUtils {
	return &EmailUtils{config}
}

func (u *EmailUtils) SendEmailOnce(toEmail string, subject string, body string) error {
	auth := smtp.PlainAuth("", u.config.GetServiceEmail(), u.config.GetEmailPass(), u.config.GetSMTP())
	to := []string{toEmail}
	msg := []byte(
		"To: " + toEmail + "\r\n" +
			"From: " + u.config.GetServiceEmail() + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" + body + "\r\n")
	err := smtp.SendMail(u.config.GetSMTP()+":"+u.config.GetSMTPPort(), auth,
		u.config.GetServiceEmail(), to, msg)
	if err != nil {
		log.Infof("Can't send email to %s: %v", toEmail, err)
	}

	return err
}

func (u *EmailUtils) SendEmail(toEmail string, subject string, body string) error {
	retries := 4
	var err error
	for retries > 0 {
		err = u.SendEmailOnce(toEmail, subject, body)
		if err == nil {
			break
		} else {
			retries--
			time.Sleep(time.Second * 10)
		}
	}
	if err != nil {
		log.Warnf("Can't send email to %s: %v", toEmail, err)
	}
	return err
}

func (u *EmailUtils) SendEmailToAdmins(subject, body string) (err error) {
	for _, email := range u.config.GetAdminEmails() {
		if err == nil {
			err = u.SendEmail(email, subject, body)
		}
	}
	return
}
