package main

import (
	"encoding/json" // for request decoding
	"fmt"
	"log"
	"net/http" // for email sending
	"net/smtp" // for server
	"os"
	"regexp"
	"strings"
	"time"
)

// structure for email request payload

type EmailRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Recipients []string `json:"recipients"`
}

// handles incomind HTTP request to send an email

func sendEmailHandler(wr http.ResponseWriter, rd *http.Request) {

	// restrict to POST method only
	
	if rd.Method != http.MethodPost {
		http.Error(wr, "Only POST Method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	// decode the request payload

	var request EmailRequest

	if err := json.NewDecoder(rd.Body).Decode(&request); err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	// validate recipient email address
	
	for _, recipient := range request.Recipients {
		if !isValidEmail(recipient) {
			http.Error(wr, fmt.Sprintf("Recipient email address '%s' is not valid", recipient), http.StatusBadRequest)
			return
		}
	}

	EmailConfig, err := getEmailConfig()

	if err != nil {
		log.Fatal(err)
		
	}

	// authenticate with SMTP server

	auth := smtp.PlainAuth("", EmailConfig.senderEmail, EmailConfig.password, EmailConfig.smtpServer)

	// format SMTP server

	addr := fmt.Sprintf("%s:%s", &EmailConfig.smtpServer, &EmailConfig.smtpPort)

	// body in MIME format

	msg := formatEmailMessage(request.Recipients, request.Subject, request.Message)

	maxRetries := 3
	retryCount := 0
	backoff := 1 * time.Second

	for {
		if err := smtp.SendMail(addr, auth, EmailConfig.senderEmail, request.Recipients, msg); err != nil {
			
			retryCount++

			if retryCount >= maxRetries { // max entries reached, return a response or an error
				
				http.Error(wr, "Failed to send an email after multiple attempts.", http.StatusInternalServerError)
				
				return
			}

			log.Printf("Attempt %d failed, retrying in %v...\n", retryCount, backoff)

			time.Sleep(backoff)

			backoff *= 2

		} else {
			
			break
		}
	}

	wr.Header().Set("Content-Type", "text/plain")
	wr.WriteHeader(http.StatusOK)
	wr.Write([]byte("Email has been sent successfully!"))
}

// structure to store email configuration

type EmailConfig struct {
	
	senderEmail string
	password string
	smtpServer string
	smtpPort string
}

// get email configuration from environment variables

func getEmailConfig() (EmailConfig, error) {
	
	config := EmailConfig{
		senderEmail: os.Getenv("SENDER_EMAIL"),
		password: os.Getenv("EMAIL_PASSWORD"),
		smtpServer: os.Getenv("SMTP_SERVER"),
		smtpPort: os.Getenv("SMTP_PORT"),
	}

	if config.senderEmail == "" || config.password == "" || config.smtpServer == "" || config.smtpPort == "" {
		return EmailConfig{}, fmt.Errorf("One or more environment variables are not set.")
	}

	if !isValidEmail(config.senderEmail) {
		return EmailConfig{}, fmt.Errorf("Sender email address is not valid.")
	}

	return config, nil
}

// check if provided email address is valid

func isValidEmail(email string) bool {
	
	const emailRegexPattern = `^([A-Z0-9_+-]+\.?)*[A-Z0-9_+-]@([A-Z0-9_+-][A-Z0-9_+-]*\.)+[A-Z]{2,}$/i`

	matched, err := regexp.MatchString(emailRegexPattern, email)

	if err != nil {
		return false
	}

	return matched                           
}

// format email message

func formatEmailMessage(recipients []string, subject, message string) []byte {

	return []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", strings.Join(recipients, ","), subject, message))
}


func main(){

	http.HandleFunc("/send-email", sendEmailHandler)

	log.Println("Server starting on Port 8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server start error: %s", err)
	}
}