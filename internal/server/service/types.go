package service

import (
	"GoMeetings/internal/models"
	"time"
)

type MeetingCreateRequest struct {
	Name    string `json:"Name"`
	BeginAt int64  `json:"Begin_at"`
	EndAt   int64  `json:"End_at"`
}

type UserLoginRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type UserRegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
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

type RoomCreateRequest struct {
	Name        string `json:"name" form:"name" binding:"required"`
	BeginAt     int64  `json:"begin_at" form:"begin_at" binding:"required"`
	EndAt       int64  `json:"end_at" form:"end_at" binding:"required"`
	JoinCode    string `json:"join_code" form:"join_code" binding:"omitempty"`
	ShortCode   string `json:"short_code" form:"short_code" binding:"omitempty"`
	DisplayName string `json:"display_name" form:"display_name" binding:"omitempty"`
}

type RoomEditRequest struct {
	Identify  string `json:"identity" form:"identity" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
	BeginAt   int64  `json:"begin_at" form:"begin_at" binding:"required"`
	EndAt     int64  `json:"end_at" form:"end_at" binding:"required"`
	JoinCode  string `json:"join_code" form:"join_code" binding:"omitempty"`
	ShortCode string `json:"short_code" form:"short_code" binding:"omitempty"`
}

type RoomListRequest struct {
	Page     int    `form:"page"`
	Size     int    `form:"size"`
	Keyword  string `form:"keyword"`
	Identity string `form:"identity"`
}

type UserRoomListRequest struct {
	UserIdentity string `form:"user_identity" binding:"required"`
	Page         int    `form:"page"`
	Size         int    `form:"size"`
	Keyword      string `form:"keyword"`
}

type RoomJoinRequest struct {
	Identity    string `json:"identity" form:"identity" binding:"required"`
	DisplayName string `json:"display_name" form:"display_name" binding:"required"`
	JoinCode    string `json:"join_code" form:"join_code" binding:"required"`
}

type RoomLeaveRequest struct {
	Identity string `json:"identity" form:"identity" binding:"required"`
}

type ScreenShareStartRequest struct {
	Identity string `json:"identity" form:"identity" binding:"required"`
	StreamID string `json:"stream_id" form:"stream_id" binding:"omitempty"`
}

type ScreenShareStopRequest struct {
	Identity string `json:"identity" form:"identity" binding:"required"`
}

type ScreenShareStatusReply struct {
	Active    bool   `json:"active"`
	OwnerID   uint   `json:"owner_id,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
	StreamID  string `json:"stream_id,omitempty"`
	StartedAt int64  `json:"started_at,omitempty"`
	EndedAt   int64  `json:"ended_at,omitempty"`
}

type RoomListItem struct {
	Identity string       `json:"identity"`
	Name     string       `json:"name"`
	BeginAt  time.Time    `json:"begin_at"`
	EndAt    time.Time    `json:"end_at"`
	CreateID uint         `json:"create_id"`
	Joined   bool         `json:"joined"`
	Members  []RoomMember `json:"members,omitempty"`
}

type RoomListReply struct {
	Total int64          `json:"total"`
	List  []RoomListItem `json:"list"`
}

type RoomMember struct {
	UserID      uint   `json:"user_id"`
	DisplayName string `json:"display_name"`
	JoinedAt    int64  `json:"joined_at"`
}

type RoomMembersReply struct {
	Identity string       `json:"identity"`
	Members  []RoomMember `json:"members"`
}
