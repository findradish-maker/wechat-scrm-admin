package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"
	"wechat-enterprise-backend/internal/domain"

	"github.com/mozillazg/go-pinyin"
	"gorm.io/gorm/clause"
)

func (s *AccountService) ListContacts(ctx context.Context, wxid string, forceRefresh bool) (*ContactListResult, error) {
	if forceRefresh {
		return s.syncContacts(ctx, wxid)
	}

	contacts, err := s.loadCachedContacts(ctx, wxid)
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		return s.syncContacts(ctx, wxid)
	}

	return &ContactListResult{
		Contacts: mapContactSummaries(contacts),
	}, nil
}

func (s *AccountService) ReloadContactsAsync(wxid string) {
	if strings.TrimSpace(wxid) == "" {
		return
	}
	go func(targetWxid string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := s.syncContacts(ctx, targetWxid); err != nil {
			log.Printf("reload contacts failed for %s: %v", targetWxid, err)
		}
	}(wxid)
}

func (s *AccountService) loadCachedContacts(ctx context.Context, wxid string) ([]domain.WechatContact, error) {
	var contacts []domain.WechatContact
	if err := s.db.WithContext(ctx).
		Where("owner_wxid = ?", wxid).
		Order("contact_type asc, COALESCE(NULLIF(remark_py_initial, ''), NULLIF(py_initial, ''), NULLIF(remark, ''), NULLIF(nickname, ''), wxid) asc").
		Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

func (s *AccountService) syncContacts(ctx context.Context, wxid string) (*ContactListResult, error) {
	usernames, currentWxSeq, currentChatRoomSeq, continueFlag, err := s.fetchAllContactUsernames(ctx, wxid)
	if err != nil {
		return nil, err
	}

	existingUsernames, err := s.loadExistingContactWxids(ctx, wxid, usernames)
	if err != nil {
		return nil, err
	}

	pendingUsernames := make([]string, 0, len(usernames))
	groupIDs := make([]string, 0, len(usernames))
	for _, username := range usernames {
		if _, exists := existingUsernames[username]; exists {
			continue
		}
		pendingUsernames = append(pendingUsernames, username)
		if detectContactType(username) == "group" {
			groupIDs = append(groupIDs, username)
		}
	}

	now := time.Now()
	for _, batch := range chunkStrings(pendingUsernames, 20) {
		detailEnvelope, detailPayload, raw, err := s.wechatClient.GetContractDetail(ctx, wxid, batch)
		if err != nil {
			return nil, err
		}
		if !detailEnvelope.Success {
			return nil, errors.New(detailEnvelope.Message)
		}
		for _, item := range detailPayload.ContactList {
			if strings.TrimSpace(item.UserName.String) == "" {
				continue
			}
			rawPayload, _ := json.Marshal(item)
			contact := domain.WechatContact{
				OwnerWxid:       wxid,
				Wxid:            item.UserName.String,
				Nickname:        item.NickName.String,
				Alias:           item.Alias,
				VerifyFlag:      item.VerifyFlag,
				Remark:          item.Remark.String,
				PyInitial:       item.Pyinitial.String,
				QuanPin:         item.QuanPin.String,
				RemarkPyInitial: item.RemarkPyinitial.String,
				RemarkQuanPin:   item.RemarkQuanPin.String,
				Avatar:          firstNonEmpty(item.BigHeadImgURL, item.SmallHeadImgURL),
				Signature:       item.Signature,
				Province:        item.Province,
				City:            item.City,
				ContactType:     detectContactType(item.UserName.String),
				RawPayload:      firstNonEmpty(string(rawPayload), string(raw)),
				LastSyncedAt:    &now,
			}
			if contact.ContactType == "group" {
				contact.MemberCount = int(item.NewChatroomData.MemberCount)
				contact.ChatRoomOwner = strings.TrimSpace(item.ChatRoomOwner)
			}
			if err := s.upsertContact(ctx, contact); err != nil {
				return nil, err
			}
		}
	}

	for _, groupID := range groupIDs[:minInt(len(groupIDs), 20)] {
		infoEnvelope, infoPayload, infoRaw, err := s.wechatClient.GetChatRoomInfo(ctx, wxid, groupID)
		if err != nil {
			continue
		}
		if !infoEnvelope.Success {
			continue
		}
		if len(infoPayload.ContactList) == 0 {
			continue
		}

		item := infoPayload.ContactList[0]
		contact := domain.WechatContact{
			OwnerWxid:     wxid,
			Wxid:          item.UserName.String,
			Nickname:      item.NickName.String,
			Remark:        item.Remark.String,
			Avatar:        firstNonEmpty(item.BigHeadImgURL, item.SmallHeadImgURL),
			ContactType:   "group",
			MemberCount:   int(item.NewChatroomData.MemberCount),
			ChatRoomOwner: item.ChatRoomOwner,
			VerifyFlag:    0,
			RawPayload:    string(infoRaw),
			LastSyncedAt:  &now,
		}

		detailEnvelope, detailPayload, detailRaw, detailErr := s.wechatClient.GetChatRoomInfoDetail(ctx, wxid, groupID)
		if detailErr == nil && detailEnvelope.Success {
			contact.Announcement = detailPayload.Announcement
			contact.AnnouncementEditor = detailPayload.AnnouncementEditor
			if detailPayload.AnnouncementPublishTime > 0 {
				publishedAt := time.Unix(detailPayload.AnnouncementPublishTime, 0)
				contact.AnnouncementPublishTime = &publishedAt
			}
			if strings.TrimSpace(string(detailRaw)) != "" {
				contact.RawPayload = string(detailRaw)
			}
		}

		if err := s.upsertContact(ctx, contact); err != nil {
			return nil, err
		}
	}

	contacts, err := s.loadCachedContacts(ctx, wxid)
	if err != nil {
		return nil, err
	}

	return &ContactListResult{
		CurrentWxcontactSeq:       currentWxSeq,
		CurrentChatRoomContactSeq: currentChatRoomSeq,
		ContinueFlag:              continueFlag,
		Contacts:                  mapContactSummaries(contacts),
	}, nil
}

func (s *AccountService) loadExistingContactWxids(ctx context.Context, wxid string, usernames []string) (map[string]struct{}, error) {
	existing := make(map[string]struct{}, len(usernames))
	if len(usernames) == 0 {
		return existing, nil
	}

	var rows []string
	if err := s.db.WithContext(ctx).
		Model(&domain.WechatContact{}).
		Where("owner_wxid = ? AND wxid IN ?", wxid, usernames).
		Pluck("wxid", &rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		existing[item] = struct{}{}
	}
	return existing, nil
}

func (s *AccountService) fetchAllContactUsernames(ctx context.Context, wxid string) ([]string, int64, int64, int64, error) {
	seen := make(map[string]struct{})
	usernames := make([]string, 0, 512)
	var currentWxSeq int64
	var currentChatRoomSeq int64
	var continueFlag int64

	for page := 0; page < 500; page++ {
		envelope, payload, _, err := s.wechatClient.GetContractList(ctx, wxid, currentWxSeq, currentChatRoomSeq)
		if err != nil {
			return nil, 0, 0, 0, err
		}
		if !envelope.Success {
			s.handleOfflineEnvelope(ctx, wxid, envelope)
			return nil, 0, 0, 0, errors.New(envelope.Message)
		}

		for _, username := range payload.ContactUsernameList {
			if shouldSkipContact(username) {
				continue
			}
			if _, exists := seen[username]; exists {
				continue
			}
			seen[username] = struct{}{}
			usernames = append(usernames, username)
		}

		nextWxSeq := payload.CurrentWxcontactSeq
		nextChatRoomSeq := payload.CurrentChatRoomContactSeq
		continueFlag = payload.ContinueFlag
		if len(payload.ContactUsernameList) == 0 || continueFlag == 0 || (nextWxSeq == currentWxSeq && nextChatRoomSeq == currentChatRoomSeq) {
			currentWxSeq = nextWxSeq
			currentChatRoomSeq = nextChatRoomSeq
			break
		}
		currentWxSeq = nextWxSeq
		currentChatRoomSeq = nextChatRoomSeq
	}

	return usernames, currentWxSeq, currentChatRoomSeq, continueFlag, nil
}

func (s *AccountService) upsertContact(ctx context.Context, contact domain.WechatContact) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "owner_wxid"},
			{Name: "wxid"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"nickname":                  clause.Expr{SQL: "VALUES(nickname)"},
			"alias":                     clause.Expr{SQL: "VALUES(alias)"},
			"remark":                    clause.Expr{SQL: "VALUES(remark)"},
			"py_initial":                clause.Expr{SQL: "COALESCE(NULLIF(VALUES(py_initial), ''), py_initial)"},
			"quan_pin":                  clause.Expr{SQL: "COALESCE(NULLIF(VALUES(quan_pin), ''), quan_pin)"},
			"remark_py_initial":         clause.Expr{SQL: "COALESCE(NULLIF(VALUES(remark_py_initial), ''), remark_py_initial)"},
			"remark_quan_pin":           clause.Expr{SQL: "COALESCE(NULLIF(VALUES(remark_quan_pin), ''), remark_quan_pin)"},
			"avatar":                    clause.Expr{SQL: "VALUES(avatar)"},
			"signature":                 clause.Expr{SQL: "VALUES(signature)"},
			"province":                  clause.Expr{SQL: "VALUES(province)"},
			"city":                      clause.Expr{SQL: "VALUES(city)"},
			"contact_type":              clause.Expr{SQL: "VALUES(contact_type)"},
			"verify_flag":               clause.Expr{SQL: "VALUES(verify_flag)"},
			"member_count":              clause.Expr{SQL: "VALUES(member_count)"},
			"chat_room_owner":           clause.Expr{SQL: "VALUES(chat_room_owner)"},
			"announcement":              clause.Expr{SQL: "VALUES(announcement)"},
			"announcement_editor":       clause.Expr{SQL: "VALUES(announcement_editor)"},
			"announcement_publish_time": clause.Expr{SQL: "VALUES(announcement_publish_time)"},
			"raw_payload":               clause.Expr{SQL: "VALUES(raw_payload)"},
			"last_synced_at":            clause.Expr{SQL: "VALUES(last_synced_at)"},
			"updated_at":                clause.Expr{SQL: "VALUES(updated_at)"},
		}),
	}).Create(&contact).Error
}

func mapContactSummaries(contacts []domain.WechatContact) []ContactSummary {
	summaries := make([]ContactSummary, 0, len(contacts))
	for _, item := range contacts {
		category, categoryLabel := detectContactCategory(item.Wxid, item.ContactType, item.VerifyFlag)
		sortKey, sortLetter := buildContactSortMeta(item)
		summaries = append(summaries, ContactSummary{
			Wxid:                    item.Wxid,
			DisplayName:             contactDisplayName(item),
			Nickname:                item.Nickname,
			Alias:                   item.Alias,
			Remark:                  item.Remark,
			SortKey:                 sortKey,
			SortLetter:              sortLetter,
			Avatar:                  item.Avatar,
			Signature:               item.Signature,
			Province:                item.Province,
			City:                    item.City,
			ContactType:             item.ContactType,
			ContactCategory:         category,
			ContactCategoryLabel:    categoryLabel,
			VerifyFlag:              item.VerifyFlag,
			MemberCount:             item.MemberCount,
			ChatRoomOwner:           item.ChatRoomOwner,
			Announcement:            truncateText(item.Announcement, 160),
			AnnouncementPublishTime: item.AnnouncementPublishTime,
			LastSyncedAt:            item.LastSyncedAt,
		})
	}
	sort.SliceStable(summaries, func(i, j int) bool {
		left := summaries[i]
		right := summaries[j]
		if left.ContactCategory != right.ContactCategory {
			return categorySortOrder(left.ContactCategory) < categorySortOrder(right.ContactCategory)
		}
		if left.SortLetter != right.SortLetter {
			return left.SortLetter < right.SortLetter
		}
		if left.SortKey != right.SortKey {
			return left.SortKey < right.SortKey
		}
		return left.Wxid < right.Wxid
	})
	return summaries
}

func buildContactSortMeta(contact domain.WechatContact) (string, string) {
	sortKey := strings.ToUpper(strings.TrimSpace(contactSortSource(contact)))
	if sortKey == "" {
		return "#", "#"
	}

	sortLetter := "#"
	for _, char := range sortKey {
		upper := unicode.ToUpper(char)
		if upper >= 'A' && upper <= 'Z' {
			sortLetter = string(upper)
			break
		}
		if unicode.IsDigit(upper) {
			sortLetter = "#"
			break
		}
	}
	return sortKey, sortLetter
}

func contactSortSource(contact domain.WechatContact) string {
	displayName := strings.TrimSpace(contactDisplayName(contact))
	if latin := latinSortSource(displayName); latin != "" {
		return latin
	}
	if py := pinyinSortSource(displayName); py != "" {
		return py
	}

	switch {
	case strings.TrimSpace(contact.Remark) != "":
		return firstNonEmpty(
			contact.RemarkPyInitial,
			contact.RemarkQuanPin,
			contact.Remark,
		)
	case strings.TrimSpace(contact.Nickname) != "":
		return firstNonEmpty(
			contact.PyInitial,
			contact.QuanPin,
			contact.Nickname,
		)
	case strings.TrimSpace(contact.Alias) != "":
		return contact.Alias
	default:
		return contact.Wxid
	}
}

func latinSortSource(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	for _, char := range trimmed {
		upper := unicode.ToUpper(char)
		if upper >= 'A' && upper <= 'Z' {
			return string(upper) + trimmed
		}
		if unicode.IsDigit(upper) {
			return trimmed
		}
		break
	}
	return ""
}

func pinyinSortSource(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	args := pinyin.NewArgs()
	args.Style = pinyin.FirstLetter
	parts := pinyin.Pinyin(trimmed, args)
	if len(parts) == 0 {
		return ""
	}
	letters := make([]string, 0, len(parts))
	for _, item := range parts {
		if len(item) == 0 {
			continue
		}
		part := strings.ToUpper(strings.TrimSpace(item[0]))
		if part == "" {
			continue
		}
		letters = append(letters, part)
	}
	return strings.Join(letters, "")
}

func categorySortOrder(category string) int {
	switch category {
	case "contact":
		return 1
	case "group":
		return 2
	case "public_account":
		return 3
	default:
		return 99
	}
}

func detectContactType(wxid string) string {
	switch {
	case strings.HasSuffix(wxid, "@chatroom"):
		return "group"
	case strings.HasPrefix(wxid, "gh_"):
		return "official_account"
	default:
		return "contact"
	}
}

func detectContactCategory(wxid, contactType string, verifyFlag int64) (string, string) {
	switch contactType {
	case "group":
		return "group", "群聊"
	case "official_account":
		return "public_account", "公众号"
	default:
		return "contact", "联系人"
	}
}

func contactDisplayName(contact domain.WechatContact) string {
	return firstNonEmpty(contact.Remark, contact.Nickname, contact.Alias, contact.Wxid)
}

func chunkStrings(items []string, size int) [][]string {
	if size <= 0 || len(items) == 0 {
		return nil
	}
	chunks := make([][]string, 0, (len(items)+size-1)/size)
	for start := 0; start < len(items); start += size {
		end := start + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[start:end])
	}
	return chunks
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
