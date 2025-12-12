package service

import (
	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func MeetingCreate(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	in := MeetingCreateRequest{}
	err := c.ShouldBindJSON(&in)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params error" + err.Error(),
		})
		return
	}

	err = models.DB.Create(&models.RoomBasic{
		Identify: helper.GenerateUUID(),
		Name:     in.Name,
		BeginAt:  time.UnixMilli(in.BeginAt),
		EndAt:    time.UnixMilli(in.EndAt),
		CreateID: uc.Id,
	}).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "system error" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
	})
}
