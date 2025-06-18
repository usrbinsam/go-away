package store_test

import (
	"testing"

	"github.com/usrbinsam/go-away/internal/store"
)

func TestStore_Open(t *testing.T) {
	st := store.SqlStore{}
	err := st.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	t.Run("Unsubscribed", func(t *testing.T) {
		st.RecordUnsubscribe("aabbcc", "list@list.org", "sam@example.com")

		if !st.Unsubscribed("list@list.org", "sam@example.com") {
			t.Errorf("expected to find unsubscribe record")
		}
	})

	t.Run("Seen", func(t *testing.T) {
		st.MarkSeen("aabbcc", "sam@example.com")
		if !st.Seen("aabbcc", "sam@example.com") {
			t.Errorf("expected to find seen record")
		}

		if st.Seen("notseen", "sam@xample.com") {
			t.Errorf("expected not to find seen record for non-existent message")
		}
	})
}
