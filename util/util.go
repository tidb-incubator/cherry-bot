package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const LimitIgnoreLabels = "status/DNM,status/WIP,S: DNM,S: WIP"

// RetryOnError defines a action with retry when "fn" returns error,
// we can specify the number and interval of retries
// code snippet from https://github.com/pingcap/schrodinger
func RetryOnError(ctx context.Context, retryCount int, fn func() error) error {
	var err error
	for i := 0; i < retryCount; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		err = fn()
		if err == nil {
			break
		}

		Error(err)
		Sleep(ctx, 2*time.Second)
	}

	return errors.Wrap(err, "retry error")
}

// Sleep defines special `sleep` with context
func Sleep(ctx context.Context, sleepTime time.Duration) {
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		return
	}
}

func SendEMail(mailTo []string, subject string, body string) error {
	// TODO read password.txt
	b, err := ioutil.ReadFile("/root/github-bot/gmail.txt")
	if err != nil {
		err = errors.Wrap(err, "read gmail file failed")
		return err
	}
	strs := strings.Split(string(b), ",")
	if len(strs) < 2 {
		return errors.New("gmail.txt is invalid")
	}
	from := strs[0]
	// gmail special password
	specialPasswordStr := strs[1]

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
	err = smtp.SendMail("smtp.gmail.com:587", auth, from, mailTo, []byte(message))
	if err != nil {
		return err
	}
	fmt.Println("Send one email to ", mailTo)
	return nil
}
