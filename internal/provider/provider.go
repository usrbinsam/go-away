package provider

import "github.com/usrbinsam/go-away/internal/message"

// A Provider defines the interface for an inbox provider (i.e., gmail, generic IMAP, etc.)
type Provider interface {
	GetMail() []*message.Message
	Send(to, subject, body string) error
}
