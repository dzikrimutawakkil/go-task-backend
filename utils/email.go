package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailConfig holds the SMTP configuration.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// GetEmailConfig loads email configuration from environment variables.
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		Host:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		Port:     587,
		Username: getEnvOrDefault("SMTP_USER", ""),
		Password: getEnvOrDefault("SMTP_PASSWORD", ""),
		From:     getEnvOrDefault("SMTP_FROM", getEnvOrDefault("SMTP_USER", "")),
	}
}

// getEnvOrDefault is a helper to get env var or return a default.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// SendEmail sends an email via SMTP.
// It returns an error if the email cannot be sent.
func SendEmail(to, subject, body string) error {
	config := GetEmailConfig()

	if config.Username == "" || config.Password == "" {
		GetLogger().Warn("SMTP not configured, skipping email", "to", to, "subject", subject)
		return nil // Skip if SMTP not configured (dev mode)
	}

	// Build email message
	msg := buildEmailMessage(config.From, to, subject, body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Setup TLS config
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}

	// Start TLS connection
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP: %w", err)
	}
	defer func() {
		// Perbaikan: abaikan error client.Close() di defer secara eksplisit
		_ = client.Close()
	}()

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err := client.Mail(config.Username); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		// Tutup writer jika gagal write, tapi jangan pedulikan error tutupnya karena sudah error
		_ = writer.Close() 
		return fmt.Errorf("failed to write email body: %w", err)
	}

	// Perbaikan: writer.Close() HARUS dicek karena ini yang memicu pengiriman aktual (commit)
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	GetLogger().Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}

// buildEmailMessage builds a standard email message.
func buildEmailMessage(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "quoted-printable"

	var sb strings.Builder
	for key, value := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	sb.WriteString("\r\n")
	sb.WriteString(body)

	return sb.String()
}

// SendInviteEmail sends an organization invitation email.
func SendInviteEmail(to, orgName, inviteURL string) error {
	subject := fmt.Sprintf("You're invited to join %s", orgName)
	body := fmt.Sprintf(
		`Hello!

You have been invited to join the organization "%s" on our platform.

Click the link below to accept the invitation:
%s

This link will expire in 7 days.

Best regards,
The Team`,
		orgName,
		inviteURL,
	)

	return SendEmail(to, subject, body)
}