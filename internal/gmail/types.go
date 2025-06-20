package gmail

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
