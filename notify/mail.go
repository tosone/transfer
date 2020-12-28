package notify

import (
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"

	"transfer/database"
)

type EmailType string

const (
	DownloadSuccess EmailType = "Download File Success"
)

func Mail(task database.Task, subject EmailType) (err error) {
	var msg = gomail.NewMessage()
	msg.SetHeader("From", viper.GetString("Email.from"))
	msg.SetHeader("To", viper.GetString("Email.to"))
	msg.SetHeader("Subject", string(subject))
	switch subject {
	case DownloadSuccess:
		msg.SetBody("text/html", fmt.Sprintf(emailDownloadSuccessTmpl, string(subject), task.URL, task.URL))
	default:
		err = fmt.Errorf("not support subject type")
		return
	}

	var mail = gomail.NewDialer(
		viper.GetString("Email.host"),
		viper.GetInt("Email.port"),
		viper.GetString("Email.username"),
		viper.GetString("Email.password"),
	)
	if err = mail.DialAndSend(msg); err != nil {
		return
	}
	return
}
