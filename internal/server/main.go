package main

// @title GoMeetings API
// @version 1.0
// @description Meeting management & WebRTC signaling backend
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
import (
	"GoMeetings/internal/models"
	"GoMeetings/internal/server/router"
	"log"

	_ "GoMeetings/docs"

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
