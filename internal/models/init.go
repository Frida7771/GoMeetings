package models

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func NewDB() {
	pass := os.Getenv("DB_PASS")

	dsn := fmt.Sprintf(
		"root:%s@tcp(127.0.0.1:3306)/meeting?charset=utf8mb4&parseTime=True&loc=Local",
		pass,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	db.AutoMigrate(&RoomBasic{}, &RoomUser{}, &UserBasic{}, &RoomScreenShare{})

	DB = db
}
