package models

import (
	"time"

	"gorm.io/gorm"
)

type RoomBasic struct {
	gorm.Model
	Identify  string    `gorm:"column:identify;type:varchar(36);uniqueIndex;not null" json:"identify"`
	Name      string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	BeginAt   time.Time `gorm:"column:begin_at;type:datetime;not null" json:"begin_at"`
	EndAt     time.Time `gorm:"column:end_at;type:datetime;not null" json:"end_at"`
	CreateID  uint      `gorm:"column:create_id;type:int(20);not null" json:"create_id"` //create_id
	JoinCode  string    `gorm:"column:join_code;type:varchar(16);not null" json:"-"`
	ShortCode string    `gorm:"column:short_code;type:varchar(16)" json:"-"`
}

func (table *RoomBasic) TableName() string {
	return "room_basic"
}
