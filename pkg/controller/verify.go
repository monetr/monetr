package controller

import (
	"fmt"
	"net/smtp"
)

func (c *Controller) sendEmailVerification(emailAddress string) (string, error) {
	conf := c.configuration.SMTP
	auth := smtp.PlainAuth(
		conf.Identity,
		conf.Username,
		conf.Password,
		conf.Host,
	)

	to := []string{emailAddress}
	from := "no-reply@mail.harderthanitneedstobe.com"
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: Please Verify Your Email Address\r\n"+
		"\r\n"+
		"This is the email body.\r\n",
		emailAddress,
		from,
	))
	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	if err := smtp.SendMail(address, auth, from, to, msg); err != nil {
		return "", err
	}

	return "", nil
}
