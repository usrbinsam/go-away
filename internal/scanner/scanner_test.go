package scanner_test

import (
	"strings"
	"testing"

	"github.com/usrbinsam/go-away/internal/message"
	"github.com/usrbinsam/go-away/internal/scanner"
)

type TestScannable struct {
	messageHeaders []message.Header
	messageBody    []byte
}

func (t *TestScannable) Headers() []message.Header {
	return t.messageHeaders
}

func (t *TestScannable) Body() []byte {
	return t.messageBody
}

func TestHeaderScanner_ScanMatch(t *testing.T) {
	v := message.NewMessage(
		[]message.Header{
			{Name: "From", Value: "foo@example.com"},
			{Name: "List-Unsubscribe", Value: "<mailto:abuse@fbi.gov>"},
		},
		strings.NewReader("Click nowhere to unsubscribe."),
	)

	scanner := &scanner.HeaderScanner{}
	if !scanner.Scan(v) {
		t.Errorf("expected message to be unsubscribe-able, but it was not")
	}
}

func TestHeaderScanner_ScanNoMatch(t *testing.T) {
	v := message.NewMessage(
		[]message.Header{
			{Name: "From", Value: "foo@bar.com"},
		},
		strings.NewReader("Click nowhere to unsubscribe."),
	)

	scanner := &scanner.HeaderScanner{}
	if scanner.Scan(v) {
		t.Errorf("message should not be unsubscribe-able, but it was")
	}
}
