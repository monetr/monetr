package pubsub

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (
	Notification interface {
		Channel() string
		Payload() string
	}

	PublishSubscribe interface {
		Subscribe(ctx context.Context, channel string) (Listener, error)
		Notify(ctx context.Context, channel, payload string) error
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
		channel      string
		log          *logrus.Entry
		listener     *pg.Listener
		closeChannel chan struct{}
		dataChannel  chan Notification
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

func (p *postgresPubSub) Subscribe(ctx context.Context, channel string) (Listener, error) {
	listener := p.db.Listen(ctx, channel)

	pgListener := &postgresListener{
		channel:      channel,
		log:          p.log.WithContext(ctx).WithField("channel", channel),
		listener:     listener,
		dataChannel:  make(chan Notification, 1),
		closeChannel: make(chan struct{}, 1),
	}
	go pgListener.backgroundListener()
	return pgListener, nil
}

func (p *postgresPubSub) Notify(ctx context.Context, channel, payload string) error {
	_, err := p.db.ExecContext(ctx, fmt.Sprintf(`NOTIFY %s, ?`, channel), payload)

	return errors.Wrap(err, "failed to notify channel")
}

func (p *postgresListener) backgroundListener() {
	for {
		select {
		case message := <-p.listener.Channel():
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
			_ = p.listener.Unlisten(context.Background(), p.channel)
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
