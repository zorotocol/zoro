package oracle

import (
	"context"
	"fmt"
	gomail "gopkg.in/gomail.v2"
	"net/url"
	"strconv"
)

func SMTPSender(config *url.URL) Sender {
	host := config.Hostname()
	port, err := strconv.Atoi(config.Port())
	if err != nil {
		panic("invalid port")
	}
	user := config.User.Username()
	from := config.Query().Get("from")
	pass, hasPass := config.User.Password()
	if !hasPass {
		panic("no password in smtp uri")
	}
	return func(ctx context.Context, mail *Mail) error {
		msg := gomail.NewMessage()
		msg.SetHeader("From", from)
		msg.SetHeader("To", mail.Email)
		msg.SetHeader("Subject", "Zorotocol Config")
		msg.SetBody("text/plain", fmt.Sprintf("password: %s\ndeadline: %s\ntransaction: (%d) %s", mail.Key, mail.Deadline, mail.LogIndex, mail.Tx))
		dialer := gomail.NewDialer(host, port, user, pass)
		return dialer.DialAndSend(msg)
	}
}
