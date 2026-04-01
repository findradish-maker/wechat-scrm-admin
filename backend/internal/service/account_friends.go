package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type FriendSearchInput struct {
	Keyword     string `json:"keyword"`
	FromScene   int32  `json:"fromScene"`
	SearchScene int32  `json:"searchScene"`
}

type FriendSearchResult struct {
	Keyword string               `json:"keyword"`
	Results []FriendSearchTarget `json:"results"`
}

type FriendSearchTarget struct {
	Wxid        string `json:"wxid"`
	DisplayName string `json:"displayName"`
	Nickname    string `json:"nickname"`
	Alias       string `json:"alias"`
	Avatar      string `json:"avatar"`
	Signature   string `json:"signature"`
	Province    string `json:"province"`
	City        string `json:"city"`
	VerifyFlag  uint32 `json:"verifyFlag"`
	MatchType   uint32 `json:"matchType"`
	CanAdd      bool   `json:"canAdd"`
	V1          string `json:"v1"`
	V2          string `json:"v2"`
	Source      string `json:"source"`
}

type SendFriendRequestInput struct {
	V1            string `json:"v1"`
	V2            string `json:"v2"`
	VerifyContent string `json:"verifyContent"`
	Scene         int64  `json:"scene"`
	Opcode        int32  `json:"opcode"`
	TargetWxid    string `json:"targetWxid"`
}

type SendFriendRequestResult struct {
	TargetWxid string `json:"targetWxid"`
	Username   string `json:"username"`
}

type FriendRelationResult struct {
	TargetWxid    string `json:"targetWxid"`
	RelationCode  uint32 `json:"relationCode"`
	RelationKey   string `json:"relationKey"`
	RelationLabel string `json:"relationLabel"`
	NickName      string `json:"nickName"`
	HeadImgURL    string `json:"headImgUrl"`
	Sign          string `json:"sign"`
	UpstreamRet   int64  `json:"upstreamRet"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}

type FriendBlacklistInput struct {
	Action string `json:"action"`
}

type FriendRelationBatchInput struct {
	UserNames []string `json:"userNames"`
}

type FriendRelationBatchResult struct {
	Failed  []FriendBatchFailure   `json:"failed"`
	Results []FriendRelationResult `json:"results"`
}

type SendFriendRequestBatchInput struct {
	Requests []SendFriendRequestInput `json:"requests"`
}

type SendFriendRequestBatchResult struct {
	Failed  []FriendBatchFailure      `json:"failed"`
	Results []SendFriendRequestResult `json:"results"`
}

type FriendBatchFailure struct {
	Reason     string `json:"reason"`
	TargetWxid string `json:"targetWxid"`
}

type FriendOperationResult struct {
	TargetWxid string `json:"targetWxid"`
	Action     string `json:"action"`
	Message    string `json:"message"`
}

func (s *AccountService) SearchFriendCandidates(ctx context.Context, wxid string, input FriendSearchInput) (*FriendSearchResult, error) {
	keyword := strings.TrimSpace(input.Keyword)
	if keyword == "" {
		return nil, errors.New("搜索关键词不能为空")
	}

	fromScene := input.FromScene
	searchScene := input.SearchScene
	if searchScene == 0 {
		searchScene = 1
	}

	envelope, payload, _, err := s.wechatClient.SearchFriend(ctx, wxid, keyword, fromScene, searchScene)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	results := make([]FriendSearchTarget, 0, 8)
	seen := map[string]struct{}{}

	appendTarget := func(item FriendSearchTarget) {
		item.Wxid = strings.TrimSpace(item.Wxid)
		if item.Wxid == "" {
			return
		}
		if _, exists := seen[item.Wxid]; exists {
			return
		}
		seen[item.Wxid] = struct{}{}
		item.DisplayName = firstNonEmpty(item.Nickname, item.Alias, item.Wxid)
		item.CanAdd = strings.TrimSpace(item.V1) != "" && strings.TrimSpace(item.V2) != ""
		results = append(results, item)
	}

	appendTarget(FriendSearchTarget{
		Wxid:      payload.UserName.String,
		Nickname:  payload.NickName.String,
		Alias:     payload.Alias,
		Avatar:    firstNonEmpty(payload.BigHeadImgURL, payload.SmallHeadImgURL),
		Signature: payload.Signature,
		Province:  payload.Province,
		City:      payload.City,
		V1:        payload.UserName.String,
		V2:        payload.AntispamTicket,
		Source:    "primary",
	})

	for _, item := range payload.ContactList {
		appendTarget(FriendSearchTarget{
			Wxid:       item.UserName.String,
			Nickname:   item.NickName.String,
			Alias:      item.Alias,
			Avatar:     firstNonEmpty(item.BigHeadImgURL, item.SmallHeadImgURL),
			Signature:  item.Signature,
			Province:   item.Province,
			City:       item.City,
			VerifyFlag: item.VerifyFlag,
			MatchType:  item.MatchType,
			V1:         item.UserName.String,
			V2:         item.AntispamTicket,
			Source:     "list",
		})
	}

	return &FriendSearchResult{
		Keyword: keyword,
		Results: results,
	}, nil
}

func (s *AccountService) SendFriendRequest(ctx context.Context, wxid string, input SendFriendRequestInput) (*SendFriendRequestResult, error) {
	v1 := strings.TrimSpace(input.V1)
	v2 := strings.TrimSpace(input.V2)
	if v1 == "" || v2 == "" {
		return nil, errors.New("缺少好友校验参数，请先通过搜索联系人获取")
	}

	verifyContent := strings.TrimSpace(input.VerifyContent)
	if verifyContent == "" {
		verifyContent = "你好，我是企业微信运营同学"
	}

	scene := input.Scene
	if scene == 0 {
		scene = 3
	}
	opcode := input.Opcode
	if opcode == 0 {
		opcode = 2
	}

	envelope, payload, _, err := s.wechatClient.SendFriendRequest(ctx, wxid, v1, v2, verifyContent, scene, opcode)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	return &SendFriendRequestResult{
		TargetWxid: firstNonEmpty(strings.TrimSpace(input.TargetWxid), v1),
		Username:   strings.TrimSpace(payload.Username),
	}, nil
}

func (s *AccountService) CheckFriendRelation(ctx context.Context, wxid, targetWxid string) (*FriendRelationResult, error) {
	targetWxid = strings.TrimSpace(targetWxid)
	if targetWxid == "" {
		return nil, errors.New("目标好友不能为空")
	}

	envelope, payload, _, err := s.wechatClient.GetFriendState(ctx, wxid, targetWxid)
	if err != nil {
		return nil, err
	}
	if payload != nil && payload.BaseResponse.Ret != 0 {
		errorMessage := strings.TrimSpace(payload.BaseResponse.ErrMsg.String)
		if errorMessage == "" {
			errorMessage = firstNonEmpty(strings.TrimSpace(envelope.Message), fmt.Sprintf("好友状态检测失败(ret=%d)", payload.BaseResponse.Ret))
		}
		return &FriendRelationResult{
			TargetWxid:    targetWxid,
			RelationCode:  payload.FriendRelation,
			RelationKey:   "check_failed",
			RelationLabel: "检测失败",
			NickName:      payload.NickName,
			HeadImgURL:    payload.HeadImgURL,
			Sign:          payload.Sign,
			UpstreamRet:   payload.BaseResponse.Ret,
			ErrorMessage:  errorMessage,
		}, nil
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	key, label := mapFriendRelation(payload.FriendRelation)
	return &FriendRelationResult{
		TargetWxid:    targetWxid,
		RelationCode:  payload.FriendRelation,
		RelationKey:   key,
		RelationLabel: label,
		NickName:      payload.NickName,
		HeadImgURL:    payload.HeadImgURL,
		Sign:          payload.Sign,
		UpstreamRet:   payload.BaseResponse.Ret,
	}, nil
}

func (s *AccountService) DeleteFriend(ctx context.Context, wxid, targetWxid string) (*FriendOperationResult, error) {
	targetWxid = strings.TrimSpace(targetWxid)
	if targetWxid == "" {
		return nil, errors.New("目标好友不能为空")
	}

	envelope, _, err := s.wechatClient.DeleteFriend(ctx, wxid, targetWxid)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	return &FriendOperationResult{
		TargetWxid: targetWxid,
		Action:     "delete",
		Message:    "删除好友请求已提交",
	}, nil
}

func (s *AccountService) SetFriendBlacklist(ctx context.Context, wxid, targetWxid string, input FriendBlacklistInput) (*FriendOperationResult, error) {
	targetWxid = strings.TrimSpace(targetWxid)
	if targetWxid == "" {
		return nil, errors.New("目标好友不能为空")
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	val := int32(15)
	actionLabel := "add"
	if action == "remove" || action == "unblacklist" {
		val = 7
		actionLabel = "remove"
	}

	envelope, _, err := s.wechatClient.BlacklistFriend(ctx, wxid, targetWxid, val)
	if err != nil {
		return nil, err
	}
	if !envelope.Success {
		s.handleOfflineEnvelope(ctx, wxid, envelope)
		return nil, errors.New(envelope.Message)
	}

	return &FriendOperationResult{
		TargetWxid: targetWxid,
		Action:     actionLabel,
		Message:    fmt.Sprintf("黑名单操作已提交：%s", actionLabel),
	}, nil
}

func (s *AccountService) CheckFriendRelationBatch(ctx context.Context, wxid string, input FriendRelationBatchInput) (*FriendRelationBatchResult, error) {
	targets := uniqueUserNames(input.UserNames)
	if len(targets) == 0 {
		return nil, errors.New("请至少传入一个目标联系人")
	}

	result := &FriendRelationBatchResult{
		Results: make([]FriendRelationResult, 0, len(targets)),
		Failed:  make([]FriendBatchFailure, 0),
	}

	for _, target := range targets {
		item, err := s.CheckFriendRelation(ctx, wxid, target)
		if err != nil {
			result.Failed = append(result.Failed, FriendBatchFailure{
				TargetWxid: target,
				Reason:     err.Error(),
			})
			continue
		}
		result.Results = append(result.Results, *item)
	}

	return result, nil
}

func (s *AccountService) SendFriendRequestBatch(ctx context.Context, wxid string, input SendFriendRequestBatchInput) (*SendFriendRequestBatchResult, error) {
	if len(input.Requests) == 0 {
		return nil, errors.New("请至少传入一个添加请求")
	}

	result := &SendFriendRequestBatchResult{
		Results: make([]SendFriendRequestResult, 0, len(input.Requests)),
		Failed:  make([]FriendBatchFailure, 0),
	}

	for _, req := range input.Requests {
		item, err := s.SendFriendRequest(ctx, wxid, req)
		target := firstNonEmpty(strings.TrimSpace(req.TargetWxid), strings.TrimSpace(req.V1))
		if err != nil {
			result.Failed = append(result.Failed, FriendBatchFailure{
				TargetWxid: target,
				Reason:     err.Error(),
			})
			continue
		}
		result.Results = append(result.Results, *item)
	}

	return result, nil
}

func uniqueUserNames(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func mapFriendRelation(code uint32) (string, string) {
	switch code {
	case 0:
		return "friend", "好友"
	case 1:
		return "deleted", "已删除"
	case 4:
		return "blocked_by_self", "已拉黑对方"
	case 5:
		return "blocked_by_target", "被对方拉黑"
	default:
		return "unknown", fmt.Sprintf("未知状态(%d)", code)
	}
}
