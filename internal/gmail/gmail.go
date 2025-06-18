package gmail

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GmailProvider struct {
	accessToken  string
	refreshToken string
	httpClient   *http.Client
}

type GmailMessageHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GmailMessagePartBody struct {
	AttachmentId string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         []byte `json:"data"`
}

type GmailMessagePart struct {
	PartId   string               `json:"partId"`
	MimeType string               `json:"mimeType"`
	Filename string               `json:"filename,omitempty"`
	Headers  []GmailMessageHeader `json:"headers"`
	Body     GmailMessagePartBody `json:"body"`
	Parts    []GmailMessagePart   `json:"parts"`
}

// GmailMessage is documented at https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages#GmailMessage
type GmailMessage struct {
	Id        string           `json:"id"`
	ThreadId  string           `json:"threadId"`
	Snippet   string           `json:"snippet,omitempty"`
	Payload   GmailMessagePart `json:"payload"`
	Raw       string           `json:"raw,omitempty"`
	LabelIds  []string         `json:"labelIds,omitempty"`
	HistoryId string           `json:"historyId"`
}

type GmailMessageListItem struct {
	Id string `json:"id"`
	// ThreadId string `json:"threadId"`
}

type GmailMessageListResponse struct {
	Messages           []GmailMessageListItem `json:"messages"`
	NextPageToken      string                 `json:"nextPageToken"`
	ResultSizeEstimate int                    `json:"resultSizeEstimate"`
}

func NewGmailProvider(accessToken, refreshToken string) *GmailProvider {
	return &GmailProvider{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		httpClient:   &http.Client{},
	}
}

func (gmail *GmailProvider) baseRequest(method, url string, body io.Reader) *http.Request {
	if gmail.accessToken == "" {
		panic("missing access token")
	}

	url = "https://gmail.googleapis.com" + url
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	req.Header.Add("authorization", "Bearer "+gmail.accessToken)
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
