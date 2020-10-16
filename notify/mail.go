package notify

import (
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"

	"transfer/database"
)

func Mail(content database.Task) (err error) {
	var msg = gomail.NewMessage()
	msg.SetHeader("From", viper.GetString("Email.from"))
	msg.SetHeader("To", viper.GetString("Email.to"))
	msg.SetHeader("Subject", viper.GetString("Email.subject"))
	msg.SetBody("text/html", content.Name)

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
