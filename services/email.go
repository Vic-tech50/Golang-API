package services

import (
	"fmt"
	"net/mail"
	"net/smtp" //import for email
	"os"
	"strings"
)

func SendEmail(to, subject, body string) error {
	// get crediential
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if host == "" || port == "" ||
		username == "" || password == "" || from == "" {
		return fmt.Errorf("email configuration is incomplete")
	}

	if strings.ContainsAny(to+from, "\r\n") ||
		strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid email header")
	}

	auth := smtp.PlainAuth("", username, password, host)

	message := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body

	err := smtp.SendMail(
		host+":"+port,
		auth,
		from,
		[]string{to},
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func SendHTMLEmail(to, subject, htmlBody string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if host == "" || port == "" ||
		username == "" || password == "" || from == "" {
		return fmt.Errorf("email configuration is incomplete")
	}

	recipient, err := mail.ParseAddress(to)
	if err != nil || recipient.Address != to {
		return fmt.Errorf("invalid recipient email")
	}

	sender, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("invalid sender email configuration")
	}

	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid email subject")
	}

	message := "From: " + sender.String() + "\r\n" +
		"To: " + recipient.String() + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlBody

	auth := smtp.PlainAuth("", username, password, host)

	return smtp.SendMail(
		host+":"+port,
		auth,
		sender.Address,
		[]string{recipient.Address},
		[]byte(message),
	)
}

/*

$env:SMTP_HOST = "sandbox.smtp.mailtrap.io"
$env:SMTP_PORT = "2525"
$env:SMTP_USERNAME = "6ccfd11bd194ce"
$env:SMTP_PASSWORD = "576b4acdd4d8d3"
$env:SMTP_FROM = "hello@example.com"

*/
