package main

import (
	"GoMeetings/internal/models"
	"GoMeetings/internal/server/router"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	models.NewDB()
	e := router.Router()
	err := e.Run()
	if err != nil {
		log.Fatalln("run error.", err)
		return
	}
}
