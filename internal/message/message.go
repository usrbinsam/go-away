package message

import "io"

type Header struct {
	Name  string
	Value string
}

type Message struct {
	headers []Header
	body    io.Reader
}

func NewMessage(headers []Header, body io.Reader) *Message { // XXX: rethink this
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
