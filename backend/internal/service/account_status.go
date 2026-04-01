package service

import (
	"context"
	"strings"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/internal/wechat"
)

func isAccountOfflineEnvelope(envelope *wechat.Envelope) bool {
	if envelope == nil {
		return false
	}
	if envelope.Code == -13 {
		return true
	}
	message := strings.TrimSpace(envelope.Message)
	return strings.Contains(message, "用户可能退出")
}

func (s *AccountService) markAccountOffline(ctx context.Context, wxid string) {
	wxid = strings.TrimSpace(wxid)
	if wxid == "" {
		return
	}
	_ = s.db.WithContext(ctx).
		Model(&domain.WechatAccount{}).
		Where("wxid = ?", wxid).
		Update("status", "offline").Error
}

func (s *AccountService) handleOfflineEnvelope(ctx context.Context, wxid string, envelope *wechat.Envelope) bool {
	if !isAccountOfflineEnvelope(envelope) {
		return false
	}
	s.markAccountOffline(ctx, wxid)
	return true
}
