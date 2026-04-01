package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/internal/wechat"

	"gorm.io/gorm"
)

type ConversationQuery struct {
	Page     int
	PageSize int
	Keyword  string
}

type ConversationSummary struct {
	ConversationID   string `json:"conversationId"`
	ConversationType string `json:"conversationType"`
	TargetWxid       string `json:"targetWxid"`
	TargetName       string `json:"targetName"`
	TargetAvatar     string `json:"targetAvatar"`
	Remark           string `json:"remark"`
	GroupName        string `json:"groupName"`
	LastMessage      string `json:"lastMessage"`
	LastMessageType  string `json:"lastMessageType"`
	LastMessageTime  int64  `json:"lastMessageTime"`
	UnreadCount      int64  `json:"unreadCount"`
	MemberCount      int    `json:"memberCount"`
	ChatRoomOwner    string `json:"chatRoomOwner"`
	AccountWxid      string `json:"accountWxid"`
}

type ConversationListResult struct {
	Page          int                   `json:"page"`
	PageSize      int                   `json:"pageSize"`
	Total         int64                 `json:"total"`
	TotalPages    int                   `json:"totalPages"`
	Conversations []ConversationSummary `json:"conversations"`
}

type ConversationDetail struct {
	ConversationID   string                    `json:"conversationId"`
	ConversationType string                    `json:"conversationType"`
	TargetWxid       string                    `json:"targetWxid"`
	TargetName       string                    `json:"targetName"`
	TargetAvatar     string                    `json:"targetAvatar"`
	Remark           string                    `json:"remark"`
	GroupName        string                    `json:"groupName"`
	Announcement     string                    `json:"announcement"`
	MemberCount      int                       `json:"memberCount"`
	ChatRoomOwner    string                    `json:"chatRoomOwner"`
	GroupMembers     []ConversationGroupMember `json:"groupMembers,omitempty"`
	AccountWxid      string                    `json:"accountWxid"`
}

type ConversationGroupMember struct {
	UserName           string `json:"userName"`
	NickName           string `json:"nickName"`
	ChatroomMemberFlag int64  `json:"chatroomMemberFlag"`
	InviterUserName    string `json:"inviterUserName"`
}

type ConversationMessagesResult struct {
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
	Messages   []ConversationMessage `json:"messages"`
}

type ConversationMessage struct {
	MessageID      uint        `json:"messageId"`
	MsgID          int64       `json:"msgId"`
	ConversationID string      `json:"conversationId"`
	FromWxid       string      `json:"fromWxid"`
	ToWxid         string      `json:"toWxid"`
	SenderWxid     string      `json:"senderWxid"`
	SenderName     string      `json:"senderName"`
	SenderAvatar   string      `json:"senderAvatar"`
	IsSelf         bool        `json:"isSelf"`
	MessageType    string      `json:"messageType"`
	Content        string      `json:"content"`
	ContentMeta    interface{} `json:"contentMeta"`
	CreatedAt      int64       `json:"createdAt"`
	Status         string      `json:"status"`
	ChatType       string      `json:"chatType"`
}

type SendImageInput struct {
	Base64 string `json:"base64"`
}

type SendEmojiInput struct {
	Md5      string `json:"md5"`
	TotalLen int64  `json:"totalLen"`
}

type EmojiCatalogItem struct {
	Md5        string `json:"md5"`
	TotalLen   int64  `json:"totalLen"`
	Width      int64  `json:"width"`
	Height     int64  `json:"height"`
	LastUsedAt int64  `json:"lastUsedAt"`
	Label      string `json:"label"`
}

type ShareCardInput struct {
	CardWxid     string `json:"cardWxid"`
	CardNickname string `json:"cardNickname"`
	CardAlias    string `json:"cardAlias"`
}

type ShareLinkInput struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	ThumbURL    string `json:"thumbUrl"`
}

type conversationAggregateRow struct {
	ChatWxid         string
	ConversationType string
	LastMessageTime  int64
	LastRecordID     uint
}

type conversationProfile struct {
	TargetName    string
	TargetAvatar  string
	Remark        string
	GroupName     string
	Announcement  string
	MemberCount   int
	ChatRoomOwner string
}

func (s *AccountService) ListConversations(ctx context.Context, wxid string, query ConversationQuery) (*ConversationListResult, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize, 20, 200)
	account, contacts, contactIndex, err := s.loadAccountConversationContext(ctx, wxid)
	if err != nil {
		return nil, err
	}

	var rows []conversationAggregateRow
	if err := applyVisibleMessagesScope(s.db.WithContext(ctx)).
		Model(&domain.WechatMessage{}).
		Select("chat_wxid, conversation_type, MAX(create_time) as last_message_time, MAX(id) as last_record_id").
		Where("owner_wxid = ?", wxid).
		Group("chat_wxid, conversation_type").
		Order("last_message_time desc, last_record_id desc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	recordIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.LastRecordID > 0 {
			recordIDs = append(recordIDs, row.LastRecordID)
		}
	}

	lastMessages := make(map[uint]domain.WechatMessage, len(recordIDs))
	if len(recordIDs) > 0 {
		var messages []domain.WechatMessage
		if err := s.db.WithContext(ctx).Where("id IN ?", recordIDs).Find(&messages).Error; err != nil {
			return nil, err
		}
		for _, item := range messages {
			lastMessages[item.ID] = item
		}
	}

	keyword := strings.ToLower(strings.TrimSpace(query.Keyword))
	conversations := make([]ConversationSummary, 0, len(rows))
	for _, row := range rows {
		lastMessage, exists := lastMessages[row.LastRecordID]
		if !exists {
			continue
		}
		profile := resolveConversationProfile(account, contactIndex, row.ChatWxid, row.ConversationType)
		summary := ConversationSummary{
			ConversationID:   row.ChatWxid,
			ConversationType: mapConversationType(row.ConversationType),
			TargetWxid:       row.ChatWxid,
			TargetName:       profile.TargetName,
			TargetAvatar:     profile.TargetAvatar,
			Remark:           profile.Remark,
			GroupName:        profile.GroupName,
			LastMessage:      buildConversationPreview(lastMessage),
			LastMessageType:  mapMessageType(lastMessage.Kind),
			LastMessageTime:  row.LastMessageTime,
			UnreadCount:      0,
			MemberCount:      profile.MemberCount,
			ChatRoomOwner:    profile.ChatRoomOwner,
			AccountWxid:      wxid,
		}
		if keyword != "" {
			haystacks := []string{
				strings.ToLower(summary.TargetName),
				strings.ToLower(summary.Remark),
				strings.ToLower(summary.GroupName),
				strings.ToLower(summary.LastMessage),
				strings.ToLower(summary.TargetWxid),
			}
			matched := false
			for _, haystack := range haystacks {
				if strings.Contains(haystack, keyword) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		conversations = append(conversations, summary)
	}

	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastMessageTime > conversations[j].LastMessageTime
	})

	total := int64(len(conversations))
	start := (page - 1) * pageSize
	if start >= len(conversations) {
		return &ConversationListResult{
			Page:          page,
			PageSize:      pageSize,
			Total:         total,
			TotalPages:    totalPages(total, pageSize),
			Conversations: []ConversationSummary{},
		}, nil
	}
	end := start + pageSize
	if end > len(conversations) {
		end = len(conversations)
	}

	pageConversations := conversations[start:end]

	_ = contacts
	return &ConversationListResult{
		Page:          page,
		PageSize:      pageSize,
		Total:         total,
		TotalPages:    totalPages(total, pageSize),
		Conversations: pageConversations,
	}, nil
}

func (s *AccountService) ListRecentConversationEmojis(ctx context.Context, wxid string, limit int) ([]EmojiCatalogItem, error) {
	if strings.TrimSpace(wxid) == "" {
		return nil, errors.New("账号不能为空")
	}
	if limit <= 0 {
		limit = 48
	}
	if limit > 200 {
		limit = 200
	}

	var rows []domain.WechatMessage
	if err := s.db.WithContext(ctx).
		Where("owner_wxid = ? AND kind = ?", wxid, "emoji").
		Order("create_time desc, id desc").
		Limit(limit * 6).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	results := make([]EmojiCatalogItem, 0, limit)
	seen := make(map[string]struct{}, limit)
	for _, row := range rows {
		meta := parseConversationMeta(row)
		value, ok := meta.(map[string]interface{})
		if !ok {
			continue
		}
		md5 := strings.TrimSpace(fmt.Sprint(value["md5"]))
		if md5 == "" {
			continue
		}
		totalLen := toInt64(value["length"])
		key := fmt.Sprintf("%s:%d", md5, totalLen)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, EmojiCatalogItem{
			Md5:        md5,
			TotalLen:   totalLen,
			Width:      toInt64(value["width"]),
			Height:     toInt64(value["height"]),
			LastUsedAt: row.CreateTime,
			Label:      fmt.Sprintf("表情 %s", shortHash(md5)),
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (s *AccountService) enrichConversationSummariesFromContactDetail(ctx context.Context, wxid string, summaries []ConversationSummary) {
	targetIDs := make([]string, 0, len(summaries))
	for _, item := range summaries {
		if shouldEnrichByContactDetail(item.TargetWxid, item.ConversationType) {
			targetIDs = append(targetIDs, item.TargetWxid)
		}
	}
	if len(targetIDs) == 0 {
		return
	}

	envelope, payload, _, err := s.wechatClient.GetContractDetail(ctx, wxid, targetIDs)
	if err != nil || envelope == nil || !envelope.Success || payload == nil || len(payload.ContactList) == 0 {
		return
	}

	detailMap := make(map[string]wechat.ContactDetail, len(payload.ContactList))
	for _, detail := range payload.ContactList {
		key := strings.TrimSpace(detail.UserName.String)
		if key == "" {
			continue
		}
		detailMap[key] = detail
	}

	now := time.Now()
	for index := range summaries {
		item := &summaries[index]
		if !shouldEnrichByContactDetail(item.TargetWxid, item.ConversationType) {
			continue
		}
		detail, exists := detailMap[item.TargetWxid]
		if !exists {
			continue
		}

		if strings.TrimSpace(detail.NickName.String) != "" {
			item.TargetName = detail.NickName.String
			item.GroupName = detail.NickName.String
		}
		if strings.TrimSpace(detail.Remark.String) != "" {
			item.Remark = detail.Remark.String
		}
		if strings.TrimSpace(detail.ChatRoomOwner) != "" {
			item.ChatRoomOwner = detail.ChatRoomOwner
		}
		if detail.NewChatroomData.MemberCount > 0 {
			item.MemberCount = int(detail.NewChatroomData.MemberCount)
		}
		if avatar := firstNonEmpty(detail.BigHeadImgURL, detail.SmallHeadImgURL); strings.TrimSpace(avatar) != "" {
			item.TargetAvatar = avatar
		}

		contactType := detectContactType(item.TargetWxid)
		memberCount := 0
		chatRoomOwner := ""
		if item.ConversationType == "group" || strings.HasSuffix(item.TargetWxid, "@chatroom") {
			memberCount = item.MemberCount
			chatRoomOwner = item.ChatRoomOwner
		}
		contact := domain.WechatContact{
			OwnerWxid:     wxid,
			Wxid:          item.TargetWxid,
			Nickname:      firstNonEmpty(item.TargetName, item.GroupName),
			Alias:         strings.TrimSpace(detail.Alias),
			Remark:        item.Remark,
			Avatar:        item.TargetAvatar,
			ContactType:   contactType,
			MemberCount:   memberCount,
			ChatRoomOwner: chatRoomOwner,
			LastSyncedAt:  &now,
		}
		_ = s.upsertContact(ctx, contact)
	}
}

func (s *AccountService) GetConversationDetail(ctx context.Context, wxid, conversationID string) (*ConversationDetail, error) {
	account, _, contactIndex, err := s.loadAccountConversationContext(ctx, wxid)
	if err != nil {
		return nil, err
	}

	baseType := detectConversationType(conversationID)
	var row domain.WechatMessage
	if err := s.db.WithContext(ctx).
		Where("owner_wxid = ? AND chat_wxid = ?", wxid, conversationID).
		Order("create_time desc, msg_seq desc").
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profile := resolveConversationProfile(account, contactIndex, conversationID, baseType)
			detail := &ConversationDetail{
				ConversationID:   conversationID,
				ConversationType: mapConversationType(baseType),
				TargetWxid:       conversationID,
				TargetName:       profile.TargetName,
				TargetAvatar:     profile.TargetAvatar,
				Remark:           profile.Remark,
				GroupName:        profile.GroupName,
				Announcement:     profile.Announcement,
				MemberCount:      profile.MemberCount,
				ChatRoomOwner:    profile.ChatRoomOwner,
				AccountWxid:      wxid,
			}
			s.enrichConversationDetailFromContactDetail(ctx, wxid, conversationID, detail, contactIndex)
			return detail, nil
		}
		return nil, err
	}

	profile := resolveConversationProfile(account, contactIndex, conversationID, row.ConversationType)
	detail := &ConversationDetail{
		ConversationID:   conversationID,
		ConversationType: mapConversationType(row.ConversationType),
		TargetWxid:       conversationID,
		TargetName:       profile.TargetName,
		TargetAvatar:     profile.TargetAvatar,
		Remark:           profile.Remark,
		GroupName:        profile.GroupName,
		Announcement:     profile.Announcement,
		MemberCount:      profile.MemberCount,
		ChatRoomOwner:    profile.ChatRoomOwner,
		AccountWxid:      wxid,
	}
	s.enrichConversationDetailFromContactDetail(ctx, wxid, conversationID, detail, contactIndex)
	return detail, nil
}

func (s *AccountService) enrichConversationDetailFromContactDetail(
	ctx context.Context,
	wxid, conversationID string,
	detail *ConversationDetail,
	contactIndex map[string]domain.WechatContact,
) {
	if detail == nil || !shouldEnrichByContactDetail(conversationID, detail.ConversationType) {
		return
	}
	if !shouldFetchConversationDetailFromUpstream(conversationID, detail, contactIndex) {
		return
	}

	envelope, payload, _, err := s.wechatClient.GetContractDetail(ctx, wxid, []string{conversationID})
	if err != nil || envelope == nil || !envelope.Success || payload == nil || len(payload.ContactList) == 0 {
		return
	}

	item := payload.ContactList[0]
	if strings.TrimSpace(item.NickName.String) != "" {
		detail.TargetName = item.NickName.String
		detail.GroupName = item.NickName.String
	}
	if strings.TrimSpace(item.Remark.String) != "" {
		detail.Remark = item.Remark.String
	}
	if avatar := firstNonEmpty(item.BigHeadImgURL, item.SmallHeadImgURL); strings.TrimSpace(avatar) != "" {
		detail.TargetAvatar = avatar
	}

	if detail.ConversationType != "group" {
		return
	}

	if strings.TrimSpace(item.ChatRoomOwner) != "" {
		detail.ChatRoomOwner = item.ChatRoomOwner
	}
	if item.NewChatroomData.MemberCount > 0 {
		detail.MemberCount = int(item.NewChatroomData.MemberCount)
	}
	if len(item.NewChatroomData.ChatRoomMember) == 0 {
		return
	}

	members := make([]ConversationGroupMember, 0, len(item.NewChatroomData.ChatRoomMember))
	for _, member := range item.NewChatroomData.ChatRoomMember {
		members = append(members, ConversationGroupMember{
			UserName:           strings.TrimSpace(member.UserName),
			NickName:           strings.TrimSpace(member.NickName),
			ChatroomMemberFlag: member.ChatroomMemberFlag,
			InviterUserName:    strings.TrimSpace(member.InviterUserName),
		})
	}
	detail.GroupMembers = members
}

func shouldFetchConversationDetailFromUpstream(
	conversationID string,
	detail *ConversationDetail,
	contactIndex map[string]domain.WechatContact,
) bool {
	contact, exists := contactIndex[conversationID]
	if !exists {
		return true
	}

	if detail == nil {
		return false
	}

	if strings.TrimSpace(contact.Nickname) == "" &&
		strings.TrimSpace(contact.Remark) == "" &&
		strings.TrimSpace(contact.Avatar) == "" {
		return true
	}

	if detail.ConversationType == "group" {
		if contact.MemberCount <= 0 && detail.MemberCount <= 0 {
			return true
		}
		if strings.TrimSpace(contact.ChatRoomOwner) == "" && strings.TrimSpace(detail.ChatRoomOwner) == "" {
			return true
		}
	}

	return false
}

func shouldEnrichByContactDetail(targetWxid, conversationType string) bool {
	return conversationType == "group" || strings.HasSuffix(targetWxid, "@chatroom") || strings.HasPrefix(targetWxid, "gh_")
}

func (s *AccountService) ListConversationMessages(ctx context.Context, wxid, conversationID string, page, pageSize int) (*ConversationMessagesResult, error) {
	page, pageSize = normalizePage(page, pageSize, 30, 200)
	account, _, contactIndex, err := s.loadAccountConversationContext(ctx, wxid)
	if err != nil {
		return nil, err
	}

	dbQuery := s.db.WithContext(ctx).
		Model(&domain.WechatMessage{}).
		Where("owner_wxid = ? AND chat_wxid = ?", wxid, conversationID)
	dbQuery = applyVisibleMessagesScope(dbQuery)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []domain.WechatMessage
	if err := dbQuery.
		Order("create_time desc, msg_seq desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreateTime == rows[j].CreateTime {
			if rows[i].MsgSeq == rows[j].MsgSeq {
				return rows[i].ID < rows[j].ID
			}
			return rows[i].MsgSeq < rows[j].MsgSeq
		}
		return rows[i].CreateTime < rows[j].CreateTime
	})

	s.enrichConversationMessageParticipants(ctx, wxid, conversationID, rows, contactIndex)

	messages := make([]ConversationMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, mapConversationMessage(row, account, contactIndex))
	}

	return &ConversationMessagesResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages(total, pageSize),
		Messages:   messages,
	}, nil
}

func (s *AccountService) DeleteConversation(ctx context.Context, wxid, conversationID string) error {
	if strings.TrimSpace(wxid) == "" {
		return errors.New("账号不能为空")
	}
	if strings.TrimSpace(conversationID) == "" {
		return errors.New("会话不能为空")
	}

	return s.db.WithContext(ctx).
		Where("owner_wxid = ? AND chat_wxid = ?", wxid, conversationID).
		Delete(&domain.WechatMessage{}).Error
}

func (s *AccountService) enrichConversationMessageParticipants(
	ctx context.Context,
	wxid string,
	conversationID string,
	rows []domain.WechatMessage,
	contactIndex map[string]domain.WechatContact,
) {
	candidates := make(map[string]struct{}, 32)

	if shouldEnrichByContactDetail(conversationID, detectConversationType(conversationID)) {
		if _, exists := contactIndex[conversationID]; !exists {
			candidates[conversationID] = struct{}{}
		}
	}

	for _, row := range rows {
		sender := strings.TrimSpace(row.SenderWxid)
		if sender == "" || sender == wxid {
			continue
		}
		if _, exists := contactIndex[sender]; exists {
			continue
		}
		// 群聊成员或公众号发送者都尽量补齐，避免消息气泡落回 wxid。
		candidates[sender] = struct{}{}
	}

	if len(candidates) == 0 {
		return
	}

	targetIDs := make([]string, 0, len(candidates))
	for id := range candidates {
		targetIDs = append(targetIDs, id)
	}

	now := time.Now()
	for _, batch := range chunkStrings(targetIDs, 20) {
		envelope, payload, _, err := s.wechatClient.GetContractDetail(ctx, wxid, batch)
		if err != nil || envelope == nil || !envelope.Success || payload == nil {
			continue
		}
		for _, detail := range payload.ContactList {
			contactWxid := strings.TrimSpace(detail.UserName.String)
			if contactWxid == "" {
				continue
			}

			contact := domain.WechatContact{
				OwnerWxid:    wxid,
				Wxid:         contactWxid,
				Nickname:     strings.TrimSpace(detail.NickName.String),
				Alias:        strings.TrimSpace(detail.Alias),
				Remark:       strings.TrimSpace(detail.Remark.String),
				Avatar:       firstNonEmpty(detail.BigHeadImgURL, detail.SmallHeadImgURL),
				Signature:    strings.TrimSpace(detail.Signature),
				Province:     strings.TrimSpace(detail.Province),
				City:         strings.TrimSpace(detail.City),
				ContactType:  detectContactType(contactWxid),
				LastSyncedAt: &now,
			}
			if contact.ContactType == "group" {
				contact.MemberCount = int(detail.NewChatroomData.MemberCount)
				contact.ChatRoomOwner = strings.TrimSpace(detail.ChatRoomOwner)
			}

			_ = s.upsertContact(ctx, contact)
			contactIndex[contactWxid] = contact
		}
	}
}

func (s *AccountService) SendConversationText(ctx context.Context, wxid, conversationID string, input SendTextInput) (*ConversationMessage, error) {
	envelope, payload, _, err := s.wechatClient.SendText(ctx, wxid, conversationID, input.Content)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}
	return s.storeOutgoingMessage(ctx, wxid, conversationID, "text", input.Content, truncateText(input.Content, 120), nil, normalizeTextAck(payload))
}

func (s *AccountService) SendConversationImage(ctx context.Context, wxid, conversationID string, input SendImageInput) (*ConversationMessage, error) {
	base64 := normalizeBase64Payload(input.Base64)
	if base64 == "" {
		return nil, errors.New("图片内容不能为空")
	}
	envelope, payload, _, err := s.wechatClient.UploadImage(ctx, wxid, conversationID, base64)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}
	meta := &MessageImage{
		Base64:   "data:image/png;base64," + base64,
		ThumbURL: "data:image/png;base64," + base64,
	}
	return s.storeOutgoingMessage(ctx, wxid, conversationID, "image", "[图片]", "[图片]", meta, firstAck(payload))
}

func (s *AccountService) SendConversationEmoji(ctx context.Context, wxid, conversationID string, input SendEmojiInput) (*ConversationMessage, error) {
	if strings.TrimSpace(input.Md5) == "" {
		return nil, errors.New("表情 MD5 不能为空")
	}
	envelope, payload, _, err := s.wechatClient.SendEmoji(ctx, wxid, conversationID, input.Md5, input.TotalLen)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}
	meta := &MessageEmoji{
		MD5:    strings.TrimSpace(input.Md5),
		Length: input.TotalLen,
	}
	return s.storeOutgoingMessage(ctx, wxid, conversationID, "emoji", "[表情]", "[表情]", meta, firstAck(payload))
}

func (s *AccountService) SendConversationCard(ctx context.Context, wxid, conversationID string, input ShareCardInput) (*ConversationMessage, error) {
	if strings.TrimSpace(input.CardWxid) == "" {
		return nil, errors.New("名片 wxid 不能为空")
	}
	envelope, payload, _, err := s.wechatClient.ShareCard(ctx, wxid, conversationID, input.CardWxid, input.CardNickname, input.CardAlias)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}
	meta := &MessageCard{
		Wxid:     strings.TrimSpace(input.CardWxid),
		Nickname: strings.TrimSpace(input.CardNickname),
		Alias:    strings.TrimSpace(input.CardAlias),
	}
	preview := firstNonEmpty(meta.Nickname, meta.Alias, meta.Wxid, "名片消息")
	return s.storeOutgoingMessage(ctx, wxid, conversationID, "card", preview, preview, meta, firstAck(payload))
}

func (s *AccountService) SendConversationLink(ctx context.Context, wxid, conversationID string, input ShareLinkInput) (*ConversationMessage, error) {
	if strings.TrimSpace(input.URL) == "" {
		return nil, errors.New("分享链接不能为空")
	}
	xmlContent := buildLinkShareXML(input)
	envelope, payload, _, err := s.wechatClient.ShareLink(ctx, wxid, conversationID, xmlContent)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}
	meta := &MessageArticle{
		Title:     strings.TrimSpace(input.Title),
		Summary:   strings.TrimSpace(input.Description),
		URL:       strings.TrimSpace(input.URL),
		Cover:     strings.TrimSpace(input.ThumbURL),
		Publisher: "手动分享",
	}
	preview := firstNonEmpty(meta.Title, meta.URL, "分享链接")
	return s.storeOutgoingMessage(ctx, wxid, conversationID, "article", preview, preview, meta, firstAck(payload))
}

func (s *AccountService) loadAccountConversationContext(ctx context.Context, wxid string) (*domain.WechatAccount, []domain.WechatContact, map[string]domain.WechatContact, error) {
	var account domain.WechatAccount
	if err := s.db.WithContext(ctx).Where("wxid = ?", wxid).First(&account).Error; err != nil {
		return nil, nil, nil, err
	}
	contacts, err := s.loadCachedContacts(ctx, wxid)
	if err != nil {
		return nil, nil, nil, err
	}
	contactIndex := make(map[string]domain.WechatContact, len(contacts))
	for _, contact := range contacts {
		contactIndex[contact.Wxid] = contact
	}
	return &account, contacts, contactIndex, nil
}

func resolveConversationProfile(account *domain.WechatAccount, contactIndex map[string]domain.WechatContact, chatWxid, conversationType string) conversationProfile {
	if account != nil && chatWxid == account.Wxid {
		return conversationProfile{
			TargetName:   firstNonEmpty(account.Nickname, account.Alias, account.Wxid),
			TargetAvatar: account.Avatar,
		}
	}
	if contact, exists := contactIndex[chatWxid]; exists {
		return conversationProfile{
			TargetName:    contactDisplayName(contact),
			TargetAvatar:  contact.Avatar,
			Remark:        contact.Remark,
			GroupName:     contact.Nickname,
			Announcement:  contact.Announcement,
			MemberCount:   contact.MemberCount,
			ChatRoomOwner: contact.ChatRoomOwner,
		}
	}
	return conversationProfile{
		TargetName: chatWxid,
	}
}

func mapConversationMessage(row domain.WechatMessage, account *domain.WechatAccount, contactIndex map[string]domain.WechatContact) ConversationMessage {
	meta := parseConversationMeta(row)
	return ConversationMessage{
		MessageID:      row.ID,
		MsgID:          row.MsgID,
		ConversationID: row.ChatWxid,
		FromWxid:       row.FromWxid,
		ToWxid:         row.ToWxid,
		SenderWxid:     row.SenderWxid,
		SenderName:     resolveConversationSenderName(row, account, contactIndex),
		SenderAvatar:   resolveConversationSenderAvatar(row, account, contactIndex),
		IsSelf:         row.IsSelf,
		MessageType:    mapMessageType(row.Kind),
		Content:        firstNonEmpty(row.Content, row.Preview),
		ContentMeta:    meta,
		CreatedAt:      row.CreateTime,
		Status:         firstNonEmpty(row.DeliveryStatus, deliveryStatusFromBool(row.IsSelf)),
		ChatType:       mapConversationType(row.ConversationType),
	}
}

func parseConversationMeta(row domain.WechatMessage) interface{} {
	if strings.TrimSpace(row.ContentMetaJSON) != "" {
		var result interface{}
		if err := json.Unmarshal([]byte(row.ContentMetaJSON), &result); err == nil {
			return result
		}
	}
	switch row.Kind {
	case "article":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.ArticleJSON), &value)
		return value
	case "quote":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.QuoteJSON), &value)
		return value
	case "voice":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.VoiceJSON), &value)
		return value
	case "video":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.VideoJSON), &value)
		return value
	case "image":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.ImageJSON), &value)
		return value
	case "card":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.CardJSON), &value)
		return value
	case "emoji":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.EmojiJSON), &value)
		return value
	case "system", "system_notice":
		var value map[string]interface{}
		_ = json.Unmarshal([]byte(row.SystemJSON), &value)
		return value
	default:
		return nil
	}
}

func buildConversationPreview(message domain.WechatMessage) string {
	preview := firstNonEmpty(message.Preview, message.Content)
	if message.ConversationType == "group" && message.SenderDisplay != "" {
		return fmt.Sprintf("%s：%s", message.SenderDisplay, preview)
	}
	return preview
}

func resolveConversationSenderName(row domain.WechatMessage, account *domain.WechatAccount, contactIndex map[string]domain.WechatContact) string {
	if row.SenderDisplay != "" {
		return row.SenderDisplay
	}
	if account != nil && row.SenderWxid == account.Wxid {
		return firstNonEmpty(account.Nickname, account.Alias, account.Wxid)
	}
	if contact, exists := contactIndex[row.SenderWxid]; exists {
		return contactDisplayName(contact)
	}
	return row.SenderWxid
}

func resolveConversationSenderAvatar(row domain.WechatMessage, account *domain.WechatAccount, contactIndex map[string]domain.WechatContact) string {
	if account != nil && row.SenderWxid == account.Wxid {
		return account.Avatar
	}
	if contact, exists := contactIndex[row.SenderWxid]; exists {
		return contact.Avatar
	}
	return ""
}

func mapConversationType(value string) string {
	if value == "group" {
		return "group"
	}
	return "single"
}

func detectConversationType(chatWxid string) string {
	if strings.HasSuffix(chatWxid, "@chatroom") {
		return "group"
	}
	return "direct"
}

func mapMessageType(kind string) string {
	switch kind {
	case "article":
		return "link"
	case "voice", "video", "image", "emoji", "card", "quote", "system", "system_notice", "app":
		return kind
	default:
		return "text"
	}
}

func normalizeTextAck(payload *wechat.SendTextPayload) *wechat.SendMessageAck {
	if payload == nil || len(payload.List) == 0 {
		return nil
	}
	item := payload.List[0]
	return &wechat.SendMessageAck{
		Ret:         item.Ret,
		ToUserName:  item.ToUserName,
		ToUsetName:  item.ToUserName,
		MsgID:       item.MsgID,
		ClientMsgID: item.ClientMsgID,
		CreateTime:  item.CreateTime,
		ServerTime:  item.ServerTime,
		Type:        item.Type,
		NewMsgID:    item.NewMsgID,
	}
}

func firstAck(payload *wechat.SendMessagePayload) *wechat.SendMessageAck {
	if payload == nil || len(payload.List) == 0 {
		return nil
	}
	return &payload.List[0]
}

func normalizeBase64Payload(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, ",") && strings.HasPrefix(trimmed, "data:") {
		parts := strings.SplitN(trimmed, ",", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return trimmed
}

func buildLinkShareXML(input ShareLinkInput) string {
	title := strings.TrimSpace(input.Title)
	desc := strings.TrimSpace(input.Description)
	urlValue := strings.TrimSpace(input.URL)
	thumbURL := strings.TrimSpace(input.ThumbURL)
	return fmt.Sprintf(
		`<msg><appmsg appid="" sdkver="0"><title><![CDATA[%s]]></title><des><![CDATA[%s]]></des><action></action><type>5</type><showtype>0</showtype><content><![CDATA[]]></content><contentattr>0</contentattr><url><![CDATA[%s]]></url><thumburl><![CDATA[%s]]></thumburl></appmsg></msg>`,
		title,
		desc,
		urlValue,
		thumbURL,
	)
}

func (s *AccountService) storeOutgoingMessage(ctx context.Context, wxid, conversationID, kind, content, preview string, meta interface{}, ack *wechat.SendMessageAck) (*ConversationMessage, error) {
	account, _, contactIndex, err := s.loadAccountConversationContext(ctx, wxid)
	if err != nil {
		return nil, err
	}
	profile := resolveConversationProfile(account, contactIndex, conversationID, detectConversationType(conversationID))

	createTime := time.Now().Unix()
	msgID := -time.Now().UnixNano()
	newMsgID := int64(0)
	msgType := int64(49)
	if ack != nil {
		if ack.CreateTime > 0 {
			createTime = ack.CreateTime
		}
		if ack.MsgID != 0 {
			msgID = ack.MsgID
		}
		newMsgID = ack.NewMsgID
		if ack.Type != 0 {
			msgType = ack.Type
		}
	}

	record := domain.WechatMessage{
		OwnerWxid:        wxid,
		MsgID:            msgID,
		NewMsgID:         newMsgID,
		MsgSeq:           createTime,
		MsgType:          msgType,
		FromWxid:         wxid,
		ToWxid:           conversationID,
		ChatWxid:         conversationID,
		ChatDisplay:      profile.TargetName,
		ConversationType: detectConversationType(conversationID),
		SenderWxid:       wxid,
		SenderDisplay:    firstNonEmpty(account.Nickname, account.Alias, account.Wxid),
		Kind:             kind,
		Content:          content,
		Preview:          preview,
		ContentMetaJSON:  marshalJSON(meta),
		ImageJSON:        marshalJSON(ifKind(kind, "image", meta)),
		CardJSON:         marshalJSON(ifKind(kind, "card", meta)),
		EmojiJSON:        marshalJSON(ifKind(kind, "emoji", meta)),
		ArticleJSON:      marshalJSON(ifKind(kind, "article", meta)),
		DeliveryStatus:   "sent",
		IsSelf:           true,
		CreateTime:       createTime,
		ParseStatus:      "parsed",
		RawContent:       content,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, err
	}
	message := mapConversationMessage(record, account, contactIndex)
	return &message, nil
}

func ifKind(kind, expected string, value interface{}) interface{} {
	if kind == expected {
		return value
	}
	return nil
}

func toInt64(value interface{}) int64 {
	switch current := value.(type) {
	case int:
		return int64(current)
	case int32:
		return int64(current)
	case int64:
		return current
	case float32:
		return int64(current)
	case float64:
		return int64(current)
	case json.Number:
		result, _ := current.Int64()
		return result
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(current), 10, 64)
		return parsed
	default:
		return 0
	}
}

func shortHash(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func deliveryStatusFromBool(isSelf bool) string {
	if isSelf {
		return "sent"
	}
	return "received"
}
