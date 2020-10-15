package notify

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/spf13/viper"

	"transfer/database"
)

func SMS(content database.Task) (err error) {
	var regionID = viper.GetString("SMS.Ali.regionID")
	var accessKeyID = viper.GetString("SMS.Ali.accessKey")
	var accessKeySecret = viper.GetString("SMS.Ali.secretKey")

	var client *dysmsapi.Client
	if client, err = dysmsapi.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret); err != nil {
		return
	}

	var request = dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = "17010209162"
	request.SignName = "职历"
	request.TemplateCode = "SMS_163853223"
	request.TemplateParam = "{\"code\":\"200100\"}"

	response, err := client.SendSms(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
	return
}
