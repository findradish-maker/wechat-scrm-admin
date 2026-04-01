package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"wechat-enterprise-backend/internal/config"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/internal/wechat"

	"gorm.io/gorm"
)

type AccountService struct {
	db              *gorm.DB
	wechatClient    *wechat.Client
	aiProviders     map[string]aiProvider
	defaultPlatform string
}

type CreateLoginSessionInput struct {
	Platform   string `json:"platform"`
	DeviceName string `json:"deviceName"`
}

type LoginSessionState struct {
	SessionID string                 `json:"sessionId"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	QRBase64  string                 `json:"qrBase64"`
	QRURL     string                 `json:"qrUrl"`
	UUID      string                 `json:"uuid"`
	ExpiresAt *time.Time             `json:"expiresAt"`
	Upstream  map[string]interface{} `json:"upstream"`
	Account   *domain.WechatAccount  `json:"account,omitempty"`
}

type BootstrapResult struct {
	Account         domain.WechatAccount `json:"account"`
	HeartbeatStatus string               `json:"heartbeatStatus"`
}

type ContactSummary struct {
	Wxid                    string     `json:"wxid"`
	DisplayName             string     `json:"displayName"`
	Nickname                string     `json:"nickname"`
	Alias                   string     `json:"alias"`
	Remark                  string     `json:"remark"`
	SortKey                 string     `json:"sortKey"`
	SortLetter              string     `json:"sortLetter"`
	Avatar                  string     `json:"avatar"`
	Signature               string     `json:"signature"`
	Province                string     `json:"province"`
	City                    string     `json:"city"`
	ContactType             string     `json:"contactType"`
	ContactCategory         string     `json:"contactCategory"`
	ContactCategoryLabel    string     `json:"contactCategoryLabel"`
	VerifyFlag              int64      `json:"verifyFlag"`
	MemberCount             int        `json:"memberCount"`
	ChatRoomOwner           string     `json:"chatRoomOwner"`
	Announcement            string     `json:"announcement"`
	AnnouncementPublishTime *time.Time `json:"announcementPublishTime,omitempty"`
	LastSyncedAt            *time.Time `json:"lastSyncedAt,omitempty"`
}

type ContactListResult struct {
	CurrentWxcontactSeq       int64            `json:"currentWxcontactSeq"`
	CurrentChatRoomContactSeq int64            `json:"currentChatRoomContactSeq"`
	ContinueFlag              int64            `json:"continueFlag"`
	Page                      int              `json:"page"`
	PageSize                  int              `json:"pageSize"`
	Total                     int64            `json:"total"`
	TotalPages                int              `json:"totalPages"`
	Contacts                  []ContactSummary `json:"contacts"`
}

type MessageSummary struct {
	MsgID            int64           `json:"msgId"`
	NewMsgID         int64           `json:"newMsgId"`
	MsgType          int64           `json:"msgType"`
	FromWxid         string          `json:"fromWxid"`
	ToWxid           string          `json:"toWxid"`
	ChatWxid         string          `json:"chatWxid"`
	ChatDisplay      string          `json:"chatDisplay"`
	ConversationType string          `json:"conversationType"`
	SenderWxid       string          `json:"senderWxid"`
	SenderDisplay    string          `json:"senderDisplay"`
	Content          string          `json:"content"`
	Preview          string          `json:"preview"`
	Kind             string          `json:"kind"`
	CreateTime       int64           `json:"createTime"`
	MsgSeq           int64           `json:"msgSeq"`
	IsSelf           bool            `json:"isSelf"`
	Article          *MessageArticle `json:"article,omitempty"`
	Quote            *MessageQuote   `json:"quote,omitempty"`
	Voice            *MessageVoice   `json:"voice,omitempty"`
	Video            *MessageVideo   `json:"video,omitempty"`
	Image            *MessageImage   `json:"image,omitempty"`
	Card             *MessageCard    `json:"card,omitempty"`
	Emoji            *MessageEmoji   `json:"emoji,omitempty"`
	System           *MessageSystem  `json:"system,omitempty"`
	ParseStatus      string          `json:"parseStatus,omitempty"`
	ParseError       string          `json:"parseError,omitempty"`
	DecodeXML        string          `json:"decodeXml,omitempty"`
}

type SyncMessagesResult struct {
	ContinueFlag int64            `json:"continueFlag"`
	Status       int64            `json:"status"`
	KeyBuf       string           `json:"keyBuf"`
	Message      string           `json:"message"`
	SyncedCount  int              `json:"syncedCount"`
	Messages     []MessageSummary `json:"messages"`
}

type MessageListResult struct {
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"totalPages"`
	Messages   []MessageSummary `json:"messages"`
}

type DashboardOverview struct {
	Summary  DashboardSummary      `json:"summary"`
	Accounts []DashboardAccountRow `json:"accounts"`
}

type DashboardSummary struct {
	TotalAccounts         int64 `json:"totalAccounts"`
	OnlineAccounts        int64 `json:"onlineAccounts"`
	OfflineAccounts       int64 `json:"offlineAccounts"`
	AccountsWithSync      int64 `json:"accountsWithSync"`
	AccountsActive24Hours int64 `json:"accountsActive24Hours"`
	TotalContacts         int64 `json:"totalContacts"`
	DirectContacts        int64 `json:"directContacts"`
	OfficialAccounts      int64 `json:"officialAccounts"`
	Groups                int64 `json:"groups"`
	TotalMessages         int64 `json:"totalMessages"`
	Messages24Hours       int64 `json:"messages24Hours"`
}

type DashboardAccountRow struct {
	Wxid                 string     `json:"wxid"`
	Nickname             string     `json:"nickname"`
	Avatar               string     `json:"avatar"`
	Status               string     `json:"status"`
	LastHeartbeatAt      *time.Time `json:"lastHeartbeatAt"`
	LastSyncAt           *time.Time `json:"lastSyncAt"`
	DirectContactCount   int64      `json:"directContactCount"`
	OfficialAccountCount int64      `json:"officialAccountCount"`
	GroupCount           int64      `json:"groupCount"`
	MessageCount         int64      `json:"messageCount"`
}

type SendTextInput struct {
	ToWxid  string `json:"toWxid"`
	Content string `json:"content"`
}

type SendTextResult struct {
	ToWxid      string `json:"toWxid"`
	ClientMsgID int64  `json:"clientMsgId"`
	CreateTime  int64  `json:"createTime"`
	NewMsgID    int64  `json:"newMsgId"`
	Ret         int64  `json:"ret"`
}

func NewAccountService(db *gorm.DB, wechatClient *wechat.Client, defaultPlatform string, aiConfig config.AIConfig) *AccountService {
	return &AccountService{
		db:              db,
		wechatClient:    wechatClient,
		aiProviders:     buildAIProviders(aiConfig),
		defaultPlatform: defaultPlatform,
	}
}

func (s *AccountService) ListAccounts(ctx context.Context) ([]domain.WechatAccount, error) {
	var accounts []domain.WechatAccount
	if err := s.db.WithContext(ctx).Order("updated_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	for index := range accounts {
		account := &accounts[index]
		if strings.ToLower(strings.TrimSpace(account.Status)) != "online" {
			account.Status = "offline"
			continue
		}

		envelope, cacheInfo, raw, err := s.wechatClient.GetCacheInfo(ctx, account.Wxid)
		if err != nil {
			account.Status = "offline"
			account.LastCachePayload = firstNonEmpty(string(raw), account.LastCachePayload)
			_ = s.db.WithContext(ctx).Model(&domain.WechatAccount{}).
				Where("wxid = ?", account.Wxid).
				Updates(map[string]interface{}{
					"status":             account.Status,
					"last_cache_payload": account.LastCachePayload,
				}).Error
			continue
		}
		if s.handleOfflineEnvelope(ctx, account.Wxid, envelope) {
			account.Status = "offline"
			account.LastCachePayload = firstNonEmpty(string(raw), account.LastCachePayload)
			continue
		}
		if envelope.Success && (envelope.Code == 0 || envelope.Code == 1) {
			now := time.Now()
			account.Status = "online"
			account.Avatar = firstNonEmpty(cacheInfo.HeadURL, account.Avatar)
			account.Nickname = firstNonEmpty(cacheInfo.Nickname, account.Nickname)
			account.Alias = firstNonEmpty(cacheInfo.Alias, account.Alias)
			account.Mobile = firstNonEmpty(cacheInfo.Mobile, account.Mobile)
			account.DeviceID = firstNonEmpty(cacheInfo.DeviceID, account.DeviceID)
			account.DeviceName = firstNonEmpty(cacheInfo.DeviceName, account.DeviceName)
			account.LastHeartbeatAt = &now
			account.LastCachePayload = string(raw)
			_ = s.db.WithContext(ctx).Save(account).Error
			continue
		}

		account.Status = "offline"
		account.LastCachePayload = firstNonEmpty(string(raw), account.LastCachePayload)
		_ = s.db.WithContext(ctx).Model(&domain.WechatAccount{}).
			Where("wxid = ?", account.Wxid).
			Updates(map[string]interface{}{
				"status":             account.Status,
				"last_cache_payload": account.LastCachePayload,
			}).Error
	}
	return accounts, nil
}

func (s *AccountService) CreateLoginSession(ctx context.Context, input CreateLoginSessionInput) (*LoginSessionState, error) {
	platform := strings.TrimSpace(input.Platform)
	if platform == "" {
		platform = s.defaultPlatform
	}
	deviceName := strings.TrimSpace(input.DeviceName)
	if deviceName == "" {
		deviceName = "Enterprise Console"
	}

	envelope, payload, _, err := s.wechatClient.CreateLoginQR(ctx, platform, deviceName)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, "", envelope)
		return nil, errors.New(envelope.Message)
	}

	var expiresAt *time.Time
	if payload.ExpiredTime != "" {
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", payload.ExpiredTime, time.Local); err == nil {
			expiresAt = &parsed
		}
	}

	session := domain.LoginSession{
		SessionID:   buildSessionID(),
		Platform:    platform,
		DeviceName:  deviceName,
		DeviceID:    envelope.DeviceID,
		UUID:        payload.UUID,
		QRBase64:    payload.QRBase64,
		QRURL:       payload.QRURL,
		Data62:      envelope.Data62,
		Status:      "pending",
		ExpiresAt:   expiresAt,
		LastCode:    envelope.Code,
		LastMessage: envelope.Message,
	}
	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, err
	}

	return &LoginSessionState{
		SessionID: session.SessionID,
		Status:    session.Status,
		Message:   envelope.Message,
		QRBase64:  session.QRBase64,
		QRURL:     session.QRURL,
		UUID:      session.UUID,
		ExpiresAt: session.ExpiresAt,
		Upstream: map[string]interface{}{
			"code":    envelope.Code,
			"message": envelope.Message,
		},
	}, nil
}

func (s *AccountService) CreateAwakenLoginSession(ctx context.Context, wxid string) (*LoginSessionState, error) {
	wxid = strings.TrimSpace(wxid)
	if wxid == "" {
		return nil, errors.New("账号不能为空")
	}

	var account domain.WechatAccount
	if err := s.db.WithContext(ctx).Where("wxid = ?", wxid).First(&account).Error; err != nil {
		return nil, err
	}

	envelope, payload, _, err := s.wechatClient.AwakenLogin(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	expireAt := time.Now().Add(time.Duration(payload.ExpiredTime) * time.Second)
	session := domain.LoginSession{
		SessionID:   buildSessionID(),
		Platform:    firstNonEmpty(account.Platform, s.defaultPlatform),
		DeviceID:    account.DeviceID,
		DeviceName:  firstNonEmpty(account.DeviceName, "Enterprise Console"),
		UUID:        payload.Uuid,
		QRURL:       buildQRCodeURL(payload.Uuid),
		Status:      "awaken_pending",
		Wxid:        wxid,
		ExpiresAt:   &expireAt,
		LastCode:    envelope.Code,
		LastMessage: envelope.Message,
	}
	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, err
	}

	return &LoginSessionState{
		SessionID: session.SessionID,
		Status:    session.Status,
		Message:   envelope.Message,
		QRURL:     session.QRURL,
		UUID:      session.UUID,
		ExpiresAt: session.ExpiresAt,
		Upstream: map[string]interface{}{
			"code":       envelope.Code,
			"message":    envelope.Message,
			"checkTime":  payload.CheckTime,
			"expiredSec": payload.ExpiredTime,
			"wxid":       wxid,
		},
	}, nil
}

func (s *AccountService) PollLoginSession(ctx context.Context, sessionID string) (*LoginSessionState, error) {
	var session domain.LoginSession
	if err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}

	envelope, raw, err := s.wechatClient.CheckLoginQR(ctx, session.UUID)
	if err != nil {
		return nil, err
	}
	session.LastRawResponse = string(raw)
	session.LastCode = envelope.Code
	session.LastMessage = envelope.Message

	upstream := map[string]interface{}{
		"code":    envelope.Code,
		"message": envelope.Message,
	}
	var account *domain.WechatAccount

	if envelope.Success {
		state := decodeLoginStatus(envelope.Data)
		if state != nil {
			upstream["status"] = state.Status
			upstream["expiredTime"] = state.ExpiredTime
		}
		resolvedWxid := s.extractWXID(envelope.Data)
		if preview := extractLoginPreview(envelope.Data); len(preview) > 0 {
			upstream["accountPreview"] = preview
		}
		if resolvedWxid != "" {
			session.Wxid = resolvedWxid
			session.Status = "authenticated"
			boot, err := s.bootstrapAccount(ctx, resolvedWxid, session.Platform)
			if err == nil {
				session.Status = "ready"
				account = &boot.Account
				upstream["heartbeatStatus"] = boot.HeartbeatStatus
			} else {
				upstream["bootstrapError"] = err.Error()
			}
		} else {
			session.Status = mapPollingStatus(state)
		}
	} else {
		s.handleOfflineEnvelope(ctx, session.Wxid, envelope)
		session.Status = "failed"
	}

	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		return nil, err
	}

	return &LoginSessionState{
		SessionID: session.SessionID,
		Status:    session.Status,
		Message:   session.LastMessage,
		QRBase64:  session.QRBase64,
		QRURL:     session.QRURL,
		UUID:      session.UUID,
		ExpiresAt: session.ExpiresAt,
		Upstream:  upstream,
		Account:   account,
	}, nil
}

func (s *AccountService) BootstrapAccount(ctx context.Context, wxid string) (*BootstrapResult, error) {
	result, err := s.bootstrapAccount(ctx, wxid, "")
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AccountService) StartHeartbeat(ctx context.Context, wxid string) (string, error) {
	envelope, _, err := s.wechatClient.StartHeartbeat(ctx, wxid)
	if err != nil {
		return "", err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return "", errors.New(envelope.Message)
	}
	if err := s.db.WithContext(ctx).Model(&domain.WechatAccount{}).Where("wxid = ?", wxid).Updates(map[string]interface{}{
		"status":            "online",
		"last_heartbeat_at": time.Now(),
	}).Error; err != nil {
		return "", err
	}
	return envelope.Message, nil
}

func (s *AccountService) StopHeartbeat(ctx context.Context, wxid string) (string, error) {
	envelope, _, err := s.wechatClient.StopHeartbeat(ctx, wxid)
	if err != nil {
		return "", err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return "", errors.New(envelope.Message)
	}
	if err := s.db.WithContext(ctx).Model(&domain.WechatAccount{}).Where("wxid = ?", wxid).Update("status", "offline").Error; err != nil {
		return "", err
	}
	return envelope.Message, nil
}

func (s *AccountService) Logout(ctx context.Context, wxid string) (string, error) {
	envelope, _, err := s.wechatClient.Logout(ctx, wxid)
	if err != nil {
		return "", err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return "", errors.New(envelope.Message)
	}
	if err := s.db.WithContext(ctx).Model(&domain.WechatAccount{}).Where("wxid = ?", wxid).Update("status", "offline").Error; err != nil {
		return "", err
	}
	return envelope.Message, nil
}

func (s *AccountService) SyncMessages(ctx context.Context, wxid string) (*SyncMessagesResult, error) {
	var account domain.WechatAccount
	if err := s.db.WithContext(ctx).Where("wxid = ?", wxid).First(&account).Error; err != nil {
		return nil, err
	}

	envelope, payload, _, err := s.wechatClient.SyncMessages(ctx, wxid, account.CurrentSyncKey)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	contacts, err := s.loadCachedContacts(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		if _, err := s.syncContacts(ctx, wxid); err == nil {
			contacts, _ = s.loadCachedContacts(ctx, wxid)
		}
	}

	contactIndex := make(map[string]domain.WechatContact, len(contacts))
	for _, contact := range contacts {
		contactIndex[contact.Wxid] = contact
	}

	messages := make([]MessageSummary, 0, len(payload.AddMsgs))
	for _, item := range payload.AddMsgs {
		summary := parseMessageSummary(item, wxid, &account, contactIndex)
		if err := persistMessage(ctx, s.db, wxid, summary, item.Content.String); err != nil {
			return nil, err
		}
		messages = append(messages, summary)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"last_sync_at":     now,
		"current_sync_key": payload.KeyBuf.Buffer,
	}
	if err := s.db.WithContext(ctx).Model(&domain.WechatAccount{}).Where("wxid = ?", wxid).Updates(updates).Error; err != nil {
		return nil, err
	}

	s.maybeProcessAIReplies(ctx, wxid, messages)

	return &SyncMessagesResult{
		ContinueFlag: payload.ContinueFlag,
		Status:       payload.Status,
		KeyBuf:       payload.KeyBuf.Buffer,
		Message:      envelope.Message,
		SyncedCount:  len(messages),
		Messages:     messages,
	}, nil
}

func (s *AccountService) SendTextMessage(ctx context.Context, wxid string, input SendTextInput) (*SendTextResult, error) {
	if strings.TrimSpace(input.ToWxid) == "" {
		return nil, errors.New("接收对象不能为空")
	}
	if strings.TrimSpace(input.Content) == "" {
		return nil, errors.New("发送内容不能为空")
	}

	envelope, payload, _, err := s.wechatClient.SendText(ctx, wxid, input.ToWxid, input.Content)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		return nil, errors.New(envelope.Message)
	}
	if len(payload.List) == 0 {
		return nil, errors.New("微信发送成功但未返回消息回执")
	}

	result := payload.List[0]
	return &SendTextResult{
		ToWxid:      result.ToUserName.String,
		ClientMsgID: result.ClientMsgID,
		CreateTime:  result.CreateTime,
		NewMsgID:    result.NewMsgID,
		Ret:         result.Ret,
	}, nil
}

func (s *AccountService) bootstrapAccount(ctx context.Context, wxid, platform string) (*BootstrapResult, error) {
	cacheEnvelope, cacheInfo, cacheRaw, err := s.wechatClient.GetCacheInfo(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if !cacheEnvelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, cacheEnvelope)
		return nil, errors.New(cacheEnvelope.Message)
	}

	initEnvelope, initPayload, initRaw, err := s.wechatClient.Init(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if !initEnvelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, initEnvelope)
		return nil, errors.New(initEnvelope.Message)
	}

	heartbeatEnvelope, _, err := s.wechatClient.StartHeartbeat(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if !heartbeatEnvelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, heartbeatEnvelope)
		return nil, errors.New(heartbeatEnvelope.Message)
	}

	var account domain.WechatAccount
	now := time.Now()
	err = s.db.WithContext(ctx).Where("wxid = ?", wxid).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		account = domain.WechatAccount{Wxid: wxid}
	} else if err != nil {
		return nil, err
	}

	account.Platform = firstNonEmpty(platform, account.Platform)
	account.Wxid = firstNonEmpty(cacheInfo.Wxid, account.Wxid)
	account.Nickname = firstNonEmpty(cacheInfo.Nickname, firstUserField(initPayload, "nickname"), account.Nickname)
	account.Alias = firstNonEmpty(cacheInfo.Alias, firstUserField(initPayload, "alias"), account.Alias)
	account.Mobile = firstNonEmpty(cacheInfo.Mobile, firstUserField(initPayload, "mobile"), account.Mobile)
	account.Avatar = firstNonEmpty(cacheInfo.HeadURL, account.Avatar)
	account.Signature = firstNonEmpty(firstUserField(initPayload, "signature"), account.Signature)
	account.DeviceID = firstNonEmpty(cacheInfo.DeviceID, account.DeviceID)
	account.DeviceName = firstNonEmpty(cacheInfo.DeviceName, account.DeviceName)
	account.Status = "online"
	account.LastHeartbeatAt = &now
	account.LastLoginAt = &now
	account.CurrentSyncKey = firstNonEmpty(initPayload.CurrentSyncKey.Buffer, account.CurrentSyncKey)
	account.MaxSyncKey = firstNonEmpty(initPayload.MaxSyncKey.Buffer, account.MaxSyncKey)
	account.LastCachePayload = string(cacheRaw)
	account.LastInitPayload = string(initRaw)

	if account.ID == 0 {
		if err := s.db.WithContext(ctx).Create(&account).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.WithContext(ctx).Save(&account).Error; err != nil {
			return nil, err
		}
	}

	return &BootstrapResult{
		Account:         account,
		HeartbeatStatus: heartbeatEnvelope.Message,
	}, nil
}

func (s *AccountService) extractWXID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	paths := [][]string{
		{"wxid"},
		{"Wxid"},
		{"WxId"},
		{"AcctSectResp", "UserName"},
		{"AcctSectResp", "userName"},
		{"acctSectResp", "userName"},
	}
	for _, path := range paths {
		if value := walkString(payload, path...); value != "" {
			return value
		}
	}
	return ""
}

func decodeLoginStatus(raw json.RawMessage) *wechat.LoginCheckPayload {
	if len(raw) == 0 {
		return nil
	}
	state := &wechat.LoginCheckPayload{}
	if err := json.Unmarshal(raw, state); err != nil {
		return nil
	}
	if state.UUID == "" && state.Status == 0 && state.ExpiredTime == 0 {
		return nil
	}
	return state
}

func mapPollingStatus(state *wechat.LoginCheckPayload) string {
	if state == nil {
		return "pending"
	}
	switch state.Status {
	case 0:
		return "pending"
	case 1:
		return "scanned"
	case 2:
		return "confirmed"
	default:
		return "pending"
	}
}

func walkString(payload map[string]interface{}, path ...string) string {
	var current interface{} = payload
	for _, key := range path {
		node, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = node[key]
	}
	if text, ok := current.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func extractLoginPreview(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}

	preview := map[string]string{
		"wxid":     walkString(payload, "acctSectResp", "userName"),
		"nickname": walkString(payload, "acctSectResp", "nickName"),
		"mobile":   walkString(payload, "acctSectResp", "bindMobile"),
		"alias":    walkString(payload, "acctSectResp", "alias"),
	}

	nonEmpty := 0
	for _, value := range preview {
		if strings.TrimSpace(value) != "" {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		return nil
	}
	return preview
}

func firstUserField(payload *wechat.InitPayload, field string) string {
	if payload == nil || len(payload.ModUserInfos) == 0 {
		return ""
	}
	user := payload.ModUserInfos[0]
	switch field {
	case "nickname":
		return user.NickName.String
	case "mobile":
		return user.BindMobile.String
	case "alias":
		return user.Alias
	case "signature":
		return user.Signature
	default:
		return ""
	}
}

func shouldSkipContact(wxid string) bool {
	if strings.TrimSpace(wxid) == "" {
		return true
	}
	systemIDs := map[string]struct{}{
		"medianote":   {},
		"floatbottle": {},
		"qmessage":    {},
		"newsapp":     {},
		"weixin":      {},
		"fmessage":    {},
		"filehelper":  {},
		"qqmail":      {},
	}
	_, exists := systemIDs[strings.TrimSpace(wxid)]
	return exists
}

func compactMessageContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) > 220 {
		return trimmed[:220] + "..."
	}
	return trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func fallbackStatus(values ...string) string {
	return firstNonEmpty(values...)
}

func buildSessionID() string {
	return fmt.Sprintf("ls_%d", time.Now().UnixNano())
}

func buildQRCodeURL(uuid string) string {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return ""
	}
	return "https://api.2dcode.biz/v1/create-qr-code?data=http://weixin.qq.com/x/" + uuid
}
