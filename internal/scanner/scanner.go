package scanner

import (
	"strings"

	"github.com/usrbinsam/go-away/internal/message"
	"github.com/usrbinsam/go-away/internal/unsubscriber"
)

type Scanner interface {
	Scan(*message.Message) (bool, unsubscriber.Unsubscriber)
}

type HeaderScanner struct {
	SendFunc      func(string, string, string) error
	unsubscribers []unsubscriber.Unsubscriber
}

func (hs *HeaderScanner) Scan(message *message.Message) bool {
	for _, header := range message.Headers() {
		name := strings.ToLower(header.Name)
		if name == "list-unsubscribe" || name == "list-unsubscribe-post" {
			return true
		}
	}
	return false
}
