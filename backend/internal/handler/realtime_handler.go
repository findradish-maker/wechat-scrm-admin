package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
	"wechat-enterprise-backend/internal/realtime"
	"wechat-enterprise-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type RealtimeHandler struct {
	accountService *service.AccountService
	hub            *realtime.Hub
	mu             sync.Mutex
	pollers        map[string]chan struct{}
	syncLocks      map[string]*sync.Mutex
	upgrader       websocket.Upgrader
}

type realtimeSender struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type realtimeMessage struct {
	Timestamp      string         `json:"timestamp"`
	Category       int64          `json:"category"`
	MsgID          int64          `json:"msgId"`
	ConversationID string         `json:"conversationId"`
	Sender         realtimeSender `json:"sender"`
	Content        string         `json:"content"`
}

func NewRealtimeHandler(accountService *service.AccountService, hub *realtime.Hub) *RealtimeHandler {
	return &RealtimeHandler{
		accountService: accountService,
		hub:            hub,
		pollers:        make(map[string]chan struct{}),
		syncLocks:      make(map[string]*sync.Mutex),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (h *RealtimeHandler) ServeWS(c *gin.Context) {
	wxid := strings.TrimSpace(c.Param("wxid"))
	if wxid == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.hub.Register(wxid, conn)
	h.ensurePolling(wxid)
	defer func() {
		h.hub.Unregister(wxid, conn)
		h.stopPollingIfIdle(wxid)
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Time{})
	conn.SetPongHandler(func(_ string) error {
		return conn.SetReadDeadline(time.Time{})
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *RealtimeHandler) SyncMessagePreview(c *gin.Context) {
	wxid := strings.TrimSpace(c.Param("wxid"))
	if wxid == "" {
		c.String(http.StatusBadRequest, "")
		return
	}

	result, err := h.accountService.SyncMessages(c.Request.Context(), wxid)
	if err != nil {
		c.String(http.StatusBadRequest, "")
		return
	}

	payload := formatRealtimeMessages(result.Messages)
	c.String(http.StatusOK, payload)
}

func (h *RealtimeHandler) SyncMessage(c *gin.Context) {
	wxid := strings.TrimSpace(c.Param("wxid"))
	if wxid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "wxid is required"})
		return
	}

	result, err := h.syncAndBroadcast(c.Request.Context(), wxid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"continueFlag": result.ContinueFlag,
		"message":      result.Message,
		"syncedCount":  result.SyncedCount,
	})
}

func (h *RealtimeHandler) ensurePolling(wxid string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.pollers[wxid]; ok {
		return
	}

	stopCh := make(chan struct{})
	h.pollers[wxid] = stopCh
	go h.runPolling(wxid, stopCh)
}

func (h *RealtimeHandler) stopPollingIfIdle(wxid string) {
	if h.hub.Count(wxid) > 0 {
		return
	}

	h.mu.Lock()
	stopCh, ok := h.pollers[wxid]
	if ok {
		delete(h.pollers, wxid)
	}
	h.mu.Unlock()

	if ok {
		close(stopCh)
	}
}

func (h *RealtimeHandler) runPolling(wxid string, stopCh chan struct{}) {
	trigger := make(chan struct{}, 1)
	trigger <- struct{}{}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			select {
			case trigger <- struct{}{}:
			default:
			}
		case <-trigger:
			if h.hub.Count(wxid) == 0 {
				h.stopPollingIfIdle(wxid)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			_, _ = h.syncAndBroadcast(ctx, wxid)
			cancel()
		}
	}
}

func (h *RealtimeHandler) syncAndBroadcast(ctx context.Context, wxid string) (*service.SyncMessagesResult, error) {
	lock := h.getSyncLock(wxid)
	lock.Lock()
	defer lock.Unlock()

	result, err := h.accountService.SyncMessages(ctx, wxid)
	if err != nil {
		return nil, err
	}

	payload := formatRealtimeMessages(result.Messages)
	if payload != "" {
		h.hub.Broadcast(wxid, []byte(payload))
	}
	return result, nil
}

func (h *RealtimeHandler) getSyncLock(wxid string) *sync.Mutex {
	h.mu.Lock()
	defer h.mu.Unlock()

	lock, ok := h.syncLocks[wxid]
	if ok {
		return lock
	}

	lock = &sync.Mutex{}
	h.syncLocks[wxid] = lock
	return lock
}

func formatRealtimeMessages(messages []service.MessageSummary) string {
	if len(messages) == 0 {
		return ""
	}

	lines := make([]string, 0, len(messages))
	for _, item := range messages {
		payload, err := json.Marshal(realtimeMessage{
			Timestamp:      formatRealtimeTimestamp(item.CreateTime),
			Category:       item.MsgType,
			MsgID:          item.MsgID,
			ConversationID: item.ChatWxid,
			Sender: realtimeSender{
				ID:       item.SenderWxid,
				Nickname: item.SenderDisplay,
			},
			Content: item.Content,
		})
		if err != nil {
			continue
		}
		lines = append(lines, string(payload))
	}
	return strings.Join(lines, "\n")
}

func formatRealtimeTimestamp(unixSeconds int64) string {
	if unixSeconds <= 0 {
		return "Unknown Time"
	}
	return time.Unix(unixSeconds, 0).
		In(time.Local).
		Format("2006-01-02 15:04:05")
}
