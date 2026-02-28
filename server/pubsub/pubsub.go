package pubsub

import (
	"context"
	"fmt"
	"hash/crc32"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

var (
	_ Notification     = &postgresNotification{}
	_ PublishSubscribe = &postgresPubSub{}
	_ Listener         = &postgresListener{}
)

type (
	postgresNotification struct {
		base pg.Notification
	}

	postgresPubSub struct {
		log *logrus.Entry
		db  *pg.DB
	}

	postgresListener struct {
		accountId     models.ID[models.Account]
		channel       string
		hashedChannel string
		log           *logrus.Entry
		listener      *pg.Listener
		closeChannel  chan struct{}
		dataChannel   chan Notification
	}
)

func NewPostgresPubSub(log *logrus.Entry, db *pg.DB) PublishSubscribe {
	return &postgresPubSub{
		log: log,
		db:  db,
	}
}

func (p *postgresNotification) Channel() string {
	return p.base.Channel
}

func (p *postgresNotification) Payload() string {
	return p.base.Payload
}

func (p *postgresPubSub) Subscribe(
	ctx context.Context,
	accountId models.ID[models.Account],
	channel string,
) (Listener, error) {
	hashedChannel := p.hashChannel(accountId, channel)
	listener := p.db.Listen(ctx, hashedChannel)

	pgListener := &postgresListener{
		accountId:     accountId,
		channel:       channel,
		hashedChannel: hashedChannel,
		log:           p.log.WithContext(ctx).WithField("channel", channel),
		listener:      listener,
		dataChannel:   make(chan Notification, 0),
		closeChannel:  make(chan struct{}, 1),
	}
	go pgListener.backgroundListener()
	return pgListener, nil
}

// hashChannel creates a new channel slug based on the provided account ID and
// the desired channel name. The account ID is not modified and is prefixed on
// the channel name. But the hash of the channel is added to make sure that even
// if the channel name is too long that we don't accidently subscribe to a
// prefix.
func (p *postgresPubSub) hashChannel(
	accountId models.ID[models.Account],
	channel string,
) string {
	return fmt.Sprintf(
		"%s:%08X",
		accountId.String(),
		crc32.Checksum([]byte(channel), crc32.IEEETable),
	)
}

func (p *postgresPubSub) Notify(
	ctx context.Context,
	accountId models.ID[models.Account],
	channel, payload string,
) error {
	span := sentry.StartSpan(ctx, "PubSub - Notify")
	defer span.Finish()

	hashedChannel := p.hashChannel(accountId, channel)

	p.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"channel":       channel,
			"hashedChannel": hashedChannel,
		}).
		Debug("sending notification on channel")

	_, err := p.db.ExecContext(
		span.Context(),
		fmt.Sprintf(`NOTIFY "%s", ?`, hashedChannel),
		payload,
	)

	return errors.Wrap(err, "failed to notify channel")
}

func (p *postgresListener) backgroundListener() {
	for {
		select {
		case message := <-p.listener.Channel():
			if message.Channel != p.hashedChannel {
				p.log.WithFields(logrus.Fields{
					"channel":       p.channel,
					"hashedChannel": p.hashedChannel,
					"received":      message.Channel,
				}).Warn("ignoring message on channel")
				continue
			}
			transformedMessage := &postgresNotification{
				base: message,
			}

			select {
			case p.dataChannel <- transformedMessage:
				p.log.Trace("successfully dispatched notification")
			default:
				p.log.Trace("message on channel dropped because data channel is full")
			}
		case <-p.closeChannel:
			p.log.Trace("received close message, exiting loop")
			close(p.closeChannel)
			close(p.dataChannel)
			// Release the listener connection
			_ = p.listener.Unlisten(context.Background(), p.hashedChannel)
			_ = p.listener.Close()
			return
		}
	}
}

func (p *postgresListener) Channel() <-chan Notification {
	return p.dataChannel
}

func (p *postgresListener) Close() error {
	p.closeChannel <- struct{}{}
	return nil
}
