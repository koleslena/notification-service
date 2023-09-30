package services

import (
	"notification-service/models"
	"notification-service/services/interfaces"
)

type ChannelsSizes struct {
	MessagesSize      int
	ResponsesSize     int
	NotificationsSize int
}

type Channels struct {
	messages      chan *models.MsgPostEvent
	msgResponses  chan *models.MsgResponseEvent
	notifications chan *models.Notification
	logger        interfaces.Logger
}

func NewChannels(sizes ChannelsSizes, logger interfaces.Logger) *Channels {
	return &Channels{
		messages:      make(chan *models.MsgPostEvent, sizes.MessagesSize),
		msgResponses:  make(chan *models.MsgResponseEvent, sizes.ResponsesSize),
		notifications: make(chan *models.Notification, sizes.NotificationsSize),
		logger:        logger,
	}
}

func (ch *Channels) MessagesChan() chan *models.MsgPostEvent {
	return ch.messages
}

func (ch *Channels) ResponsesChan() chan *models.MsgResponseEvent {
	return ch.msgResponses
}

func (ch *Channels) NotificationsChan() chan *models.Notification {
	return ch.notifications
}

func (ch *Channels) AddMessage(msg *models.MsgPostEvent) {
	ch.logger.Infof("add message to channel {%v}", msg)
	ch.messages <- msg
}

func (ch *Channels) AddResponse(resp *models.MsgResponseEvent) {
	ch.logger.Infof("add response to channel {%v}", resp)
	ch.msgResponses <- resp
}

func (ch *Channels) AddNotification(n *models.Notification) {
	ch.logger.Infof("add notification to channel {%v}", n)
	ch.notifications <- n
}
