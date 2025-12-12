package router

import (
	"GoMeetings/internal/server/service"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//meeting
	r.POST("/meeting/create", service.MeetingCreate)

	//user login
	r.POST("/user/login", service.UserLogin)
	return r
}
