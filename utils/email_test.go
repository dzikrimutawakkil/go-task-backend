package utils

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// TestSendEmail tests the email sending functionality.
// This test is skipped by default. Run with:
//
//	go test ./utils/... -v -run TestSendEmail
//
// To run this test, you need SMTP configured in your .env file:
//
//	SMTP_HOST=smtp.gmail.com
//	SMTP_PORT=587
//	SMTP_USER=your-email@gmail.com
//	SMTP_PASSWORD=your-app-password
//	SMTP_FROM=your-email@gmail.com
//
// Usage:
//
//	# Run with default recipient (test@example.com)
//	go test ./utils/... -v -run TestSendEmail
//
//	# Run with custom recipient
//	TEST_EMAIL_RECIPIENT=mutawakkilalhamdika@gmail.com go test ./utils/... -v -run TestSendEmail
func TestSendEmail(t *testing.T) {
	// Load .env for testing
	_ = godotenv.Load()

	config := GetEmailConfig()

	// Check if SMTP is configured
	if config.Username == "" || config.Password == "" {
		t.Skip("SMTP not configured. Set SMTP_USER and SMTP_PASSWORD in .env to run this test")
	}

	// Get recipient from environment or use default
	recipient := os.Getenv("TEST_EMAIL_RECIPIENT")
	if recipient == "" {
		recipient = "test@example.com"
	}

	t.Logf("SMTP Config: Host=%s, Port=587, Username=%s, From=%s",
		config.Host, config.Username, config.From)
	t.Logf("Sending test email to: %s", recipient)

	// Test email content
	subject := "✅ GoTask Backend - SMTP Test Email"
	body := `Hello!

This is a test email from GoTask Backend.

If you received this email, it means:
✅ SMTP is properly configured
✅ Email sending is working correctly

Best regards,
GoTask Backend System`

	// Send email
	err := SendEmail(recipient, subject, body)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	t.Logf("✅ Email sent successfully to %s", recipient)
}
