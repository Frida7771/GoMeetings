package models

import "gorm.io/gorm"

type RoomScreenShare struct {
	gorm.Model
	Rid       uint   `gorm:"column:rid;type:int(11);not null;index" json:"rid"`
	OwnerUid  uint   `gorm:"column:owner_uid;type:int(11);not null" json:"owner_uid"`
	StreamID  string `gorm:"column:stream_id;type:varchar(128)" json:"stream_id"`
	Active    bool   `gorm:"column:active;type:tinyint(1);not null;default:1" json:"active"`
	StartedAt int64  `gorm:"column:started_at;type:bigint;not null" json:"started_at"`
	EndedAt   *int64 `gorm:"column:ended_at;type:bigint" json:"ended_at"`
}

func (RoomScreenShare) TableName() string {
	return "room_screen_share"
}
