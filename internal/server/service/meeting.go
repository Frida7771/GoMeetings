package service

import (
	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func MeetingList(c *gin.Context) {
	in := new(MeetingListRequest)
	err := c.ShouldBindQuery(in)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params error: " + err.Error(),
		})
		return
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Size <= 0 {
		in.Size = 20
	}

	var list []models.RoomBasic
	var cnt int64

	tx := models.DB.Model(&models.RoomBasic{})
	if in.Keyword != "" {
		tx = tx.Where("name LIKE ?", "%"+in.Keyword+"%")
	}

	err = tx.Count(&cnt).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "system error: " + err.Error(),
		})
		return
	}

	err = tx.Limit(in.Size).Offset((in.Page - 1) * in.Size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "system error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  list,
			"count": cnt,
		},
	})
}

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

func MeetingEdit(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	in := MeetingEditRequest{}
	err := c.ShouldBindJSON(&in)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params error: " + err.Error(),
		})
		return
	}
	if in.Identify == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params error: identity is required",
		})
		return
	}

	result := models.DB.Model(new(models.RoomBasic)).Where("identify = ? AND create_id = ?", in.Identify, uc.Id).
		Updates(map[string]any{
			"name":     in.Name,
			"begin_at": time.UnixMilli(in.BeginAt),
			"end_at":   time.UnixMilli(in.EndAt),
		})
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "system error: " + result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "meeting not found or you don't have permission to edit",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
	})
}

func MeetingDelete(c *gin.Context) {
	identity := c.Query("identity")
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	err := models.DB.Model(new(models.RoomBasic)).Where("identify = ? AND create_id = ?", identity, uc.Id).Delete(&models.RoomBasic{}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "system error: " + err.Error(),
		})
		return
	}
}
