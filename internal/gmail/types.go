package gmail

import (
	"strings"

	"github.com/usrbinsam/go-away/internal/message"
)

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

func (gmailMessage *GmailMessage) GetHeader(name string) string {
	for _, header := range gmailMessage.Payload.Headers {
		if strings.EqualFold(header.Name, name) {
			return header.Value
		}
	}
	return ""
}

func (gmailMessage *GmailMessage) Body() string {
	return string(gmailMessage.Payload.Body.Data)
}

func (gmailMessage *GmailMessage) ToMessage() *message.Message {
	// annoying conversion because slice invariance is impossible in Go
	headers := make([]message.Header, len(gmailMessage.Payload.Headers))
	for i, header := range gmailMessage.Payload.Headers {
		headers[i] = message.Header{Name: header.Name, Value: header.Value}
	}

	return message.NewMessage(headers, gmailMessage.Body())
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
