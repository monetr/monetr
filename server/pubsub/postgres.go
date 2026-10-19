package pubsub

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	_ Notification     = &postgresNotification{}
	_ PublishSubscribe = &postgresPubSub{}
	_ Listener         = &postgresListener{}
)

type (
	postgresNotification struct {
		base pgdriver.Notification
	}

	postgresPubSub struct {
		log *slog.Logger
		db  *bun.DB
	}

	postgresListener struct {
		accountId     models.ID[models.Account]
		channel       string
		hashedChannel string
		log           *slog.Logger
		listener      *pgdriver.Listener
		closeChannel  chan struct{}
		dataChannel   chan Notification
	}
)

func NewPostgresPubSub(
	log *slog.Logger,
	db *bun.DB,
) PublishSubscribe {
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
	hashedChannel := hashChannel(accountId, channel)
	listener := pgdriver.NewListener(p.db)
	if err := listener.Listen(ctx, hashedChannel); err != nil {
		return nil, errors.Wrap(err, "failed to listen on channel")
	}

	pgListener := &postgresListener{
		accountId:     accountId,
		channel:       channel,
		hashedChannel: hashedChannel,
		log:           p.log.With("channel", channel),
		listener:      listener,
		dataChannel:   make(chan Notification),
		closeChannel:  make(chan struct{}, 1),
	}
	go pgListener.backgroundListener()
	return pgListener, nil
}

func (p *postgresPubSub) Notify(
	ctx context.Context,
	accountId models.ID[models.Account],
	channel, payload string,
) error {
	span := sentry.StartSpan(ctx, "PubSub - Notify")
	defer span.Finish()

	hashedChannel := hashChannel(accountId, channel)

	p.log.DebugContext(span.Context(), "sending notification on channel",
		"channel", channel,
		"hashedChannel", hashedChannel,
	)

	err := pgdriver.Notify(span.Context(), p.db, hashedChannel, payload)

	return errors.Wrap(err, "failed to notify channel")
}

func (p *postgresListener) backgroundListener() {
	notifications := p.listener.Channel()
	for {
		select {
		case message := <-notifications:
			if message.Channel != p.hashedChannel {
				p.log.WarnContext(context.Background(), "ignoring message on channel",
					"channel", p.channel,
					"hashedChannel", p.hashedChannel,
					"received", message.Channel,
				)
				continue
			}
			transformedMessage := &postgresNotification{
				base: message,
			}

			select {
			case p.dataChannel <- transformedMessage:
				p.log.Log(context.Background(), logging.LevelTrace, "successfully dispatched notification")
			default:
				p.log.Log(context.Background(), logging.LevelTrace, "message on channel dropped because data channel is full")
			}
		case <-p.closeChannel:
			p.log.Log(context.Background(), logging.LevelTrace, "received close message, exiting loop")
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
