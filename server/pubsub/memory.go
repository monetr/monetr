package pubsub

import (
	"context"
	"log/slog"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

var (
	_ PublishSubscribe = &memoryPubSub{}
	_ Listener         = &memoryListener{}
	_ Notification     = memoryNotification{}
)

type memoryNotification struct {
	channel string
	payload string
}

func (n memoryNotification) Channel() string {
	return n.channel
}

func (n memoryNotification) Payload() string {
	return n.payload
}

type memoryPubSub struct {
	log       *slog.Logger
	lock      sync.RWMutex
	listeners map[string]map[*memoryListener]struct{}
	buffer    int
}

func NewInMemoryPubSub(
	log *slog.Logger,
	buffer int,
) PublishSubscribe {
	return &memoryPubSub{
		log:       log,
		listeners: map[string]map[*memoryListener]struct{}{},
		buffer:    buffer,
	}
}

func (m *memoryPubSub) Subscribe(
	ctx context.Context,
	accountId models.ID[models.Account],
	channel string,
) (Listener, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	hashed := hashChannel(accountId, channel)
	listener := &memoryListener{
		channel: hashed,
		ch:      make(chan Notification, m.buffer),
		done:    make(chan struct{}),
		parent:  m,
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if m.listeners[hashed] == nil {
		m.listeners[hashed] = map[*memoryListener]struct{}{}
	}
	m.listeners[hashed][listener] = struct{}{}

	return listener, nil
}

func (m *memoryPubSub) Notify(
	ctx context.Context,
	accountId models.ID[models.Account],
	channel, payload string,
) error {
	span := sentry.StartSpan(ctx, "PubSub - Notify")
	defer span.Finish()

	hashedChannel := hashChannel(accountId, channel)
	notification := memoryNotification{
		channel: hashedChannel,
		payload: payload,
	}

	m.log.DebugContext(span.Context(), "sending notification on channel",
		"channel", channel,
		"hashedChannel", hashedChannel,
	)

	m.lock.RLock()
	defer m.lock.RUnlock()
	for listener := range m.listeners[hashedChannel] {
		select {
		case listener.ch <- notification:
		case <-listener.done:
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "failed to notify channel")
		}
	}

	return nil
}

type memoryListener struct {
	channel string
	ch      chan Notification
	done    chan struct{}
	once    sync.Once
	parent  *memoryPubSub
}

func (l *memoryListener) Channel() <-chan Notification {
	return l.ch
}

func (l *memoryListener) Close() error {
	l.once.Do(func() {
		// Unblock any notify blocked on this listener
		close(l.done)

		l.parent.lock.Lock()
		delete(l.parent.listeners[l.channel], l)
		if len(l.parent.listeners[l.channel]) == 0 {
			delete(l.parent.listeners, l.channel)
		}
		l.parent.lock.Unlock()

		close(l.ch)
	})
	return nil
}
