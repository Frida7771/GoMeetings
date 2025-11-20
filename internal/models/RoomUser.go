package models

import "gorm.io/gorm"

type RoomUser struct {
	gorm.Model
	Rid uint `gorm:"column:rid;type:int(11);not null" json:"rid"` //room id
	Uid uint `gorm:"column:uid;type:int(11);not null" json:"uid"` //user id
}

func (table *RoomUser) TableName() string {
	return "room_user"
}
