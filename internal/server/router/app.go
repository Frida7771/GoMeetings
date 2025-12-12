package router

import (
	"GoMeetings/internal/middlewares"
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

	r.Use(middlewares.Cors())

	auth := r.Group("/auth", middlewares.Auth())
	//user login
	auth.POST("/user/login", service.UserLogin)

	//meeting list
	auth.GET("/meeting/list", service.MeetingList)

	//meeting create
	auth.POST("/meeting/create", service.MeetingCreate)
	//meeting edit
	auth.PUT("/meeting/edit", service.MeetingEdit)

	//meeting delete
	auth.DELETE("/meeting/delete", service.MeetingDelete)

	return r
}
