package messages

import (
	"sort"
	"sync"
)

// Message одна запись чата
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

// Add кладет получателю
func Add(to string, msg Message) {
	mu.Lock()
	defer mu.Unlock()
	inbox[to] = append(inbox[to], msg)
}

// Inbox копия входящих
func Inbox(login string) []Message {
	mu.Lock()
	defer mu.Unlock()
	result := make([]Message, len(inbox[login]))
	copy(result, inbox[login])
	return result
}

// Dialog диалог с сортировкой по времени
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
