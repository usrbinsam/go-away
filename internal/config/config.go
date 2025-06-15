package config

type GmailConfig struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type AppConfig struct {
	GmailConfig GmailConfig `json:"gmail"`
}

type ConfigProvider interface {
	Exists() bool
	Load() AppConfig
	Store(*AppConfig)
}
