package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"GoMeetings/internal/helper"
	"GoMeetings/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// signalMessage represents payload exchanged through the signaling channel.
type signalMessage struct {
	UserIdentity   string          `json:"user_identity"`
	RoomIdentity   string          `json:"room_identity"`
	Key            string          `json:"key"`
	Value          json.RawMessage `json:"value"`
	TargetIdentity string          `json:"target_identity,omitempty"`
	Timestamp      int64           `json:"timestamp"`
	System         bool            `json:"system,omitempty"`
}

const (
	maxSignalPayloadSize = 64 * 1024 // 64KB
	pongWaitDuration     = 70 * time.Second
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if parsed, err := url.Parse(origin); err == nil && parsed.Host != "" {
			return true
		}
		return false
	},
}

type peerConn struct {
	conn    *websocket.Conn
	room    string
	user    string
	writeMu sync.Mutex
}

func (p *peerConn) sendBytes(payload []byte) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	return p.conn.WriteMessage(websocket.TextMessage, payload)
}

func (p *peerConn) readLoop(hub *signalHub) {
	defer func() {
		_ = p.conn.Close()
		hub.handlePeerLeave(p)
	}()

	for {
		_, raw, err := p.conn.ReadMessage()
		if err != nil {
			if !isExpectedClose(err) {
				log.Printf("signal: read error for %s: %v", p.user, err)
			}
			break
		}
		if len(raw) == 0 {
			continue
		}
		if len(raw) > maxSignalPayloadSize {
			log.Printf("signal: dropped oversized payload from %s (%d bytes)", p.user, len(raw))
			continue
		}
		hub.handleIncoming(p, raw)
	}
}

type signalHub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*peerConn
}

func newSignalHub() *signalHub {
	return &signalHub{
		rooms: make(map[string]map[string]*peerConn),
	}
}

var wsHub = newSignalHub()

// SignalWebsocket godoc
// @Summary WebRTC signaling websocket
// @Description Upgrade HTTP to WebSocket for SDP/ICE exchange
// @Tags Signaling
// @Param roomIdentity path string true "Room identity"
// @Param userIdentity path string true "User identity"
// @Param token query string true "JWT bearer token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /ws/p2p/{roomIdentity}/{userIdentity} [get]
func SignalWebsocket(c *gin.Context) {
	roomIdentity := c.Param("roomIdentity")
	userIdentity := c.Param("userIdentity")
	if roomIdentity == "" || userIdentity == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "roomIdentity and userIdentity are required",
		})
		return
	}

	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "token is required",
		})
		return
	}

	claims, err := helper.AnalyzeToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "invalid token",
		})
		return
	}

	var room models.RoomBasic
	if err := models.DB.Where("identify = ?", roomIdentity).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  "room not found",
		})
		return
	}

	if err := ensureRoomJoinWindow(&room, time.Now()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		return
	}

	var membership models.RoomUser
	if err := models.DB.Where("rid = ? AND uid = ?", room.ID, claims.Id).First(&membership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code": http.StatusForbidden,
			"msg":  "user is not a member of the room",
		})
		return
	}

	if !signalIdentityMatches(userIdentity, claims, &membership) {
		c.JSON(http.StatusForbidden, gin.H{
			"code": http.StatusForbidden,
			"msg":  "identity mismatch",
		})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("signal: websocket upgrade failed: %v", err)
		return
	}

	configureWebsocketConn(conn)
	handleSignalConn(conn, roomIdentity, userIdentity)
}

func handleSignalConn(conn *websocket.Conn, roomIdentity, userIdentity string) {
	peer, existingPeers, err := wsHub.join(roomIdentity, userIdentity, conn)
	if err != nil {
		payload := buildErrorPayload(roomIdentity, err.Error())
		_ = conn.WriteMessage(websocket.TextMessage, payload)
		_ = conn.Close()
		return
	}

	wsHub.sendPeerList(peer, existingPeers)
	wsHub.notifyPeerJoined(peer)
	peer.readLoop(wsHub)
}

func (h *signalHub) join(roomIdentity, userIdentity string, conn *websocket.Conn) (*peerConn, []string, error) {
	if roomIdentity == "" || userIdentity == "" {
		return nil, nil, errors.New("room or user identity is empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	roomPeers, ok := h.rooms[roomIdentity]
	if !ok {
		roomPeers = make(map[string]*peerConn)
		h.rooms[roomIdentity] = roomPeers
	}

	if _, exists := roomPeers[userIdentity]; exists {
		return nil, nil, fmt.Errorf("user %s already connected in room %s", userIdentity, roomIdentity)
	}

	peer := &peerConn{
		conn: conn,
		room: roomIdentity,
		user: userIdentity,
	}
	roomPeers[userIdentity] = peer

	existing := make([]string, 0, len(roomPeers)-1)
	for id := range roomPeers {
		if id == userIdentity {
			continue
		}
		existing = append(existing, id)
	}

	return peer, existing, nil
}

func (h *signalHub) handleIncoming(sender *peerConn, raw []byte) {
	var msg signalMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("signal: invalid json payload from %s: %v", sender.user, err)
		return
	}
	if strings.TrimSpace(msg.Key) == "" {
		return
	}

	h.forward(sender, &msg)
}

func (h *signalHub) forward(sender *peerConn, msg *signalMessage) {
	msg.UserIdentity = sender.user
	msg.RoomIdentity = sender.room
	msg.Timestamp = time.Now().UnixMilli()

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("signal: marshal payload error: %v", err)
		return
	}

	targets := h.selectTargets(sender.room, sender.user, msg.TargetIdentity)
	for _, peer := range targets {
		if err := peer.sendBytes(payload); err != nil {
			log.Printf("signal: send error to %s: %v", peer.user, err)
		}
	}
}

func (h *signalHub) selectTargets(roomIdentity, senderIdentity, targetIdentity string) []*peerConn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomPeers, ok := h.rooms[roomIdentity]
	if !ok {
		return nil
	}

	if targetIdentity != "" {
		if peer, ok := roomPeers[targetIdentity]; ok {
			return []*peerConn{peer}
		}
		return nil
	}

	targets := make([]*peerConn, 0, len(roomPeers)-1)
	for id, peer := range roomPeers {
		if id == senderIdentity {
			continue
		}
		targets = append(targets, peer)
	}
	return targets
}

func (h *signalHub) broadcast(roomIdentity string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomPeers, ok := h.rooms[roomIdentity]
	if !ok {
		return
	}
	for _, peer := range roomPeers {
		if err := peer.sendBytes(payload); err != nil {
			log.Printf("signal: broadcast error to %s: %v", peer.user, err)
		}
	}
}

func (h *signalHub) handlePeerLeave(peer *peerConn) {
	targets, removed := h.removePeer(peer.room, peer.user)
	if !removed {
		return
	}

	msg := signalMessage{
		UserIdentity: peer.user,
		RoomIdentity: peer.room,
		Key:          "peer_left",
		Value:        mustRawMessage(map[string]string{"user_identity": peer.user}),
		System:       true,
		Timestamp:    time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, target := range targets {
		if err := target.sendBytes(payload); err != nil {
			log.Printf("signal: notify leave error: %v", err)
		}
	}
}

func (h *signalHub) removePeer(roomIdentity, userIdentity string) ([]*peerConn, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomPeers, ok := h.rooms[roomIdentity]
	if !ok {
		return nil, false
	}
	if _, exists := roomPeers[userIdentity]; !exists {
		return nil, false
	}

	delete(roomPeers, userIdentity)

	targets := make([]*peerConn, 0, len(roomPeers))
	for _, peer := range roomPeers {
		targets = append(targets, peer)
	}

	if len(roomPeers) == 0 {
		delete(h.rooms, roomIdentity)
	}
	return targets, true
}

func (h *signalHub) sendPeerList(peer *peerConn, peers []string) {
	msg := signalMessage{
		UserIdentity: "system",
		RoomIdentity: peer.room,
		Key:          "peer_list",
		Value:        mustRawMessage(map[string][]string{"peers": peers}),
		System:       true,
		Timestamp:    time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	if err := peer.sendBytes(payload); err != nil {
		log.Printf("signal: send peer list error: %v", err)
	}
}

func (h *signalHub) notifyPeerJoined(peer *peerConn) {
	msg := signalMessage{
		UserIdentity: peer.user,
		RoomIdentity: peer.room,
		Key:          "peer_joined",
		Value:        mustRawMessage(map[string]string{"user_identity": peer.user}),
		System:       true,
		Timestamp:    time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	targets := h.selectTargets(peer.room, peer.user, "")
	for _, target := range targets {
		if err := target.sendBytes(payload); err != nil {
			log.Printf("signal: notify join error: %v", err)
		}
	}
}

func buildErrorPayload(roomIdentity, message string) []byte {
	msg := signalMessage{
		UserIdentity: "system",
		RoomIdentity: roomIdentity,
		Key:          "error",
		Value:        mustRawMessage(map[string]string{"message": message}),
		System:       true,
		Timestamp:    time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return []byte(`{"key":"error","value":{"message":"internal error"}}`)
	}
	return payload
}

func mustRawMessage(v interface{}) json.RawMessage {
	if v == nil {
		return json.RawMessage("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return json.RawMessage(b)
}

func isExpectedClose(err error) bool {
	if err == nil {
		return false
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if opErr, ok := err.(*net.OpError); ok {
		return !opErr.Timeout()
	}
	return strings.Contains(err.Error(), "closed network connection")
}

func configureWebsocketConn(conn *websocket.Conn) {
	conn.SetReadLimit(maxSignalPayloadSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWaitDuration))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWaitDuration))
	})
}

func signalIdentityMatches(identity string, claims *helper.UserClaims, membership *models.RoomUser) bool {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return false
	}
	expectedID := strconv.FormatUint(uint64(claims.Id), 10)
	if identity == expectedID {
		return true
	}
	if strings.EqualFold(identity, claims.Name) {
		return true
	}
	if membership != nil && identity == membership.DisplayName {
		return true
	}
	return false
}

func notifyScreenShareEvent(roomIdentity, key string, value interface{}) {
	msg := signalMessage{
		UserIdentity: "system",
		RoomIdentity: roomIdentity,
		Key:          key,
		Value:        mustRawMessage(value),
		System:       true,
		Timestamp:    time.Now().UnixMilli(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	wsHub.broadcast(roomIdentity, payload)
}
