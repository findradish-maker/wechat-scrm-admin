package service

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"
	"wechat-enterprise-backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func applyVisibleMessagesScope(db *gorm.DB) *gorm.DB {
	return db.Where(
		"(kind <> ?) AND NOT (kind = ? AND (raw_content LIKE ? OR system_json LIKE ? OR system_json LIKE ?))",
		"handoff",
		"system",
		"%HandOffMaster%",
		"%多端接力%",
		"%appentrypage%",
	)
}

type ContactQuery struct {
	Page         int
	PageSize     int
	ForceRefresh bool
	ContactType  string
	Category     string
	Keyword      string
}

type MessageQuery struct {
	Page             int
	PageSize         int
	Kind             string
	ConversationType string
	Keyword          string
}

func (s *AccountService) QueryContacts(ctx context.Context, wxid string, query ContactQuery) (*ContactListResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize, 20, 5000)
	if query.ForceRefresh {
		if _, err := s.syncContacts(ctx, wxid); err != nil {
			return nil, err
		}
	}

	dbQuery := s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("owner_wxid = ?", wxid)
	if strings.TrimSpace(query.ContactType) != "" {
		dbQuery = dbQuery.Where("contact_type = ?", strings.TrimSpace(query.ContactType))
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		dbQuery = dbQuery.Where(
			"wxid LIKE ? OR nickname LIKE ? OR alias LIKE ? OR remark LIKE ?",
			like, like, like, like,
		)
	}

	var contacts []domain.WechatContact
	if err := dbQuery.
		Order("contact_type asc, COALESCE(NULLIF(remark, ''), NULLIF(nickname, ''), wxid) asc").
		Find(&contacts).Error; err != nil {
		return nil, err
	}

	summaries := mapContactSummaries(contacts)
	category := strings.TrimSpace(query.Category)
	if category != "" {
		filtered := make([]ContactSummary, 0, len(summaries))
		for _, item := range summaries {
			if item.ContactCategory == category {
				filtered = append(filtered, item)
			}
		}
		summaries = filtered
	}

	total := int64(len(summaries))
	start := (page - 1) * pageSize
	if start > len(summaries) {
		start = len(summaries)
	}
	end := start + pageSize
	if end > len(summaries) {
		end = len(summaries)
	}
	paged := summaries[start:end]

	return &ContactListResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages(total, pageSize),
		Contacts:   paged,
	}, nil
}

func (s *AccountService) ListMessages(ctx context.Context, wxid string, query MessageQuery) (*MessageListResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize, 15, 200)
	dbQuery := s.db.WithContext(ctx).Model(&domain.WechatMessage{}).Where("owner_wxid = ?", wxid)

	if strings.TrimSpace(query.Kind) != "" {
		dbQuery = dbQuery.Where("kind = ?", strings.TrimSpace(query.Kind))
	}
	if strings.TrimSpace(query.ConversationType) != "" {
		dbQuery = dbQuery.Where("conversation_type = ?", strings.TrimSpace(query.ConversationType))
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		dbQuery = dbQuery.Where(
			"chat_display LIKE ? OR sender_display LIKE ? OR preview LIKE ? OR content LIKE ? OR from_wxid LIKE ? OR to_wxid LIKE ?",
			like, like, like, like, like, like,
		)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []domain.WechatMessage
	if err := dbQuery.
		Order("create_time desc, msg_seq desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return &MessageListResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages(total, pageSize),
		Messages:   mapStoredMessages(rows),
	}, nil
}

func (s *AccountService) GetDashboardOverview(ctx context.Context) (*DashboardOverview, error) {
	var accounts []domain.WechatAccount
	if err := s.db.WithContext(ctx).Order("updated_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	overview := &DashboardOverview{}
	overview.Summary.TotalAccounts = int64(len(accounts))
	for _, account := range accounts {
		if account.Status == "online" {
			overview.Summary.OnlineAccounts++
		}
		if account.Status != "online" {
			overview.Summary.OfflineAccounts++
		}
		if account.LastSyncAt != nil {
			overview.Summary.AccountsWithSync++
		}
		if account.LastHeartbeatAt != nil && account.LastHeartbeatAt.After(time.Now().Add(-24*time.Hour)) {
			overview.Summary.AccountsActive24Hours++
		}
	}

	s.db.WithContext(ctx).Model(&domain.WechatContact{}).Count(&overview.Summary.TotalContacts)
	s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("contact_type = ?", "contact").Count(&overview.Summary.DirectContacts)
	s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("contact_type = ?", "official_account").Count(&overview.Summary.OfficialAccounts)
	s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("contact_type = ?", "group").Count(&overview.Summary.Groups)
	s.db.WithContext(ctx).Model(&domain.WechatMessage{}).Count(&overview.Summary.TotalMessages)
	s.db.WithContext(ctx).Model(&domain.WechatMessage{}).Where("create_time >= ?", time.Now().Add(-24*time.Hour).Unix()).Count(&overview.Summary.Messages24Hours)

	rows := make([]DashboardAccountRow, 0, len(accounts))
	for _, account := range accounts {
		row := DashboardAccountRow{
			Wxid:            account.Wxid,
			Nickname:        account.Nickname,
			Avatar:          account.Avatar,
			Status:          account.Status,
			LastHeartbeatAt: account.LastHeartbeatAt,
			LastSyncAt:      account.LastSyncAt,
		}
		s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("owner_wxid = ? AND contact_type = ?", account.Wxid, "contact").Count(&row.DirectContactCount)
		s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("owner_wxid = ? AND contact_type = ?", account.Wxid, "official_account").Count(&row.OfficialAccountCount)
		s.db.WithContext(ctx).Model(&domain.WechatContact{}).Where("owner_wxid = ? AND contact_type = ?", account.Wxid, "group").Count(&row.GroupCount)
		s.db.WithContext(ctx).Model(&domain.WechatMessage{}).Where("owner_wxid = ?", account.Wxid).Count(&row.MessageCount)
		rows = append(rows, row)
	}
	overview.Accounts = rows

	return overview, nil
}

func persistMessage(ctx context.Context, db *gorm.DB, ownerWxid string, summary MessageSummary, rawContent string) error {
	if shouldHideMessageSummary(summary, rawContent) {
		return nil
	}

	contentMetaJSON := marshalJSON(buildMessageMeta(summary))
	articleJSON := marshalJSON(summary.Article)
	quoteJSON := marshalJSON(summary.Quote)
	voiceJSON := marshalJSON(summary.Voice)
	videoJSON := marshalJSON(summary.Video)
	imageJSON := marshalJSON(summary.Image)
	cardJSON := marshalJSON(summary.Card)
	emojiJSON := marshalJSON(summary.Emoji)
	systemJSON := marshalJSON(summary.System)

	record := domain.WechatMessage{
		OwnerWxid:        ownerWxid,
		MsgID:            summary.MsgID,
		NewMsgID:         summary.NewMsgID,
		MsgSeq:           summary.MsgSeq,
		MsgType:          summary.MsgType,
		FromWxid:         summary.FromWxid,
		ToWxid:           summary.ToWxid,
		ChatWxid:         summary.ChatWxid,
		ChatDisplay:      summary.ChatDisplay,
		ConversationType: summary.ConversationType,
		SenderWxid:       summary.SenderWxid,
		SenderDisplay:    summary.SenderDisplay,
		Kind:             summary.Kind,
		Content:          summary.Content,
		Preview:          summary.Preview,
		ContentMetaJSON:  contentMetaJSON,
		ArticleJSON:      articleJSON,
		QuoteJSON:        quoteJSON,
		VoiceJSON:        voiceJSON,
		VideoJSON:        videoJSON,
		ImageJSON:        imageJSON,
		CardJSON:         cardJSON,
		EmojiJSON:        emojiJSON,
		SystemJSON:       systemJSON,
		DeliveryStatus:   deliveryStatusFromSummary(summary),
		IsSelf:           summary.IsSelf,
		CreateTime:       summary.CreateTime,
		ParseStatus:      firstNonEmpty(strings.TrimSpace(summary.ParseStatus), "parsed"),
		ParseError:       strings.TrimSpace(summary.ParseError),
		DecodeXML:        strings.TrimSpace(summary.DecodeXML),
		RawContent:       rawContent,
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "owner_wxid"},
			{Name: "msg_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"new_msg_id",
			"msg_seq",
			"msg_type",
			"from_wxid",
			"to_wxid",
			"chat_wxid",
			"chat_display",
			"conversation_type",
			"sender_wxid",
			"sender_display",
			"kind",
			"content",
			"preview",
			"content_meta_json",
			"article_json",
			"quote_json",
			"voice_json",
			"video_json",
			"image_json",
			"card_json",
			"emoji_json",
			"system_json",
			"delivery_status",
			"is_self",
			"create_time",
			"parse_status",
			"parse_error",
			"decode_xml",
			"raw_content",
			"updated_at",
		}),
	}).Create(&record).Error
}

func mapStoredMessages(rows []domain.WechatMessage) []MessageSummary {
	result := make([]MessageSummary, 0, len(rows))
	for _, row := range rows {
		summary := MessageSummary{
			MsgID:            row.MsgID,
			NewMsgID:         row.NewMsgID,
			MsgType:          row.MsgType,
			FromWxid:         row.FromWxid,
			ToWxid:           row.ToWxid,
			ChatWxid:         row.ChatWxid,
			ChatDisplay:      row.ChatDisplay,
			ConversationType: row.ConversationType,
			SenderWxid:       row.SenderWxid,
			SenderDisplay:    row.SenderDisplay,
			Content:          row.Content,
			Preview:          row.Preview,
			Kind:             row.Kind,
			CreateTime:       row.CreateTime,
			MsgSeq:           row.MsgSeq,
			IsSelf:           row.IsSelf,
			ParseStatus:      row.ParseStatus,
			ParseError:       row.ParseError,
			DecodeXML:        row.DecodeXML,
		}
		unmarshalJSON(row.ArticleJSON, &summary.Article)
		unmarshalJSON(row.QuoteJSON, &summary.Quote)
		unmarshalJSON(row.VoiceJSON, &summary.Voice)
		unmarshalJSON(row.VideoJSON, &summary.Video)
		unmarshalJSON(row.ImageJSON, &summary.Image)
		unmarshalJSON(row.CardJSON, &summary.Card)
		unmarshalJSON(row.EmojiJSON, &summary.Emoji)
		unmarshalJSON(row.SystemJSON, &summary.System)
		result = append(result, summary)
	}
	return result
}

func buildMessageMeta(summary MessageSummary) any {
	switch summary.Kind {
	case "article":
		return summary.Article
	case "quote":
		return summary.Quote
	case "voice":
		return summary.Voice
	case "video":
		return summary.Video
	case "image":
		return summary.Image
	case "card":
		return summary.Card
	case "emoji":
		return summary.Emoji
	case "system", "system_notice":
		return summary.System
	default:
		return nil
	}
}

func shouldHideMessageSummary(summary MessageSummary, rawContent string) bool {
	if summary.Kind == "handoff" {
		return true
	}
	return isHiddenSystemPayload(summary.Kind, rawContent, marshalJSON(summary.System))
}

func isHiddenMessageRow(row domain.WechatMessage) bool {
	return isHiddenSystemPayload(row.Kind, row.RawContent, row.SystemJSON)
}

func isHiddenSystemPayload(kind, rawContent, systemJSON string) bool {
	if strings.TrimSpace(kind) == "handoff" {
		return true
	}
	if strings.TrimSpace(kind) != "system" {
		return false
	}

	rawContent = strings.TrimSpace(rawContent)
	systemJSON = strings.TrimSpace(systemJSON)

	return strings.Contains(rawContent, "HandOffMaster") ||
		strings.Contains(systemJSON, "多端接力") ||
		strings.Contains(systemJSON, "appentrypage")
}

func deliveryStatusFromSummary(summary MessageSummary) string {
	if summary.IsSelf {
		return "sent"
	}
	return "received"
}

func marshalJSON(value any) string {
	if value == nil {
		return ""
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(payload)
}

func unmarshalJSON[T any](raw string, target **T) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	value := new(T)
	if err := json.Unmarshal([]byte(raw), value); err == nil {
		*target = value
	}
}

func normalizePage(page, pageSize, defaultPageSize, maxPageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

func totalPages(total int64, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
