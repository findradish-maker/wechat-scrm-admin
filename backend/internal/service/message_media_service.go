package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/internal/wechat"
)

type ConversationImageResult struct {
	Height    int64  `json:"height"`
	MessageID uint   `json:"messageId"`
	Src       string `json:"src"`
	Width     int64  `json:"width"`
}

func (s *AccountService) DownloadConversationImage(ctx context.Context, wxid string, messageID uint) (*ConversationImageResult, error) {
	if wxid == "" || messageID == 0 {
		return nil, errors.New("消息参数无效")
	}

	var row domain.WechatMessage
	if err := s.db.WithContext(ctx).
		Where("owner_wxid = ? AND id = ?", wxid, messageID).
		First(&row).Error; err != nil {
		return nil, err
	}

	if row.Kind != "image" {
		return nil, errors.New("当前消息不是图片消息")
	}

	meta := extractImageMeta(row)
	if meta == nil {
		return nil, errors.New("图片消息缺少下载参数")
	}
	if row.MsgID <= 0 || row.MsgID > math.MaxUint32 {
		return nil, errors.New("图片消息ID无效")
	}
	if meta.Length <= 0 {
		return nil, errors.New("图片大小未知，无法下载高清图")
	}

	input := wechat.DownloadImageInput{
		Wxid:         wxid,
		ToWxid:       row.ChatWxid,
		MsgID:        uint32(row.MsgID),
		DataLen:      int(meta.Length),
		CompressType: 0,
	}
	input.Section.StartPos = 0
	input.Section.DataLen = uint32(meta.Length)

	envelope, payload, raw, err := s.wechatClient.DownloadImage(ctx, input)
	if err != nil {
		return nil, err
	}
	if envelope == nil {
		return nil, errors.New("高清图下载失败")
	}
	if !envelope.Success || envelope.Code != 0 {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, fmt.Errorf("高清图下载失败：%s", firstNonEmpty(envelope.Message, string(raw)))
	}
	if payload == nil {
		return nil, errors.New("高清图响应为空")
	}

	buffer := strings.TrimSpace(payload.Data.Buffer)
	if buffer == "" {
		return nil, errors.New("高清图数据为空")
	}

	// 标准化一下，保证前端直接可用。
	if _, err := base64.StdEncoding.DecodeString(buffer); err != nil {
		return nil, fmt.Errorf("高清图数据格式异常：%w", err)
	}

	return &ConversationImageResult{
		MessageID: messageID,
		Src:       "data:image/jpeg;base64," + buffer,
		Width:     meta.Width,
		Height:    meta.Height,
	}, nil
}

func extractImageMeta(row domain.WechatMessage) *MessageImage {
	if strings.TrimSpace(row.ImageJSON) != "" {
		value := &MessageImage{}
		if err := json.Unmarshal([]byte(row.ImageJSON), value); err == nil && value != nil {
			if value.Length > 0 || value.Base64 != "" || value.URL != "" || value.ThumbURL != "" {
				return value
			}
		}
	}

	if strings.TrimSpace(row.ContentMetaJSON) != "" {
		value := &MessageImage{}
		if err := json.Unmarshal([]byte(row.ContentMetaJSON), value); err == nil && value != nil {
			if value.Length > 0 || value.Base64 != "" || value.URL != "" || value.ThumbURL != "" {
				return value
			}
		}
	}

	if strings.TrimSpace(row.RawContent) == "" {
		return nil
	}

	return parseImageMessage(wechat.SyncMessage{}, row.RawContent)
}
