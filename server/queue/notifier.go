package queue

import (
	"context"
)

// Notifier is an interface that surfaces how the queue will know when to
// consume jobs from the queue. Notifications from this interface do not promise
// that a job can or will be consumed. Just that when a notification is sent on
// this interface, the consumer of the notification should attempt to consume a
// job from the queue. This interface will be implemented differently depending
// on the datastore that is backing the queue. For example, PostgreSQL will
// likely have a timer as well as a LISTEN/NOTIFY flow for job notifications.
// Where as SQLite will likely just use an in memory channel to notify since it
// is always a single process and polling the database won't make sense.
type Notifier interface {
	// channel returns a buffered channel that can be used to get notified when a
	// new job is ready to be processed. This fires when a job is enqueued by any
	// server in the cluster where the job's timestamp/priority is now or in the
	// past. Jobs that are timestamped to be processed in the future are not
	// notified here.
	channel() <-chan struct{}

	// notify sends a notification to all other servers consuming the job queue.
	// This tells them there is a job to be consumed. It will return an error if
	// it fails to send the notification.
	notify(ctx context.Context) error
}

var (
	_ Notifier = &memoryNotifier{}
)

type memoryNotifier struct {
	notifications chan struct{}
}

func NewMemoryNotifier(size int) Notifier {
	return &memoryNotifier{
		notifications: make(chan struct{}, size),
	}
}

// channel implements [Notifier].
func (m *memoryNotifier) channel() <-chan struct{} {
	return m.notifications
}

// notify implements [Notifier].
func (m *memoryNotifier) notify(ctx context.Context) error {
	// Try to send an empty struct on the notification channel. Context is not
	// used here because we do not want to block. If the notification channel is
	// full then we don't want to do anything at all and we can just return.
	select {
	case m.notifications <- struct{}{}:
	default:
	}
	return nil
}
