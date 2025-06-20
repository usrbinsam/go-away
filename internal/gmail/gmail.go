package gmail

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	"github.com/usrbinsam/go-away/internal/config"
)

type GmailProvider struct {
	config     *GmailConfig
	httpClient *http.Client
}

func New(inboxConfig any) *GmailProvider {
	var gmailConfig GmailConfig
	err := mapstructure.Decode(inboxConfig, &gmailConfig)
	if err != nil {
		log.Fatalf("error decoding gmail config: %s", err)
	}

	return &GmailProvider{
		config:     &gmailConfig,
		httpClient: &http.Client{},
	}
}

func Init() (config.InboxType, GmailConfig) {
	tokens := getGmailCreds()
	inboxConfig := GmailConfig(tokens)
	return InboxType, inboxConfig
}

func (gmail *GmailProvider) baseRequest(method, url string, body io.Reader) *http.Request {
	if gmail.config.AccessToken == "" {
		panic("missing access token")
	}

	url = "https://gmail.googleapis.com" + url
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	req.Header.Add("authorization", "Bearer "+gmail.config.AccessToken)
	return req
}

func (gmail *GmailProvider) GetMail() {
	req := gmail.baseRequest("GET", "/gmail/v1/users/me/messages", nil)
	req.URL.RawQuery = "maxResults=3"

	res, err := gmail.httpClient.Do(req)
	if err != nil {
		log.Fatalf("gmail: error listing messages: %s\n", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("gmail: err reading response body while listing messages: %s\n", err)
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

	for _, listItem := range parsedBody.Messages {
		_ = gmail.getMessage(listItem.Id)
		// log.Printf("%v", msg)
	}
}

func (gmail *GmailProvider) getMessage(id string) *GmailMessage {
	req := gmail.baseRequest("GET", "/gmail/v1/users/me/messages/"+id, nil)
	req.URL.RawQuery = "format=metadata"

	res, err := gmail.httpClient.Do(req)
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
