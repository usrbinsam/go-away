package config

type InboxType string

type Inbox struct {
	Type   InboxType      `json:"type"`
	Config map[string]any `json:"config"`
}

type AppConfig struct {
	Version int `json:"version"`
	Inboxes []Inbox
}

type ConfigProvider interface {
	Exists() bool
	Load() AppConfig
	Store(*AppConfig)
}
