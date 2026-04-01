package domain

import "time"

type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Status       string    `gorm:"size:32;not null;default:active" json:"status"`
}

type WechatAccount struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	Wxid             string     `gorm:"uniqueIndex;size:128;not null" json:"wxid"`
	Nickname         string     `gorm:"size:255" json:"nickname"`
	Alias            string     `gorm:"size:255" json:"alias"`
	Mobile           string     `gorm:"size:64" json:"mobile"`
	Avatar           string     `gorm:"size:500" json:"avatar"`
	Signature        string     `gorm:"type:text" json:"signature"`
	Platform         string     `gorm:"size:64" json:"platform"`
	DeviceID         string     `gorm:"size:128" json:"deviceId"`
	DeviceName       string     `gorm:"size:255" json:"deviceName"`
	Status           string     `gorm:"size:32;default:offline" json:"status"`
	LastHeartbeatAt  *time.Time `json:"lastHeartbeatAt"`
	LastLoginAt      *time.Time `json:"lastLoginAt"`
	LastSyncAt       *time.Time `json:"lastSyncAt"`
	MaxSyncKey       string     `gorm:"type:text" json:"maxSyncKey"`
	CurrentSyncKey   string     `gorm:"type:text" json:"currentSyncKey"`
	LastCachePayload string     `gorm:"type:longtext" json:"-"`
	LastInitPayload  string     `gorm:"type:longtext" json:"-"`
}

type LoginSession struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	SessionID       string     `gorm:"uniqueIndex;size:64;not null" json:"sessionId"`
	Platform        string     `gorm:"size:64;not null" json:"platform"`
	DeviceID        string     `gorm:"size:128" json:"deviceId"`
	DeviceName      string     `gorm:"size:255" json:"deviceName"`
	UUID            string     `gorm:"size:255;index" json:"uuid"`
	QRBase64        string     `gorm:"type:longtext" json:"qrBase64"`
	QRURL           string     `gorm:"size:1000" json:"qrUrl"`
	Data62          string     `gorm:"type:longtext" json:"data62"`
	Status          string     `gorm:"size:32;default:pending" json:"status"`
	Wxid            string     `gorm:"size:128" json:"wxid"`
	ExpiresAt       *time.Time `json:"expiresAt"`
	LastCode        int        `json:"lastCode"`
	LastMessage     string     `gorm:"size:500" json:"lastMessage"`
	LastRawResponse string     `gorm:"type:longtext" json:"-"`
}

type WechatContact struct {
	ID                      uint       `gorm:"primaryKey" json:"id"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
	OwnerWxid               string     `gorm:"uniqueIndex:idx_owner_contact;size:128;not null" json:"ownerWxid"`
	Wxid                    string     `gorm:"uniqueIndex:idx_owner_contact;size:128;not null" json:"wxid"`
	Nickname                string     `gorm:"size:255" json:"nickname"`
	Alias                   string     `gorm:"size:255" json:"alias"`
	Remark                  string     `gorm:"size:255" json:"remark"`
	PyInitial               string     `gorm:"size:255" json:"pyInitial"`
	QuanPin                 string     `gorm:"size:255" json:"quanPin"`
	RemarkPyInitial         string     `gorm:"size:255" json:"remarkPyInitial"`
	RemarkQuanPin           string     `gorm:"size:255" json:"remarkQuanPin"`
	Avatar                  string     `gorm:"size:500" json:"avatar"`
	Signature               string     `gorm:"type:text" json:"signature"`
	Province                string     `gorm:"size:128" json:"province"`
	City                    string     `gorm:"size:128" json:"city"`
	ContactType             string     `gorm:"size:32;index;not null" json:"contactType"`
	VerifyFlag              int64      `gorm:"not null;default:0" json:"verifyFlag"`
	MemberCount             int        `gorm:"not null;default:0" json:"memberCount"`
	ChatRoomOwner           string     `gorm:"size:128" json:"chatRoomOwner"`
	Announcement            string     `gorm:"type:longtext" json:"announcement"`
	AnnouncementEditor      string     `gorm:"size:128" json:"announcementEditor"`
	AnnouncementPublishTime *time.Time `json:"announcementPublishTime"`
	RawPayload              string     `gorm:"type:longtext" json:"-"`
	LastSyncedAt            *time.Time `json:"lastSyncedAt"`
}

type WechatMessage struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	OwnerWxid        string     `gorm:"index;size:128;not null" json:"ownerWxid"`
	MsgID            int64      `gorm:"uniqueIndex:idx_owner_msg;not null" json:"msgId"`
	NewMsgID         int64      `gorm:"index" json:"newMsgId"`
	MsgSeq           int64      `gorm:"index" json:"msgSeq"`
	MsgType          int64      `gorm:"index" json:"msgType"`
	FromWxid         string     `gorm:"size:128;index" json:"fromWxid"`
	ToWxid           string     `gorm:"size:128;index" json:"toWxid"`
	ChatWxid         string     `gorm:"size:128;index" json:"chatWxid"`
	ChatDisplay      string     `gorm:"size:255" json:"chatDisplay"`
	ConversationType string     `gorm:"size:32;index" json:"conversationType"`
	SenderWxid       string     `gorm:"size:128;index" json:"senderWxid"`
	SenderDisplay    string     `gorm:"size:255" json:"senderDisplay"`
	Kind             string     `gorm:"size:32;index" json:"kind"`
	Content          string     `gorm:"type:longtext" json:"content"`
	Preview          string     `gorm:"size:500" json:"preview"`
	ContentMetaJSON  string     `gorm:"type:longtext" json:"-"`
	ArticleJSON      string     `gorm:"type:longtext" json:"-"`
	QuoteJSON        string     `gorm:"type:longtext" json:"-"`
	VoiceJSON        string     `gorm:"type:longtext" json:"-"`
	VideoJSON        string     `gorm:"type:longtext" json:"-"`
	ImageJSON        string     `gorm:"type:longtext" json:"-"`
	CardJSON         string     `gorm:"type:longtext" json:"-"`
	EmojiJSON        string     `gorm:"type:longtext" json:"-"`
	SystemJSON       string     `gorm:"type:longtext" json:"-"`
	DeliveryStatus   string     `gorm:"size:32;index;default:received" json:"deliveryStatus"`
	IsSelf           bool       `gorm:"index" json:"isSelf"`
	CreateTime       int64      `gorm:"index" json:"createTime"`
	ParseStatus      string     `gorm:"size:32;index;default:parsed" json:"parseStatus"`
	ParseError       string     `gorm:"size:500" json:"parseError"`
	DecodeXML        string     `gorm:"type:longtext" json:"-"`
	AIReplyHandled   bool       `gorm:"index;default:false" json:"aiReplyHandled"`
	AIReplyStatus    string     `gorm:"size:32;default:pending" json:"aiReplyStatus"`
	AIReplyError     string     `gorm:"size:500" json:"aiReplyError"`
	AIReplyAt        *time.Time `json:"aiReplyAt"`
	RawContent       string     `gorm:"type:longtext" json:"-"`
}

type AIConversationSetting struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	OwnerWxid             string     `gorm:"uniqueIndex:idx_owner_conversation_ai;size:128;not null" json:"ownerWxid"`
	ConversationID        string     `gorm:"uniqueIndex:idx_owner_conversation_ai;size:128;not null" json:"conversationId"`
	Provider              string     `gorm:"size:64;not null;default:deepseek" json:"provider"`
	Model                 string     `gorm:"size:128" json:"model"`
	Enabled               bool       `gorm:"not null;default:false" json:"enabled"`
	KeywordTriggerEnabled bool       `gorm:"not null;default:false" json:"keywordTriggerEnabled"`
	TriggerKeywords       string     `gorm:"type:text" json:"triggerKeywords"`
	SystemPrompt          string     `gorm:"type:longtext" json:"systemPrompt"`
	APIKey                string     `gorm:"type:text" json:"apiKey"`
	APIBaseURL            string     `gorm:"type:text" json:"apiBaseUrl"`
	LastReplyAt           *time.Time `json:"lastReplyAt"`
}
