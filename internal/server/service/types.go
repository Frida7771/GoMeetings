package service

type MeetingCreateRequest struct {
	Name    string `json:"Name"`
	BeginAt int64  `json:"Begin_at"`
	EndAt   int64  `json:"End_at"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
