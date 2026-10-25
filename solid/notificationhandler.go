package main

import "fmt"

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) Send(customer string) {
	fmt.Printf("Уведомление по email отправлено клиенту %s\n", customer)
}

type SMSSender struct{}

func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

func (s *SMSSender) Send(customer string) {
	fmt.Printf("Уведомление по SMS отправлено клиенту %s\n", customer)
}
