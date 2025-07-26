package config

import "github.com/usrbinsam/go-away/internal/gmail"

type Inbox struct {
	InboxKey string         `json:"type"`
	Config   map[string]any `json:"config"`
}

type AppConfig struct {
	Version       int `json:"version"`
	Inboxes       []Inbox
	GmailSettings gmail.GmailAppSettings
}

type ConfigProvider interface {
	Exists() bool
	Load() AppConfig
	Store(*AppConfig)
}
