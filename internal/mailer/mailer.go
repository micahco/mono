package mailer

import (
	"bytes"
	"context"
	"net/mail"

	"github.com/a-h/templ"
	"github.com/k3a/html2text"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
	sender *mail.Address
}

// Create new mailer with SMTP credentials
func New(host string, port int, username string, password string, sender *mail.Address) (*Mailer, error) {
	m := &Mailer{
		dialer: gomail.NewDialer(host, port, username, password),
		sender: sender,
	}

	// Ping the SMTP server to verify authentication
	s, err := m.dialer.Dial()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	return m, nil
}

func (m *Mailer) Send(recepient, subject string, component templ.Component) error {
	ctx := context.Background()

	var buf bytes.Buffer
	err := component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("To", recepient)
	msg.SetHeader("From", m.sender.String())
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", html2text.HTML2Text(buf.String()))
	msg.AddAlternative("text/html", buf.String())

	return m.dialer.DialAndSend(msg)
}
