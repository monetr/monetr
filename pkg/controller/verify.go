package controller

import (
	"fmt"
	"net/smtp"
)

func (c *Controller) sendEmailVerification(emailAddress, code string) (string, error) {
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
		"From: No Reply\r\n"+
		"Subject: Please Verify Your Email Address\r\n"+
		"\r\n"+
		"Your verification code is: %s\r\n",
		emailAddress,
		code,
	))
	address := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	if err := smtp.SendMail(address, auth, from, to, msg); err != nil {
		return "", err
	}

	return "", nil
}
