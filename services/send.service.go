package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"notification-service/inits"
	"notification-service/models"
	"notification-service/services/interfaces"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SendService struct {
	logger    interfaces.Logger
	config    *inits.Config
	DB        *gorm.DB
	channels  *Channels
	scheduler *Scheduler
}

func NewSendService(DB *gorm.DB, channels *Channels, logger interfaces.Logger, config *inits.Config) *SendService {
	return &SendService{
		DB:        DB,
		logger:    logger,
		config:    config,
		channels:  channels,
		scheduler: NewScheduler(),
	}
}

func (ss *SendService) Init(ctx context.Context) error {
	for {
		select {
		case msg := <-ss.channels.MessagesChan():
			ss.logger.Infof("message to send received {%v}", msg)
			exists, err := ss.scheduler.Exists(msg.NotificationID)
			if exists && err != ErrIDNotFound {
				go func() {
					response, err := ss.sendMessage(msg)
					ss.channels.AddResponse(&models.MsgResponseEvent{MsgResponse: response, Msg: msg.Msg})
					if err != nil {
						ss.channels.AddMessage(msg)
					}
				}()
			}
		case msgResp := <-ss.channels.ResponsesChan():
			ss.logger.Infof("save response received {%v}", msgResp)
			ss.saveResult(msgResp)
		case notification := <-ss.channels.NotificationsChan():
			ss.logger.Infof("notification received {%v}", notification)
			exists, err := ss.scheduler.Exists(notification.ID)
			if exists {
				err := ss.updateTask(notification)
				if err != nil {
					ss.logger.Error("error update task for notification: $v", notification)
				}
			} else if err == ErrIDNotFound {
				err := ss.addTask(notification)
				if err != nil {
					ss.logger.Error("error adding new task for notification: $v", notification)
				}
			}
		case <-ctx.Done():
			ss.logger.Info("context done is received, stop sending messages")
			return errors.Errorf("context done is received, stop sending messages")
		}
	}
}

func (ss *SendService) InitTasks() error {
	ss.logger.Info("init tasks started.......")
	var notificationsDB []models.Notification
	results := ss.DB.Find(&notificationsDB)
	if results.Error != nil {
		return results.Error
	}
	for _, notification := range notificationsDB {
		ss.channels.AddNotification(notification.Clone())
	}
	ss.logger.Info("init tasks finished......")
	return nil
}

func (ss *SendService) addTask(notification *models.Notification) error {
	if notification.EndAt.Before(time.Now()) {
		return nil
	}
	at := notification.StartAt
	if notification.StartAt.Before(time.Now()) {
		at = time.Now().Add(10 * time.Second)
	}
	ss.logger.Infof("task added at {%v}", at)
	var t = &Task{id: notification.ID,
		StartAfter: at,
		Interval:   30 * time.Minute,
		StopTime:   notification.EndAt,
		TaskFunc: func() error {
			return ss.startSend(notification)
		}, ErrFunc: func(err error) {
			ss.startSendError(notification)
		}}
	err := ss.scheduler.AddTask(notification.ID, t)
	if err != nil {
		ss.startSendError(notification)
		ss.logger.Errorf("task add error {%v}", err)
		return err
	}
	ss.logger.Infof("task added {%v}", t)
	return nil
}

func (ss *SendService) updateTask(notification *models.Notification) error {
	var t, err = ss.scheduler.Lookup(notification.ID)
	if err != nil {
		ss.logger.Errorf("error update task for notification: $v", notification)
		return err
	}
	t.UpdateTask(func() {
		t.StartAfter = notification.StartAt
		t.StopTime = notification.EndAt
	})
	ss.logger.Infof("task updated {%v}", t)
	return nil
}

func (ss *SendService) sendMessage(msg *models.MsgPostEvent) (*models.MsgResponse, error) {
	ss.logger.Infof("send message started {%v}", msg)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := ss.config.CreateRequestWithAuth("POST", msg.Msg.ID, msg.Msg)
	if err != nil {
		ss.logger.Errorf("error on creating request to sender service: %v", err)
		return &models.MsgResponse{Code: 500, Message: err.Error()}, err
	} else {
		response, err := client.Do(req)
		if err != nil {
			ss.logger.Errorf("error on getting response from sender service: %v", err)
			return &models.MsgResponse{Code: 500, Message: err.Error()}, err
		} else {
			if response.StatusCode != 200 {
				ss.logger.Errorf("read response body error, err: %v", err)
				return &models.MsgResponse{Code: 500, Message: err.Error()}, err
			}
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				ss.logger.Errorf("read response body error, err: %v", err)
				return &models.MsgResponse{Code: 500, Message: err.Error()}, err
			}
			msgResponse := &models.MsgResponse{}
			err = json.Unmarshal(body, msgResponse)
			if err != nil {
				ss.logger.Errorf("response body parsing error, err: %v", err)
				return &models.MsgResponse{Code: 500, Message: err.Error()}, err
			}
			if msgResponse.Code != 0 {
				ss.logger.Errorf("error response fromm sending server, err: %v", msgResponse.Message)
				return &models.MsgResponse{Code: 500, Message: err.Error()}, err
			}
			ss.logger.Infof("response message received {%v}", msgResponse)
			return msgResponse, nil
		}
	}
}

func (ss *SendService) startSendError(notification *models.Notification) {
	ss.logger.Infof("restart notification {%v}", notification)
	ss.channels.AddNotification(notification)
}

func (ss *SendService) startSend(notification *models.Notification) error {
	ss.logger.Infof("start process notification {%v}", notification)
	var clients []models.Client
	var existsMessages bool
	results := ss.DB.Model(models.Message{}).Select("count(*) > 0").Where("notification_id = ?", notification.ID).Find(&existsMessages)

	ss.logger.Infof("for notification {%v}; messages exist {%b}", notification, existsMessages)
	if existsMessages {
		var errorMessages []models.MsgPost
		ss.DB.Table("messages").Select("messages.id, notifications.text, clients.phone_number").
			Joins("JOIN notifications ON notifications.id = messages.notification_id").
			Joins("JOIN clients ON clients.id = messages.client_id").
			Where("messages.state != ?", models.Sent).
			Scan(&errorMessages)
		if results.Error != nil {
			return results.Error
		}
		for _, msgPost := range errorMessages {
			msg := &models.MsgPostEvent{
				NotificationID: notification.ID,
				Msg:            msgPost.Clone(),
			}
			ss.channels.AddMessage(msg)
		}
		ss.DB.Raw("SELECT * FROM clients cc "+
			"WHERE (cc.phone_code = ? or cc.tag = ?) "+
			"and not exists(select * from messages mm where mm.client_id = cc.id)", notification.Filter.PhoneCode, notification.Filter.Tag).
			Find(&clients)
		if results.Error != nil {
			return results.Error
		}
		if len(clients) > 0 {
			ss.createMessages(notification, clients)
		}
		if len(clients) == 0 && len(errorMessages) == 0 {
			ss.logger.Infof("notification finish sending {%v}", notification)
			ss.scheduler.Del(notification.ID)
		}
	} else {
		results := ss.DB.Where("phone_code = ? ", notification.Filter.PhoneCode).
			Or("tag = ? ", notification.Filter.Tag).
			Find(&clients)
		if results.Error != nil {
			return results.Error
		}
		ss.createMessages(notification, clients)
	}
	ss.logger.Infof("finish process notification {%v}", notification)
	return nil
}

func (ss *SendService) createMessages(notification *models.Notification, clients []models.Client) {
	for _, client := range clients {
		message := models.Message{
			Text:           notification.Text,
			ClientID:       client.ID,
			NotificationID: notification.ID,
			State:          models.Created,
			CreatedAt:      time.Now(),
		}
		result := ss.DB.Create(&message)
		if result.Error != nil {
			ss.logger.Errorf("error on creating message in db: %v", result.Error)
		} else {
			msg := &models.MsgPostEvent{
				NotificationID: notification.ID,
				Msg: &models.MsgPost{
					ID:          message.ID,
					Text:        notification.Text,
					PhoneNumber: client.PhoneNumber,
				},
			}
			ss.channels.AddMessage(msg)
		}
	}
}

func (ss *SendService) saveResult(msgResp *models.MsgResponseEvent) {
	ss.logger.Infof("start save message state {%v}", msgResp)
	state := models.Sent
	if msgResp.MsgResponse.Code != 0 {
		state = models.Error
	}
	ss.DB.Model(&models.Message{}).Where("id = ?", msgResp.Msg.ID).Update("state", state)
}
