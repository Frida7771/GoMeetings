package service

import (
	"GoMeetings/internal/helper"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	println(helper.GetMd5("123456"))
}
