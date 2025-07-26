package gmail

var GmailInboxKey = "gmail"

type GmailInboxConfig struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type GmailOAuthConfig struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type GmailAppSettings struct {
	GmailOAuthConfig
}
