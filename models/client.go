package models

type Client struct {
	ID          int    `gorm:"primaryKey" json:"id,omitempty"`
	PhoneNumber int    `gorm:"type:int8;not null" json:"phone_number,omitempty" example:"79998887766"`
	PhoneCode   string `gorm:"type:varchar(255);not null" json:"phone_code,omitempty"`
	Tag         string `gorm:"type:varchar(255);not null" json:"tag,omitempty"`
	TimeZone    string `gorm:"type:varchar(255);not null" json:"time_zone,omitempty"`
}

func (n Client) Clone() *Client {
	return &Client{
		ID:          n.ID,
		PhoneNumber: n.PhoneNumber,
		PhoneCode:   n.PhoneCode,
		Tag:         n.Tag,
		TimeZone:    n.TimeZone,
	}
}

type ClientRequest struct {
	PhoneNumber int    `json:"phone_number,omitempty"`
	PhoneCode   string `json:"phone_code,omitempty"`
	Tag         string `json:"tag,omitempty"`
	TimeZone    string `json:"time_zone,omitempty"`
}
