package router

import (
	"GoMeetings/internal/middlewares"
	"GoMeetings/internal/server/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", pingHandler)

	r.Use(middlewares.Cors())

	// Static files for test pages
	r.Static("/test", "./internal/test")

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebRTC signaling websocket (no auth required to keep demo simple)
	r.GET("/ws/p2p/:roomIdentity/:userIdentity", service.SignalWebsocket)

	publicAuth := r.Group("/auth")
	publicAuth.POST("/user/login", service.UserLogin)
	publicAuth.POST("/user/register", service.UserRegister)

	auth := r.Group("/auth", middlewares.Auth())

	room := auth.Group("/room")
	room.GET("/list", service.RoomList)
	room.POST("/create", service.RoomCreate)
	room.PUT("/edit", service.RoomEdit)
	room.DELETE("/delete", service.RoomDelete)
	room.POST("/join", service.RoomJoin)
	room.POST("/leave", service.RoomLeave)
	room.GET("/members", service.RoomMembers)
	room.GET("/user-rooms", service.RoomUserRooms)
	room.POST("/share/start", service.RoomShareStart)
	room.POST("/share/stop", service.RoomShareStop)
	room.GET("/share/status", service.RoomShareStatus)

	return r
}

// Ping godoc
// @Summary Health check
// @Tags Public
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
