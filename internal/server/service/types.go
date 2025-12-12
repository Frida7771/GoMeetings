package service

import "GoMeetings/internal/models"

type MeetingCreateRequest struct {
	Name    string `json:"Name"`
	BeginAt int64  `json:"Begin_at"`
	EndAt   int64  `json:"End_at"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MeetingEditRequest struct {
	Identify string `json:"identity"` // Support lowercase "identity" from request
	*MeetingCreateRequest
}

type MeetingListRequest struct {
	Page    int    `json:"Page" form:"page"`
	Size    int    `json:"Size" form:"size"`
	Keyword string `json:"Keyword" form:"keyword"`
}

type MeetingListReply struct {
	Total int64              `json:"Total"`
	List  []models.RoomBasic `json:"List"`
}
