package service

import (
	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserLogin godoc
// @Summary User login
// @Description Authenticate user and returns JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body UserLoginRequest true "Login payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/user/login [post]
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

// UserRegister godoc
// @Summary User register
// @Description Create a user account and return JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body UserRegisterRequest true "Register payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/user/register [post]
func UserRegister(c *gin.Context) {
	var req UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "params error: " + err.Error(),
		})
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if len(username) < 3 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "username must be at least 3 characters"})
		return
	}
	if len(password) < 6 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "password must be at least 6 characters"})
		return
	}

	var existing models.UserBasic
	if err := models.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "username already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	user := models.UserBasic{
		Username: username,
		Password: helper.GetMd5(password),
	}
	if err := models.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	token, err := helper.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "Generate Token Error" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
			"user":  user,
		},
	})
}
