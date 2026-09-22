package mail

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v4"
)

// ResendMailer sends emails via the Resend API (https://resend.com).
type ResendMailer struct {
	client *resend.Client
	from   string
}

// NewResendMailer creates a Mailer backed by the Resend API.
// Returns noopMailer if apiKey or fromAddress is empty.
func NewResendMailer(apiKey, fromAddress string) Mailer {
	if apiKey == "" || fromAddress == "" {
		return noopMailer{}
	}

	return &ResendMailer{
		client: resend.NewClient(apiKey),
		from:   fromAddress,
	}
}

func (m *ResendMailer) SendEmailVerificationOTP(_ context.Context, recipientEmail string, userName string, otp string, expiresInMinutes int) error {
	htmlBody, err := renderVerificationOTPEmail(verificationOTPEmailData{
		UserName:         userName,
		OTP:              otp,
		ExpiresInMinutes: expiresInMinutes,
	})
	if err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{recipientEmail},
		Subject: "Verify your email address",
		Html:    htmlBody,
	}

	_, err = m.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send email: %w", err)
	}

	return nil
}
