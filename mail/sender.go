package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

const (
	smtpAuthAddress   = "smtp.qq.com"
	smtpServerAddress = "smtp.qq.com:587"
)

// Sender Interface provides all method of sender
type Sender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string, //抄送
		bcc []string, //秘密抄送
		attachFiles []string, //添加的文件
	) error
}

// OutLookSender is a real implement of sender interface
type OutLookSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewOutLookSender(name string, fromEmailAddress string, fromEmailPassword string) Sender {
	return &OutLookSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *OutLookSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string, //抄送
	bcc []string, //秘密抄送
	attachFiles []string, //添加的文件
) error {
	//create a newEmail
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc
	for _, file := range attachFiles {
		_, err := e.AttachFile(file)
		if err != nil {
			return fmt.Errorf("failed to attch file:%s,%w", file, err)
		}
	}
	//create a smtpAuth object for newEmail to send email
	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}
