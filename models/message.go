package models

import (
	"time"
)

type Message struct {
	ID             int           `gorm:"primaryKey" json:"id,omitempty"`
	ClientID       int           `gorm:"type:bigserial;index:idx_messages_clients" json:"client_id,omitempty"`
	NotificationID int           `gorm:"type:bigserial;index:idx_messages_notifications" json:"notification_id,omitempty"`
	Client         *Client       `gorm:"foreignKey:ClientID;association_foreignkey:id" json:"-"`
	Notification   *Notification `gorm:"foreignKey:NotificationID;association_foreignkey:id" json:"-"`
	Text           string        `gorm:"type:varchar(255);not null" json:"text,omitempty"`
	State          State         `gorm:"type:varchar(255);not null" json:"state,omitempty"`
	CreatedAt      time.Time     `json:"created_at,omitempty"`
}

type State string

const (
	Created State = "CREATED"
	Error   State = "ERROR"
	Sent    State = "SENT"
)

type MsgPost struct {
	ID          int    `json:"id,omitempty"`
	Text        string `json:"text,omitempty"`
	PhoneNumber int    `json:"phone_number,omitempty"`
}

func (m MsgPost) Clone() *MsgPost {
	return &MsgPost{
		ID:          m.ID,
		PhoneNumber: m.PhoneNumber,
		Text:        m.Text,
	}
}

type MsgResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MsgPostEvent struct {
	Msg            *MsgPost
	NotificationID int
}

type MsgResponseEvent struct {
	MsgResponse *MsgResponse
	Msg         *MsgPost
}
