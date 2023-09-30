package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"notification-service/inits"
)

type Notification struct {
	ID      int       `gorm:"primaryKey" json:"id,omitempty"`
	Text    string    `gorm:"type:varchar(255);not null" json:"text,omitempty"`
	Filter  Filter    `gorm:"jsonb;not null" json:"filter,omitempty"`
	StartAt time.Time `json:"created_at,omitempty" example:"2023-10-29T15:04:05Z" format:"date-time"`
	EndAt   time.Time `json:"end_at,omitempty" example:"2023-10-29T15:04:05Z" format:"date-time"`
}

func (n Notification) Clone() *Notification {
	return &Notification{
		ID:      n.ID,
		Text:    n.Text,
		Filter:  n.Filter,
		StartAt: n.StartAt,
		EndAt:   n.EndAt,
	}
}

type NotificationRequest struct {
	Text    string    `json:"text,omitempty"`
	Filter  Filter    `json:"filter,omitempty"`
	StartAt time.Time `json:"created_at,omitempty"`
	EndAt   time.Time `json:"end_at,omitempty"`
}

type Filter struct {
	jsonF     inits.JSONB `json:"-"`
	PhoneCode string      `gorm:"type:varchar(255);not null" json:"phone_code,omitempty"`
	Tag       string      `gorm:"type:varchar(255);not null" json:"tag,omitempty"`
}

func (f Filter) Value() (driver.Value, error) {
	if f.jsonF == nil {
		f.jsonF = make(inits.JSONB)
	}
	f.jsonF["tag"] = f.Tag
	f.jsonF["phone_code"] = f.PhoneCode
	return f.jsonF.Value()
}

func (f *Filter) Scan(value interface{}) error {
	err := f.jsonF.Scan(value)
	if err == nil {
		f.Tag = fmt.Sprintf("%s", f.jsonF["tag"])
		f.PhoneCode = fmt.Sprintf("%s", f.jsonF["phone_code"])
	}
	return err
}
