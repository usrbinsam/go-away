package gmail

import "github.com/usrbinsam/go-away/internal/config"

var InboxType config.InboxType = "gmail"

type GmailConfig struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
