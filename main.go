package main

import (
	"log"

	"github.com/usrbinsam/go-away/internal/config"
	"github.com/usrbinsam/go-away/internal/gmail"
	"github.com/usrbinsam/go-away/internal/provider"
)

func main() {
	configProvider := config.FileConfig("$HOME/.go-away.json") // TODO: handle Windows
	appConfig := configProvider.Load()

	provider := make([]provider.Provider, len(appConfig.Inboxes))

	for i, inbox := range appConfig.Inboxes {
		if inbox.Type == gmail.InboxType {
			provider[i] = gmail.New(&inbox.Config)
		} else {
			log.Fatalf("unknown inbox type: %s\n", inbox.Type)
		}
	}
}
