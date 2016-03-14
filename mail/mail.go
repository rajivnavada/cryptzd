package mail

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
)

var serviceInited = false

func newGmailMailer(username, password string) *gomail.Dialer {
	d := gomail.NewDialer("smtp.gmail.com", 587, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d
}

type Mailer interface {
	Send(name, email, message string) bool
}

type mailer struct {
	username string
	password string
}

var M = mailer{
	username: "",
	password: "",
}

func (m mailer) Send(name, email, message string) bool {
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("Crypt Keeper <%s>", m.username))
	msg.SetHeader("To", fmt.Sprintf("%s <%s>", name, email))
	msg.SetHeader("Subject", "Please activate your public key")
	msg.SetBody("text/plain", message)

	mailer := newGmailMailer(m.username, m.password)
	if err := mailer.DialAndSend(msg); err != nil {
		println("----------------------------------------")
		println("Error when sending email")
		println(err.Error())
		println("----------------------------------------")
		return false
	}
	return true
}

func InitService(username, password string) {
	M.username = username
	M.password = password
	serviceInited = true
}

func Send(name, email, message string) bool {
	if !serviceInited {
		println("Trying to use service without initiating it")
		return false
	}
	return M.Send(name, email, message)
}
