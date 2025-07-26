// Package scanner contains logic for scanning messages for an unsubscibe method
package scanner

import (
	"errors"
	"log"
	"net/url"
	"regexp"

	"github.com/usrbinsam/go-away/internal/message"
	"github.com/usrbinsam/go-away/internal/provider"
)

type UnsubscribeFunc func() error

type ScanResult struct {
	Hit         bool
	Unsubscribe UnsubscribeFunc
	Reason      string
}

type Scanner interface {
	Scan(*message.Message) (*ScanResult, error)
}

var reListUnsubsbscribe = regexp.MustCompile(`<([^>]+)>`)

type HeaderScanner struct {
	provider provider.Provider
}

func NewHeaderScanner(provider provider.Provider) *HeaderScanner {
	return &HeaderScanner{provider}
}

func (hs *HeaderScanner) Scan(message *message.Message) (*ScanResult, error) {
	for _, name := range []string{"list-unsubscribe", "list-unsubscribe-post"} {
		value := message.GetHeader(name)
		if value == "" {
			continue
		}

		to, subject, body, err := hs.getUnsubscribeTarget(value)
		if err != nil {
			return nil, err
		}

		unsubscribeFunc := func() error {
			return hs.provider.Send(to, subject, body)
		}

		return &ScanResult{true, unsubscribeFunc, "matched List-Unsubscribe header"}, nil
	}
	return &ScanResult{false, nil, "no matching List-Unsubscribe header"}, nil
}

func (hs *HeaderScanner) getUnsubscribeTarget(listUnsubscribe string) (to, subject, body string, err error) {
	matches := reListUnsubsbscribe.FindAllStringSubmatch(listUnsubscribe, -1)

	for _, match := range matches {
		if len(match) != 2 {
			log.Print("List-Unsubscribe header does not match expected format: " + listUnsubscribe)
			continue
		}

		to, subject, body, err = tryParse(match[1])
		if err != nil {
			log.Printf("error parsing List-Unsubscribe value '%s': %v", match[1], err)
			continue
		}

		return
	}
	err = errors.New("couldn't find a usable List-Unsubscribe. see logs for details")
	return
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
