package service

import (
	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	in := new(UserLoginRequest)
	err := c.ShouldBindJSON(in)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "request error",
		})
		return
	}
	if in.Username == "" || in.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "required message is empty",
		})
		return
	}
	password := helper.GetMd5(in.Password)

	data := new(models.UserBasic)
	err = models.DB.Where("username = ? and password = ?", in.Username, password).First(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "username or password error",
		})
		return
	}

	token, err := helper.GenerateToken(data.ID, data.Username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Generate Token Error" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
			"user":  data,
		},
	})
}
