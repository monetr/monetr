//go:build mini

package mail

import (
	"errors"
	"net/smtp"
)

// Sending emails via SMTP requires props TLS. When we are running things locally though in minikube I just want to be
// able to use mailhog to send emails. It's not an insignificant amount of work to give mailhog proper TLS within the
// minikube cluster so instead I have implemented an smtp.Auth object that will ignore TLS when we are running inside
// minikube.

var (
	_ smtp.Auth = &smtpMinikubeAuthentication{}
)

type smtpMinikubeAuthentication struct {
	identity, username, password string
	host                         string
}

func PlainAuth(identity, username, password, host string) smtp.Auth {
	return &smtpMinikubeAuthentication{identity, username, password, host}
}

func (a *smtpMinikubeAuthentication) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *smtpMinikubeAuthentication) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
