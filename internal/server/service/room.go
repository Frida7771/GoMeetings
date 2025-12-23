package service

import (
	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	defaultJoinCodeLength   = 6
	defaultShortCodeLength  = 6
	codeAlphabet            = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	roomEarlyJoinWindow     = 15 * time.Minute
	maxCodeCollisionRetries = 5
)

// @Summary Generate a random code
// @Description Generates a random code of the given length using the codeAlphabet
// @Tags Room
// @Accept json
// @Produce json
// @Param length query int false "Code length"
// @Success 200 {string} string "Random code"
// @Failure 500 {string} string "Internal server error"
// @Router /room/generate-code [get]
func generateCode(length int) string {
	if length <= 0 {
		length = defaultJoinCodeLength
	}
	var sb strings.Builder
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			sb.WriteByte(codeAlphabet[i%len(codeAlphabet)])
			continue
		}
		sb.WriteByte(codeAlphabet[num.Int64()])
	}
	return sb.String()
}

// RoomList godoc
// @Summary Room list
// @Description Paginated room list with join state
// @Tags Room
// @Produce json
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword filter"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/room/list [get]
func RoomList(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	req := RoomListRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	var userRooms []models.RoomUser
	if err := models.DB.Where("uid = ?", uc.Id).Find(&userRooms).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}
	joined := make(map[uint]bool, len(userRooms))
	for _, ur := range userRooms {
		joined[ur.Rid] = true
	}

	roomMembersReply(c, uc.Id, req.Identity)
}

func roomMembersReply(c *gin.Context, uid uint, identity string) {
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "identity is required"})
		return
	}
	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}

	var membership models.RoomUser
	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, uid).First(&membership).Error; err != nil {
		if room.CreateID != uid {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission"})
			return
		}
	}

	var members []models.RoomUser
	if err := models.DB.Where("rid = ?", room.ID).Find(&members).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	memberList := make([]RoomMember, 0, len(members))
	for _, m := range members {
		memberList = append(memberList, RoomMember{
			UserID:      m.Uid,
			DisplayName: m.DisplayName,
			JoinedAt:    m.CreatedAt.UnixMilli(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": RoomListReply{
			Total: int64(len(memberList)),
			List: []RoomListItem{
				{
					Identity: room.Identify,
					Name:     room.Name,
					BeginAt:  room.BeginAt,
					EndAt:    room.EndAt,
					CreateID: room.CreateID,
					Joined:   true,
					Members:  memberList,
				},
			},
		},
	})
}

// RoomUserRooms godoc
// @Summary User related rooms
// @Description List meetings created or joined by the specified user identity
// @Tags Room
// @Security BearerAuth
// @Produce json
// @Param user_identity query string true "User identity (username or numeric ID)"
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param keyword query string false "Keyword filter"
// @Success 200 {object} map[string]interface{}
// @Router /auth/room/user-rooms [get]
func RoomUserRooms(c *gin.Context) {
	req := UserRoomListRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	targetID, err := resolveUserIdentity(req.UserIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "user not found"})
		return
	}

	var userRooms []models.RoomUser
	if err := models.DB.Where("uid = ?", targetID).Find(&userRooms).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}
	joined := make(map[uint]bool, len(userRooms))
	roomIDs := make([]uint, 0, len(userRooms))
	for _, ur := range userRooms {
		joined[ur.Rid] = true
		roomIDs = append(roomIDs, ur.Rid)
	}

	query := models.DB.Model(&models.RoomBasic{}).Where("create_id = ?", targetID)
	if len(roomIDs) > 0 {
		query = query.Or("id IN ?", roomIDs)
	}
	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	var rooms []models.RoomBasic
	if err := query.Order("created_at desc").
		Limit(req.Size).Offset((req.Page - 1) * req.Size).Find(&rooms).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	roomIDList := make([]uint, 0, len(rooms))
	for _, room := range rooms {
		roomIDList = append(roomIDList, room.ID)
	}
	memberMap, err := loadMembersForRooms(roomIDList)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	list := make([]RoomListItem, 0, len(rooms))
	for _, room := range rooms {
		list = append(list, RoomListItem{
			Identity: room.Identify,
			Name:     room.Name,
			BeginAt:  room.BeginAt,
			EndAt:    room.EndAt,
			CreateID: room.CreateID,
			Joined:   joined[room.ID] || room.CreateID == targetID,
			Members:  memberMap[room.ID],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": RoomListReply{
			Total: total,
			List:  list,
		},
	})
}

// RoomCreate godoc
// @Summary Create room
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Room name"
// @Param begin_at formData integer true "Begin time (ms)"
// @Param end_at formData integer true "End time (ms)"
// @Param join_code formData string false "Custom join code"
// @Param short_code formData string false "Short code"
// @Param display_name formData string false "Owner display name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/room/create [post]
func RoomCreate(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	req := RoomCreateRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}
	if req.EndAt <= req.BeginAt {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "end time must be greater than begin time"})
		return
	}

	joinCode, err := ensureUniqueJoinCode(req.JoinCode, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "unable to allocate join code: " + err.Error()})
		return
	}
	if len(joinCode) < 4 || len(joinCode) > 16 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "join code length must be between 4 and 16"})
		return
	}
	shortCode, err := ensureUniqueShortCode(req.ShortCode, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "unable to allocate short code: " + err.Error()})
		return
	}

	room := models.RoomBasic{
		Identify:  helper.GenerateUUID(),
		Name:      req.Name,
		BeginAt:   time.UnixMilli(req.BeginAt),
		EndAt:     time.UnixMilli(req.EndAt),
		CreateID:  uc.Id,
		JoinCode:  joinCode,
		ShortCode: shortCode,
	}
	if err := models.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	ownerName := req.DisplayName
	if ownerName == "" {
		ownerName = uc.Name
		if ownerName == "" {
			ownerName = "owner"
		}
	}
	_ = models.DB.Where("rid = ? AND uid = ?", room.ID, uc.Id).Assign(models.RoomUser{
		DisplayName: ownerName,
	}).FirstOrCreate(&models.RoomUser{
		Rid:         room.ID,
		Uid:         uc.Id,
		DisplayName: ownerName,
	})

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": room})
}

// RoomEdit godoc
// @Summary Edit room
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param identity formData string true "Room identity"
// @Param name formData string true "Room name"
// @Param begin_at formData integer true "Begin time (ms)"
// @Param end_at formData integer true "End time (ms)"
// @Param join_code formData string false "Custom join code"
// @Param short_code formData string false "Short code"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/room/edit [put]
func RoomEdit(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	req := RoomEditRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}
	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", req.Identify).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}
	if room.CreateID != uc.Id {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission"})
		return
	}

	update := map[string]any{
		"name":     req.Name,
		"begin_at": time.UnixMilli(req.BeginAt),
		"end_at":   time.UnixMilli(req.EndAt),
	}
	if req.JoinCode != "" {
		code, err := ensureUniqueJoinCode(req.JoinCode, room.ID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "unable to allocate join code: " + err.Error()})
			return
		}
		if len(code) < 4 || len(code) > 16 {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "join code length must be between 4 and 16"})
			return
		}
		update["join_code"] = code
		room.JoinCode = code
	}
	if req.ShortCode != "" {
		code, err := ensureUniqueShortCode(req.ShortCode, room.ID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "unable to allocate short code: " + err.Error()})
			return
		}
		update["short_code"] = code
		room.ShortCode = code
	}
	if err := models.DB.Model(&room).Updates(update).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}
	room.Name = req.Name
	room.BeginAt = time.UnixMilli(req.BeginAt)
	room.EndAt = time.UnixMilli(req.EndAt)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": room})
}

// RoomDelete godoc
// @Summary Delete room
// @Tags Room
// @Security BearerAuth
// @Produce json
// @Param identity query string true "Room identity"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/room/delete [delete]
func RoomDelete(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "identity is required"})
		return
	}

	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}
	if room.CreateID != uc.Id {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission"})
		return
	}

	if err := models.DB.Delete(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}
	models.DB.Where("rid = ?", room.ID).Delete(&models.RoomUser{})

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "delete success"})
}

// RoomJoin godoc
// @Summary Join room
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param identity formData string true "Room identity"
// @Param display_name formData string true "Display name"
// @Param join_code formData string true "Join code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/room/join [post]
func RoomJoin(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	req := RoomJoinRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}

	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", req.Identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}
	if err := ensureRoomJoinWindow(&room, time.Now()); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}
	if strings.ToUpper(req.JoinCode) != room.JoinCode {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "join code incorrect"})
		return
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "display name is required"})
		return
	}

	var roomUser models.RoomUser
	result := models.DB.Where("rid = ? AND uid = ?", room.ID, uc.Id).First(&roomUser)
	if result.Error == nil {
		if displayName != roomUser.DisplayName {
			models.DB.Model(&roomUser).Update("display_name", displayName)
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "already joined"})
		return
	}

	if err := models.DB.Create(&models.RoomUser{
		Rid:         room.ID,
		Uid:         uc.Id,
		DisplayName: displayName,
	}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "join success"})
}

// RoomLeave godoc
// @Summary Leave room
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param identity formData string true "Room identity"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/room/leave [post]
func RoomLeave(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	req := RoomLeaveRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}

	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", req.Identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}

	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, uc.Id).
		Delete(&models.RoomUser{}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	stopScreenShareForUser(&room, uc.Id, "left_room")

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "leave success"})
}

// RoomMembers godoc
// @Summary Room member list
// @Tags Room
// @Security BearerAuth
// @Produce json
// @Param identity query string true "Room identity"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/room/members [get]
func RoomMembers(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "identity is required"})
		return
	}
	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return
	}

	var membership models.RoomUser
	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, uc.Id).First(&membership).Error; err != nil {
		if room.CreateID != uc.Id {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission"})
			return
		}
	}

	var members []models.RoomUser
	if err := models.DB.Where("rid = ?", room.ID).Find(&members).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}
	list := make([]RoomMember, 0, len(members))
	for _, m := range members {
		list = append(list, RoomMember{
			UserID:      m.Uid,
			DisplayName: m.DisplayName,
			JoinedAt:    m.CreatedAt.UnixMilli(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": RoomMembersReply{
			Identity: room.Identify,
			Members:  list,
		},
	})
}

func ensureUniqueJoinCode(candidate string, excludeID uint) (string, error) {
	return ensureUniqueRoomCode("join_code", candidate, defaultJoinCodeLength, excludeID)
}

func ensureUniqueShortCode(candidate string, excludeID uint) (string, error) {
	return ensureUniqueRoomCode("short_code", candidate, defaultShortCodeLength, excludeID)
}

func ensureUniqueRoomCode(column, candidate string, length int, excludeID uint) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(candidate))
	if code == "" {
		code = generateCode(length)
	}

	for attempt := 0; attempt < maxCodeCollisionRetries; attempt++ {
		var count int64
		query := models.DB.Model(&models.RoomBasic{}).Where(column+" = ?", code)
		if excludeID != 0 {
			query = query.Where("id <> ?", excludeID)
		}
		if err := query.Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
		code = generateCode(length + attempt + 1)
	}
	return "", fmt.Errorf("unable to generate unique %s", column)
}

func ensureRoomJoinWindow(room *models.RoomBasic, now time.Time) error {
	if now.After(room.EndAt) {
		return errors.New("meeting has already ended")
	}
	if now.Before(room.BeginAt.Add(-roomEarlyJoinWindow)) {
		return errors.New("meeting is not open for participants yet")
	}
	return nil
}

func resolveUserIdentity(identity string) (uint, error) {
	var user models.UserBasic
	if id, err := strconv.ParseUint(identity, 10, 64); err == nil {
		if err := models.DB.First(&user, id).Error; err != nil {
			return 0, err
		}
		return user.ID, nil
	}
	if err := models.DB.Where("username = ?", identity).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func loadMembersForRooms(roomIDs []uint) (map[uint][]RoomMember, error) {
	result := make(map[uint][]RoomMember)
	if len(roomIDs) == 0 {
		return result, nil
	}
	var roomMembers []models.RoomUser
	if err := models.DB.Where("rid IN ?", roomIDs).Find(&roomMembers).Error; err != nil {
		return nil, err
	}
	for _, m := range roomMembers {
		result[m.Rid] = append(result[m.Rid], RoomMember{
			UserID:      m.Uid,
			DisplayName: m.DisplayName,
			JoinedAt:    m.CreatedAt.UnixMilli(),
		})
	}
	return result, nil
}

// RoomShareStart godoc
// @Summary Start screen sharing
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param identity formData string true "Room identity"
// @Param stream_id formData string false "Client-defined stream ID"
// @Success 200 {object} map[string]string
// @Router /auth/room/share/start [post]
func RoomShareStart(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	var req ScreenShareStartRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}

	room, membership, ok := loadRoomAndMembership(c, uc.Id, req.Identity)
	if !ok {
		return
	}
	ownerName := resolveDisplayName(membership, uc.Name)

	if err := ensureRoomJoinWindow(room, time.Now()); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	var existing models.RoomScreenShare
	err := models.DB.Where("rid = ? AND active = ?", room.ID, true).First(&existing).Error
	if err == nil {
		if existing.OwnerUid != uc.Id {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "another participant is already sharing"})
			return
		}
		started := time.Now().UnixMilli()
		updates := map[string]interface{}{
			"stream_id":  req.StreamID,
			"started_at": started,
			"ended_at":   nil,
			"active":     true,
		}
		if err := models.DB.Model(&existing).Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
			return
		}
		notifyScreenShareEvent(room.Identify, "screen_share_refreshed", map[string]interface{}{
			"owner_id":   uc.Id,
			"owner_name": ownerName,
			"stream_id":  req.StreamID,
			"started_at": started,
		})
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "screen share refreshed"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	startedAt := time.Now().UnixMilli()
	share := models.RoomScreenShare{
		Rid:       room.ID,
		OwnerUid:  uc.Id,
		StreamID:  req.StreamID,
		Active:    true,
		StartedAt: startedAt,
	}
	if err := models.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	notifyScreenShareEvent(room.Identify, "screen_share_started", map[string]interface{}{
		"owner_id":   uc.Id,
		"owner_name": ownerName,
		"stream_id":  req.StreamID,
		"started_at": startedAt,
	})

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "screen share started"})
}

// RoomShareStop godoc
// @Summary Stop screen sharing
// @Tags Room
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param identity formData string true "Room identity"
// @Success 200 {object} map[string]string
// @Router /auth/room/share/stop [post]
func RoomShareStop(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	var req ScreenShareStopRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "params error: " + err.Error()})
		return
	}

	room, _, ok := loadRoomAndMembership(c, uc.Id, req.Identity)
	if !ok {
		return
	}

	var share models.RoomScreenShare
	if err := models.DB.Where("rid = ? AND active = ?", room.ID, true).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "no active screen share"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	if share.OwnerUid != uc.Id && room.CreateID != uc.Id {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission to stop screen share"})
		return
	}

	if err := deactivateScreenShare(&share); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	notifyScreenShareEvent(room.Identify, "screen_share_stopped", map[string]interface{}{
		"owner_id": share.OwnerUid,
		"ended_at": share.EndedAt,
	})

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "screen share stopped"})
}

// RoomShareStatus godoc
// @Summary Screen share status
// @Tags Room
// @Security BearerAuth
// @Produce json
// @Param identity query string true "Room identity"
// @Success 200 {object} map[string]interface{}
// @Router /auth/room/share/status [get]
func RoomShareStatus(c *gin.Context) {
	uc := c.MustGet("user_claims").(*helper.UserClaims)
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "identity is required"})
		return
	}

	room, _, ok := loadRoomAndMembership(c, uc.Id, identity)
	if !ok {
		return
	}

	var share models.RoomScreenShare
	if err := models.DB.Where("rid = ? AND active = ?", room.ID, true).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": ScreenShareStatusReply{Active: false}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "system error: " + err.Error()})
		return
	}

	var owner models.RoomUser
	var ownerPtr *models.RoomUser
	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, share.OwnerUid).First(&owner).Error; err == nil {
		ownerPtr = &owner
	}
	resp := ScreenShareStatusReply{
		Active:    true,
		OwnerID:   share.OwnerUid,
		OwnerName: resolveDisplayName(ownerPtr, ""),
		StreamID:  share.StreamID,
		StartedAt: share.StartedAt,
	}
	if share.EndedAt != nil {
		resp.EndedAt = *share.EndedAt
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}

func loadRoomAndMembership(c *gin.Context, uid uint, identity string) (*models.RoomBasic, *models.RoomUser, bool) {
	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", identity).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "room not found"})
		return nil, nil, false
	}
	var membership models.RoomUser
	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, uid).First(&membership).Error; err != nil {
		if room.CreateID != uid {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "no permission"})
			return nil, nil, false
		}
		return &room, nil, true
	}
	return &room, &membership, true
}

func deactivateScreenShare(share *models.RoomScreenShare) error {
	now := time.Now().UnixMilli()
	share.Active = false
	share.EndedAt = &now
	update := map[string]interface{}{
		"active":   false,
		"ended_at": now,
	}
	return models.DB.Model(share).Updates(update).Error
}

func stopScreenShareForUser(room *models.RoomBasic, uid uint, reason string) {
	var share models.RoomScreenShare
	if err := models.DB.Where("rid = ? AND owner_uid = ? AND active = ?", room.ID, uid, true).First(&share).Error; err != nil {
		return
	}
	if err := deactivateScreenShare(&share); err != nil {
		return
	}
	notifyScreenShareEvent(room.Identify, "screen_share_stopped", map[string]interface{}{
		"owner_id": uid,
		"reason":   reason,
		"ended_at": share.EndedAt,
	})
}

func resolveDisplayName(m *models.RoomUser, fallback string) string {
	if m != nil && strings.TrimSpace(m.DisplayName) != "" {
		return m.DisplayName
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return "host"
}
