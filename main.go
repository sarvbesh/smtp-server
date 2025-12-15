package main

import (
	"fmt"
)

// structure for email request payload

type EmailRequest struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Recepients string `json:"recepients"`
}

// structure to store email configuration

type EmailConfig struct {
	senderEmail string
	password string
	smtpServer string
	smtpPort string
}

func main(){
	fmt.Println("hello")
}
