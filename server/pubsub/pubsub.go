package pubsub

import (
	"context"
	"fmt"
	"hash/crc32"
	"log/slog"
	"strings"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/uptrace/bun"
)

type (
	Notification interface {
		// Channel returns the hashed channel that is used in the notification.
		Channel() string
		Payload() string
	}

	PublishSubscribe interface {
		// Subscribe will subscribe to a specific channel for the specified account.
		// The channel name specified will not be the one used in the actual
		// notification system. Instead the channel name is hashed in order to allow
		// for longer channel names than the underlying pubsub system may allow.
		Subscribe(
			ctx context.Context,
			accountId models.ID[models.Account],
			channel string,
		) (Listener, error)
		Publisher
	}

	Publisher interface {
		// Notify will push a notification for a specific account on the specified
		// channel. Note that the channel name is hashed when the notification is
		// sent.
		Notify(
			ctx context.Context,
			accountId models.ID[models.Account],
			channel, payload string,
		) error
	}

	Listener interface {
		Channel() <-chan Notification
		Close() error
	}
)

// hashChannel creates a new channel slug based on the provided account ID and
// the desired channel name. The account ID is not modified and is prefixed on
// the channel name. But the hash of the channel is added to make sure that even
// if the channel name is too long that we don't accidently subscribe to a
// prefix.
func hashChannel(
	accountId models.ID[models.Account],
	channel string,
) string {
	return fmt.Sprintf(
		"%s:%08X",
		accountId.String(),
		crc32.Checksum([]byte(channel), crc32.IEEETable),
	)
}

func NewPubSub(
	log *slog.Logger,
	db *bun.DB,
	configuration config.Configuration,
) PublishSubscribe {
	switch strings.ToLower(string(configuration.Database.Kind)) {
	case "postgresql":
		return NewPostgresPubSub(log, db)
	case "sqlite":
		return NewInMemoryPubSub(log, 8)
	// TODO Add valkey based pubsub?
	default:
		panic("invalid database type configured!")
	}
}
