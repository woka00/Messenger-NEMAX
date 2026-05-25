package messages

import "testing"

func TestDialogReturnsBothSidesOrderedByTime(t *testing.T) {
	mu.Lock()
	previous := inbox
	inbox = make(map[string][]Message)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		inbox = previous
		mu.Unlock()
	})

	Add("bob", Message{From: "alice", To: "bob", Text: "second", Time: 2})
	Add("alice", Message{From: "bob", To: "alice", Text: "first", Time: 1})

	dialog := Dialog("alice", "bob")
	if len(dialog) != 2 || dialog[0].Text != "first" || dialog[1].Text != "second" {
		t.Fatalf("Dialog() = %#v, want messages ordered by time", dialog)
	}
}
