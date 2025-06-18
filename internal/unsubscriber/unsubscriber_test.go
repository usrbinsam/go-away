package unsubscriber_test

import (
	"strings"
	"testing"

	"github.com/usrbinsam/go-away/internal/message"
	"github.com/usrbinsam/go-away/internal/unsubscriber"
)

type fakeMailer struct {
	callback func(fakeMessage)
}

type fakeMessage struct {
	to, subject, body string
}

func (f *fakeMailer) Send(to, subject, body string) error {
	f.callback(fakeMessage{to, subject, body})
	return nil
}

func TestRFC2369Unsubscriber_Unsubscribe(t *testing.T) {
	testCases := []struct {
		name            string
		testMessage     *message.Message
		expectedTo      string
		expectedSubject string
		expectedBody    string
	}{
		{
			name: "valid List-Unsubscribe",
			testMessage: message.NewMessage(
				[]message.Header{
					{
						Name:  "List-Unsubscribe",
						Value: "(\"Click here to unsubscribe \") <mailto:go-away@go-away.com?subject=unsubscribe&body=GO%20AWAY>",
					},
				},
				strings.NewReader("Click nowhere to unsubscribe."),
			),
			expectedTo:      "go-away@go-away.com",
			expectedSubject: "unsubscribe",
			expectedBody:    "GO AWAY",
		}, {
			name: "multi valid List-Unsubscribe",
			testMessage: message.NewMessage(
				[]message.Header{
					{
						Name:  "List-Unsubscribe",
						Value: "<http://unsubscribe.go-away.com/?email=foo>, <mailto:go-away@go-away.com?subject=unsubscribe&body=GO%20AWAY>",
					},
				},
				strings.NewReader("Click nowhere to unsubscribe."),
			),
			expectedTo:      "go-away@go-away.com",
			expectedSubject: "unsubscribe",
			expectedBody:    "GO AWAY",
		}, {
			name: "badly-formed List-Unsubscribe",
			testMessage: message.NewMessage(
				[]message.Header{
					{
						Name:  "List-Unsubscribe",
						Value: "mailto:go-away@go-away.com?subject=unsubscribe&body=GO%20AWAY",
					},
				},
				strings.NewReader("Click nowhere to unsubscribe."),
			),
			expectedTo:      "go-away@go-away.com",
			expectedSubject: "unsubscribe",
			expectedBody:    "GO AWAY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callback := func(msg fakeMessage) {
				if msg.to != tc.expectedTo {
					t.Errorf("expected to: %s, got: %s", tc.expectedTo, msg.to)
				}
				if msg.subject != tc.expectedSubject {
					t.Errorf("expected subject: %s, got: %s", tc.expectedSubject, msg.subject)
				}
				if msg.body != tc.expectedBody {
					t.Errorf("expected body: %s, got: %s", tc.expectedBody, msg.body)
				}
			}

			unsub := unsubscriber.NewRFC2369Unsubscriber(
				&fakeMailer{
					callback,
				},
			)

			err := unsub.Unsubscribe(tc.testMessage)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
