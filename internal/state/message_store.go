package state

import (
	"sync"
	"time"

	"github.com/exogmi/gossip/internal/models"
)

type MessageStore struct {
	messages    map[string][]*models.Message // Key: target (channel name or user nickname)
	maxMessages int                          // Maximum number of messages to store per target
	mu          sync.RWMutex
}

func NewMessageStore(maxMessages int) *MessageStore {
	return &MessageStore{
		messages:    make(map[string][]*models.Message),
		maxMessages: maxMessages,
	}
}

func (ms *MessageStore) StoreMessage(message *models.Message) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	target := message.Target
	ms.messages[target] = append(ms.messages[target], message)

	// Prune old messages if we've exceeded the maximum
	if len(ms.messages[target]) > ms.maxMessages {
		ms.messages[target] = ms.messages[target][len(ms.messages[target])-ms.maxMessages:]
	}

	return nil
}

func (ms *MessageStore) GetMessages(target string, limit int) ([]*models.Message, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	messages := ms.messages[target]
	if len(messages) < limit {
		return messages, nil
	}
	return messages[len(messages)-limit:], nil
}

func (ms *MessageStore) ClearMessages(target string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.messages, target)
	return nil
}

func (ms *MessageStore) PruneOldMessages() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for target, messages := range ms.messages {
		if len(messages) > ms.maxMessages {
			ms.messages[target] = messages[len(messages)-ms.maxMessages:]
		}
	}
}

// StartPeriodicCleanup starts a goroutine that periodically prunes old messages
func (ms *MessageStore) StartPeriodicCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			ms.PruneOldMessages()
		}
	}()
}
