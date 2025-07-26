package main

import (
	"log"
	"strings"

	"github.com/usrbinsam/go-away/internal/gmail"
	"github.com/usrbinsam/go-away/internal/provider"
	"github.com/usrbinsam/go-away/internal/scanner"
	"github.com/usrbinsam/go-away/internal/store"
)

type Unsubscriber struct {
	providers   []provider.Provider
	scanners    []scanner.Scanner
	safeSenders []string
}

func (u *Unsubscriber) isSafeSender(addr string) bool {
	for _, safeSender := range u.safeSenders {
		if strings.Contains(addr, safeSender) {
			return true
		}
	}
	return false
}

func goAway(unsubscriber *Unsubscriber) {
	results := make([]*scanner.ScanResult, 1)
	scanned := 0
	for _, provider := range unsubscriber.providers {
		for _, msg := range provider.GetMail() {
			sender := msg.GetHeader("From")
			if unsubscriber.isSafeSender(sender) {
				continue
			}
			scanned++

			headerScanner := scanner.NewHeaderScanner(provider)
			result, err := headerScanner.Scan(msg)
			if err != nil {
				log.Printf("error scanning message: %s", err)
				continue
			}

			if !result.Hit {
				continue
			}
			results = append(results, result)
		}
	}

	log.Printf("scanned %d messages", scanned)
	log.Printf("scanners found %d messages to unsubscribe", len(results))
}

func main() {
	st := &store.SQLStore{}
	st.Open("go-away.sqlite3")

	inboxes := st.ListInboxes()
	safeSenders := []string{}

	unsubscriber := &Unsubscriber{
		providers:   make([]provider.Provider, len(inboxes)),
		safeSenders: safeSenders,
	}

	if len(inboxes) == 0 {
		log.Fatalf("no inboxes found")
	}

	for inboxIdx, inbox := range inboxes {
		inboxConfig := store.NewInboxConfig(inbox.ID, st)

		if inbox.Provider == "gmail" {
			provider := gmail.New(st, inboxConfig)
			unsubscriber.providers[inboxIdx] = provider
			goAway(unsubscriber)

		} else {
			log.Fatalf("unknown inbox type: %s", inbox.Provider)
		}

	}
}
