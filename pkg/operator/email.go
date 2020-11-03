package operator

import (
	"fmt"
	"net/smtp"
)

func (o *Operator) SendEMail(mailTo []string, subject string, body string) error {
	from := o.Config.Email.SenderAddr
	// gmail special password
	specialPasswordStr := o.Config.Email.SpecialPassWord

	header := make(map[string]string)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + body
	auth := smtp.PlainAuth("", from, specialPasswordStr, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, mailTo, []byte(message))
	if err != nil {
		return err
	}
	fmt.Println("Send one email to ", mailTo)
	return nil
}
