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

type MessageHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MessagePartBody struct {
	AttachmentId string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         []byte `json:"data"`
}

type MessagePart struct {
	PartId   string          `json:"partId"`
	MimeType string          `json:"mimeType"`
	Filename string          `json:"filename,omitempty"`
	Headers  []MessageHeader `json:"headers"`
	Body     MessagePartBody `json:"body"`
	Parts    []MessagePart   `json:"parts"`
}

// Message is documented at https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages#Message
type Message struct {
	Id        string      `json:"id"`
	ThreadId  string      `json:"threadId"`
	Snippet   string      `json:"snippet,omitempty"`
	Payload   MessagePart `json:"payload"`
	Raw       string      `json:"raw,omitempty"`
	LabelIds  []string    `json:"labelIds,omitempty"`
	HistoryId string      `json:"historyId"`
}

type MessageListItem struct {
	Id string `json:"id"`
	// ThreadId string `json:"threadId"`
}

type MessageListResponse struct {
	Messages           []MessageListItem `json:"messages"`
	NextPageToken      string            `json:"nextPageToken"`
	ResultSizeEstimate int               `json:"resultSizeEstimate"`
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

	var parsedBody MessageListResponse
	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		log.Fatalf("gmail: error parsing message list: %s\n", err)
	}

	for _, listItem := range parsedBody.Messages {
		_ = gmail.getMessage(listItem.Id)
		// log.Printf("%v", msg)
	}
}

func (gmail *GmailProvider) getMessage(id string) *Message {
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

	var parsedMessage Message
	err = json.Unmarshal(body, &parsedMessage)
	if err != nil {
		log.Fatalf("gmail: err parsing message id %q: %s", id, err)
	}
	fmt.Println(string(body))
	return &parsedMessage
}
