package gmail

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/usrbinsam/go-away/internal/message"
	"github.com/usrbinsam/go-away/internal/store"
)

type GmailProvider struct {
	inboxConfig *store.InboxConfig
	httpClient  *http.Client
	oauthClient *oauthClient
	refreshing  bool
}

func New(store store.Store, inboxConfig *store.InboxConfig) *GmailProvider {
	var (
		clientID     = os.Getenv("GO_AWAY_GMAIL_CLIENT_ID")
		clientSecret = os.Getenv("GO_AWAY_GMAIL_CLIENT_SECRET")
	)

	if clientID == "" || clientSecret == "" {
		log.Fatalf("missing gmail oauth client credentials. ensure 'GO_AWAY_GMAIL_CLIENT_ID' and 'GO_AWAY_GMAIL_CLIENT_SECRET' are set")
	}

	provider := &GmailProvider{
		inboxConfig: inboxConfig,
		httpClient:  &http.Client{},
		oauthClient: &oauthClient{
			clientID, clientSecret,
		},
	}

	provider.Init()
	return provider
}

func (gmail *GmailProvider) Init() {
	if gmail.inboxConfig.IsSet("credentials::accessToken") && gmail.inboxConfig.IsSet("credentials::refreshToken") {
		log.Printf("gmail: using existing credentials")
		return
	}

	tokens := gmail.oauthClient.getCredentials()
	gmail.saveCredentials(tokens)
}

func (gmail *GmailProvider) saveCredentials(tokens *OAuthCredentials) {
	log.Printf("saving gmail access token: %s", tokens.AccessToken)
	gmail.inboxConfig.Set("credentials::accessToken", tokens.AccessToken)
	if tokens.RefreshToken != "" {
		gmail.inboxConfig.Set("credentials::refreshToken", tokens.RefreshToken)
		log.Printf("saving gmail refresh token: %s", tokens.RefreshToken)
	}
}

func (gmail *GmailProvider) GetMail() []*message.Message {
	req, err := http.NewRequest("GET", "https://gmail.googleapis.com/gmail/v1/users/me/messages", nil)
	if err != nil {
		log.Fatalf("gmail: error creating request: %s", err)
	}
	req.URL.RawQuery = "maxResults=3"

	log.Println("gmail: loading messages")
	res, err := gmail.do(req)
	if err != nil {
		log.Fatalf("gmail: error listing messages: %s\n", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("gmail: err reading response body while listing messages: %s", err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("gmail: non-200 status code listing messages: %q", body)
	}

	fmt.Printf("%s", string(body))

	var parsedBody GmailMessageListResponse
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		log.Fatalf("gmail: error parsing message list: %s\n", err)
	}

	messages := make([]*message.Message, len(parsedBody.Messages))

	for i, listItem := range parsedBody.Messages {
		gMessage := gmail.getMessage(listItem.Id)
		messages[i] = gMessage.ToMessage()
	}

	return messages
}

func (gmail *GmailProvider) getMessage(id string) *GmailMessage {
	req, err := http.NewRequest("GET", "https://gmail.googleapis.com/gmail/v1/users/me/messages/"+id, nil)
	if err != nil {
		log.Fatalf("gmail: unexpected err creating request: %s", err)
	}
	req.URL.RawQuery = "format=metadata"

	res, err := gmail.do(req)
	if err != nil {
		log.Fatalf("gmail: unexpected err retrieving message id %q: %s", id, err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("gmail: unexpected err reading message id body: %s", err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("gmail: HTTP %d retrieving message id %q: %s", res.StatusCode, id, string(body))
	}

	var parsedMessage GmailMessage
	err = json.Unmarshal(body, &parsedMessage)
	if err != nil {
		log.Fatalf("gmail: err parsing message id %q: %s", id, err)
	}
	fmt.Println(string(body))
	return &parsedMessage
}

func (gmail *GmailProvider) Send(to, subject, body string) error {
	return nil
}

func (gmail *GmailProvider) do(req *http.Request) (*http.Response, error) {
	accessToken := "Bearer " + gmail.inboxConfig.GetString("credentials::accessToken")
	log.Printf("access token: %s", accessToken)
	req.Header.Add("authorization", accessToken)

	res, err := gmail.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 {
		body, _ := io.ReadAll(res.Body)
		if gmail.refreshing {
			panic(fmt.Sprintf("gmail: got another 401 after refreshing the access token, %s", body))
		} else {
			log.Printf("gmail: 401, refreshing access token: %s", body)
		}

		tokens := gmail.oauthClient.refreshGmailCreds(gmail.inboxConfig.GetString("credentials::refreshToken"))
		gmail.saveCredentials(tokens)
		gmail.refreshing = true
		return gmail.do(req)
	}

	gmail.refreshing = false
	return res, err
}
