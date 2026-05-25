package messages

import (
	"sort"
	"sync"
)

// Message is a single private chat message.
type Message struct {
	From string
	To   string
	Text string
	Time int64
}

var (
	inbox = make(map[string][]Message)
	mu    sync.Mutex
)

// Add stores a message in the recipient inbox.
func Add(to string, msg Message) {
	mu.Lock()
	defer mu.Unlock()
	inbox[to] = append(inbox[to], msg)
}

// Inbox returns a copy of messages received by a user.
func Inbox(login string) []Message {
	mu.Lock()
	defer mu.Unlock()
	result := make([]Message, len(inbox[login]))
	copy(result, inbox[login])
	return result
}

// Dialog returns messages between two users ordered by creation time.
func Dialog(login, with string) []Message {
	mu.Lock()
	defer mu.Unlock()

	var all []Message
	for _, m := range inbox[login] {
		if m.From == with {
			all = append(all, m)
		}
	}
	for _, m := range inbox[with] {
		if m.From == login {
			all = append(all, m)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Time < all[j].Time
	})
	return all
}
