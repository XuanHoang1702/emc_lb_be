package mail

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

	appconfig "emc_lb/src/pkg/config"
	"emc_lb/src/pkg/utils"
)

//go:embed templates/*.html
var templateFS embed.FS

type Mailer interface {
	SendEmailVerificationOTP(context.Context, string, string, string, int) error
}

type SMTPMailer struct {
	address   string
	auth      smtp.Auth
	fromEmail string
	fromName  string
	enabled   bool
}

type noopMailer struct{}

type verificationOTPEmailData struct {
	UserName         string
	OTP              string
	ExpiresInMinutes int
}

func NewSMTPMailer() Mailer {
	// Prefer typed config; fall back to env vars if config not yet loaded
	host := utils.GetEnv("SMTP_HOST", "")
	port := utils.GetEnv("SMTP_PORT", "")
	username := utils.GetEnv("SMTP_USERNAME", "")
	password := utils.GetEnv("SMTP_PASSWORD", "")
	fromEmail := utils.GetEnv("SMTP_FROM_EMAIL", "")
	fromName := utils.GetEnv("SMTP_FROM_NAME", "EMC LB")

	if cfg, err := appconfig.Load(); err == nil {
		host = cfg.Mail.Host
		port = fmt.Sprintf("%d", cfg.Mail.Port)
		username = cfg.Mail.Username
		password = cfg.Mail.Password
		fromEmail = cfg.Mail.FromEmail
		fromName = cfg.Mail.FromName
	}

	if host == "" || port == "" || username == "" || password == "" || fromEmail == "" {
		return noopMailer{}
	}

	return &SMTPMailer{
		address:   fmt.Sprintf("%s:%s", host, port),
		auth:      smtp.PlainAuth("", username, password, host),
		fromEmail: fromEmail,
		fromName:  fromName,
		enabled:   true,
	}
}

func (noopMailer) SendEmailVerificationOTP(context.Context, string, string, string, int) error {
	return nil
}

func (m *SMTPMailer) SendEmailVerificationOTP(_ context.Context, recipientEmail string, userName string, otp string, expiresInMinutes int) error {
	if m == nil || !m.enabled {
		return nil
	}

	htmlBody, err := renderVerificationOTPEmail(verificationOTPEmailData{
		UserName:         userName,
		OTP:              otp,
		ExpiresInMinutes: expiresInMinutes,
	})
	if err != nil {
		return err
	}

	message := buildHTMLMessage(m.fromName, m.fromEmail, recipientEmail, "Verify your email address", htmlBody)
	return smtp.SendMail(m.address, m.auth, m.fromEmail, []string{recipientEmail}, []byte(message))
}

func renderVerificationOTPEmail(data verificationOTPEmailData) (string, error) {
	tpl, err := template.ParseFS(templateFS, "templates/email_verification_otp.html")
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func buildHTMLMessage(fromName string, fromEmail string, toEmail string, subject string, htmlBody string) string {
	return fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		fromName,
		fromEmail,
		toEmail,
		subject,
		htmlBody,
	)
}
