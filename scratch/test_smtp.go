package main

import (
	"fmt"
	"log"
	"os"
	"super-app-chonburi-go/config"
	"super-app-chonburi-go/pkg/mail"
)

func main() {
	// Set the DSN just in case config.LoadConfig requires it and it's not set
	os.Setenv("DB_DSN", "postgres://uat_admin:C807BI%2B%2BwFizSBH%2FFkNgL7J3qW646npW@122.155.169.235:5489/uat_chonburi?sslmode=disable")
	
	cfg := config.LoadConfig()
	fmt.Printf("Loaded config:\n")
	fmt.Printf("SMTP_HOST: %q\n", cfg.SMTPHost)
	fmt.Printf("SMTP_PORT: %q\n", cfg.SMTPPort)
	fmt.Printf("SMTP_EMAIL: %q\n", cfg.SMTPEmail)
	fmt.Printf("SMTP_PASSWORD: %q (length: %d)\n", cfg.SMTPPassword, len(cfg.SMTPPassword))

	emailSender := mail.NewSMTPEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPEmail, cfg.SMTPPassword)
	
	to := []string{"phnjk2000@gmail.com"}
	subject := "Test SMTP Configuration from Chonburi Super App"
	body := "<h1>Test</h1><p>This is a test email to verify SMTP configuration.</p>"
	
	fmt.Printf("Sending email to %v...\n", to)
	err := emailSender.SendHTML(to, subject, body)
	if err != nil {
		log.Fatalf("FAILED to send email: %v", err)
	}
	fmt.Println("SUCCESS: Email sent successfully!")
}
