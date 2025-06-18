package message

import (
	"fmt"
)

type Header struct {
	Name  string
	Value string
}

type Message struct {
	headers []Header
	body    string
}

func NewMessage(headers []Header, body string) *Message { // XXX: rethink this
	return &Message{headers, body}
}

func (m *Message) GetHeader(name string) string {
	for _, header := range m.headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func (m *Message) Headers() []Header {
	return m.headers
}

func (m *Message) RFC822() *string {
	headers := ""

	for _, header := range m.headers {
		headers += fmt.Sprintf("%s: %s\r\n", header.Name, header.Value)
	}

	v := fmt.Sprintf("%s\r\n%s", headers, m.body)
	return &v
}
