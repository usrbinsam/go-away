package unsubscriber

import (
	"errors"
	"log"
	"net/url"
	"regexp"

	"github.com/usrbinsam/go-away/internal/mailer"
	"github.com/usrbinsam/go-away/internal/message"
)

var reListUnsubsbscribe = regexp.MustCompile(`<([^>]+)>`)

// Unsubscriber defines an interface that attempts to unsubscribe from a mailing list.
// Mailing lists may implement different methods of unsubscription, such as RFC 2369 or custom methods.
type Unsubscriber interface {
	Unsubscribe(*message.Message) error
}

// RFC2369Unsubscriber implements the Unsubscriber interface for RFC 2369 compliant unsubscription.
type RFC2369Unsubscriber struct {
	mailer mailer.Mailer
}

func NewRFC2369Unsubscriber(mailer mailer.Mailer) *RFC2369Unsubscriber {
	return &RFC2369Unsubscriber{mailer}
}

// Unsubscribe attempts to unsubscribe from a mailing list using the RFC 2369 method.
// Expected to be called with a message that contains the necessary headers for unsubscription.
func (r *RFC2369Unsubscriber) Unsubscribe(msg *message.Message) error {
	listUnsubscribe := msg.GetHeader("List-Unsubscribe")

	if listUnsubscribe == "" {
		return errors.New("no List-Unsubscribe header found")
	}

	unsubscribed := false
	matches := reListUnsubsbscribe.FindAllStringSubmatch(listUnsubscribe, -1)

	for _, match := range matches {
		if len(match) != 2 {
			log.Print("List-Unsubscribe header does not match expected format: " + listUnsubscribe)
		}

		to, subject, body, err := tryParse(match[1])
		if err != nil {
			log.Printf("error parsing List-Unsubscribe value '%s': %v", match[1], err)
			continue
		}

		err = r.mailer.Send(to, subject, body)
		if err != nil {
			log.Panic("error sending unsubscription request")
		}

		unsubscribed = true
	}

	if unsubscribed {
		return nil
	}

	return errors.New("couldn't find a usable List-Unsubscribe. see logs for details")
}

func tryParse(listUnsubscribeValue string) (to, subject, body string, err error) {
	u, err := url.Parse(listUnsubscribeValue)
	if err != nil {
		return
	}

	if u.Scheme != "mailto" {
		err = errors.New("only mailto scheme is supported for List-Unsubscribe")
		return
	}

	params := u.Query()
	to = u.Opaque
	body = params.Get("body")
	subject = params.Get("subject")

	if to == "" {
		err = errors.New("expected opaque URI data")
		return
	}

	if body == "" {
		body = "Please unsubscribe me from this mailing list."
	}

	if subject == "" {
		subject = "Unsubscribe Request"
	}

	return
}
