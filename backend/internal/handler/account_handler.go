package handler

import (
	"net/http"
	"strconv"
	"wechat-enterprise-backend/internal/response"
	"wechat-enterprise-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.accountService.ListAccounts(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.Success(c, accounts)
}

func (h *AccountHandler) CreateLoginSession(c *gin.Context) {
	var request service.CreateLoginSessionInput
	_ = c.ShouldBindJSON(&request)

	result, err := h.accountService.CreateLoginSession(c.Request.Context(), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40010, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) PollLoginSession(c *gin.Context) {
	result, err := h.accountService.PollLoginSession(c.Request.Context(), c.Param("sessionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40011, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) CreateAwakenLoginSession(c *gin.Context) {
	result, err := h.accountService.CreateAwakenLoginSession(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40018, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) Bootstrap(c *gin.Context) {
	result, err := h.accountService.BootstrapAccount(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40012, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) StartHeartbeat(c *gin.Context) {
	message, err := h.accountService.StartHeartbeat(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40013, err.Error())
		return
	}
	response.Success(c, gin.H{"message": message})
}

func (h *AccountHandler) StopHeartbeat(c *gin.Context) {
	message, err := h.accountService.StopHeartbeat(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40014, err.Error())
		return
	}
	response.Success(c, gin.H{"message": message})
}

func (h *AccountHandler) Logout(c *gin.Context) {
	message, err := h.accountService.Logout(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40015, err.Error())
		return
	}
	response.Success(c, gin.H{"message": message})
}

func (h *AccountHandler) ListContacts(c *gin.Context) {
	refresh := c.Query("refresh") == "1" || c.Query("refresh") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := h.accountService.QueryContacts(c.Request.Context(), c.Param("wxid"), service.ContactQuery{
		Page:         page,
		PageSize:     pageSize,
		ForceRefresh: refresh,
		ContactType:  c.Query("contactType"),
		Category:     c.Query("category"),
		Keyword:      c.Query("keyword"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40016, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ReloadContacts(c *gin.Context) {
	h.accountService.ReloadContactsAsync(c.Param("wxid"))
	response.Success(c, gin.H{"message": "联系人后台刷新已开始"})
}

func (h *AccountHandler) ListFavorites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := h.accountService.ListFavorites(c.Request.Context(), c.Param("wxid"), service.FavoriteQuery{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40067, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetFinderProfile(c *gin.Context) {
	result, err := h.accountService.GetFinderProfile(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetFavoriteDetail(c *gin.Context) {
	favID, err := strconv.ParseInt(c.Param("favId"), 10, 32)
	if err != nil || favID <= 0 {
		response.Error(c, http.StatusBadRequest, 40068, "收藏ID无效")
		return
	}
	result, err := h.accountService.GetFavoriteDetail(c.Request.Context(), c.Param("wxid"), int32(favID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40069, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) DeleteFavorite(c *gin.Context) {
	favID, err := strconv.ParseInt(c.Param("favId"), 10, 32)
	if err != nil || favID <= 0 {
		response.Error(c, http.StatusBadRequest, 40070, "收藏ID无效")
		return
	}
	result, err := h.accountService.DeleteFavorite(c.Request.Context(), c.Param("wxid"), int32(favID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40071, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListMoments(c *gin.Context) {
	maxID, _ := strconv.ParseUint(c.DefaultQuery("maxId", "0"), 10, 64)
	result, err := h.accountService.ListMoments(
		c.Request.Context(),
		c.Param("wxid"),
		maxID,
		c.Query("firstPageMd5"),
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40043, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetMomentDetail(c *gin.Context) {
	result, err := h.accountService.GetMomentDetail(
		c.Request.Context(),
		c.Param("wxid"),
		c.Query("authorWxid"),
		c.Param("momentId"),
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40044, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) PublishMoment(c *gin.Context) {
	var request service.PublishMomentInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40045, "朋友圈发布参数无效")
		return
	}
	result, err := h.accountService.PublishMoment(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40046, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) OperateMoment(c *gin.Context) {
	var request service.OperateMomentInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40047, "朋友圈操作参数无效")
		return
	}
	result, err := h.accountService.OperateMoment(c.Request.Context(), c.Param("wxid"), c.Param("momentId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40048, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	result, err := h.accountService.ListMessages(c.Request.Context(), c.Param("wxid"), service.MessageQuery{
		Page:             page,
		PageSize:         pageSize,
		Kind:             c.Query("kind"),
		ConversationType: c.Query("conversationType"),
		Keyword:          c.Query("keyword"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40020, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListConversations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result, err := h.accountService.ListConversations(c.Request.Context(), c.Param("wxid"), service.ConversationQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40023, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListAIProviders(c *gin.Context) {
	response.Success(c, h.accountService.ListAIProviders())
}

func (h *AccountHandler) GetConversationDetail(c *gin.Context) {
	result, err := h.accountService.GetConversationDetail(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40024, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetConversationAISetting(c *gin.Context) {
	result, err := h.accountService.GetConversationAISetting(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40039, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) UpdateConversationAISetting(c *gin.Context) {
	var request service.UpdateConversationAISettingInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40040, "AI 设置参数无效")
		return
	}
	result, err := h.accountService.UpdateConversationAISetting(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40041, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GenerateConversationAIDraft(c *gin.Context) {
	var request service.GenerateConversationAIDraftInput
	_ = c.ShouldBindJSON(&request)
	result, err := h.accountService.GenerateConversationAIDraft(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40042, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListConversationMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	result, err := h.accountService.ListConversationMessages(
		c.Request.Context(),
		c.Param("wxid"),
		c.Param("conversationId"),
		page,
		pageSize,
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40025, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) DeleteConversation(c *gin.Context) {
	if err := h.accountService.DeleteConversation(c.Request.Context(), c.Param("wxid"), c.Param("conversationId")); err != nil {
		response.Error(c, http.StatusBadRequest, 40038, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *AccountHandler) DownloadConversationImage(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("messageId"), 10, 64)
	if err != nil || messageID == 0 {
		response.Error(c, http.StatusBadRequest, 40036, "消息ID无效")
		return
	}
	result, err := h.accountService.DownloadConversationImage(c.Request.Context(), c.Param("wxid"), uint(messageID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40037, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SyncMessages(c *gin.Context) {
	result, err := h.accountService.SyncMessages(c.Request.Context(), c.Param("wxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40017, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendConversationText(c *gin.Context) {
	var request service.SendTextInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40026, "发送内容不能为空")
		return
	}
	result, err := h.accountService.SendConversationText(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40027, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendConversationImage(c *gin.Context) {
	var request service.SendImageInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40028, "图片内容不能为空")
		return
	}
	result, err := h.accountService.SendConversationImage(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40029, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendConversationEmoji(c *gin.Context) {
	var request service.SendEmojiInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40030, "表情参数不能为空")
		return
	}
	result, err := h.accountService.SendConversationEmoji(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40031, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ListRecentConversationEmojis(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "48"))
	result, err := h.accountService.ListRecentConversationEmojis(c.Request.Context(), c.Param("wxid"), limit)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40043, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ShareConversationCard(c *gin.Context) {
	var request service.ShareCardInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40032, "名片参数不能为空")
		return
	}
	result, err := h.accountService.SendConversationCard(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40033, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) ShareConversationLink(c *gin.Context) {
	var request service.ShareLinkInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40034, "链接参数不能为空")
		return
	}
	result, err := h.accountService.SendConversationLink(c.Request.Context(), c.Param("wxid"), c.Param("conversationId"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40035, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) DashboardOverview(c *gin.Context) {
	result, err := h.accountService.GetDashboardOverview(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40022, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendTextMessage(c *gin.Context) {
	var request service.SendTextInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40018, "发送对象和消息内容不能为空")
		return
	}
	result, err := h.accountService.SendTextMessage(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40019, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SearchFriendCandidates(c *gin.Context) {
	var request service.FriendSearchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40036, "搜索参数格式错误")
		return
	}
	result, err := h.accountService.SearchFriendCandidates(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40037, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendFriendRequest(c *gin.Context) {
	var request service.SendFriendRequestInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40038, "添加好友参数格式错误")
		return
	}
	result, err := h.accountService.SendFriendRequest(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40039, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) CheckFriendRelation(c *gin.Context) {
	result, err := h.accountService.CheckFriendRelation(c.Request.Context(), c.Param("wxid"), c.Param("targetWxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40040, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) CheckFriendRelationBatch(c *gin.Context) {
	var request service.FriendRelationBatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40043, "批量好友检测参数格式错误")
		return
	}
	result, err := h.accountService.CheckFriendRelationBatch(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40044, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) DeleteFriend(c *gin.Context) {
	result, err := h.accountService.DeleteFriend(c.Request.Context(), c.Param("wxid"), c.Param("targetWxid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40041, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SendFriendRequestBatch(c *gin.Context) {
	var request service.SendFriendRequestBatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40045, "批量添加好友参数格式错误")
		return
	}
	result, err := h.accountService.SendFriendRequestBatch(c.Request.Context(), c.Param("wxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40046, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SetFriendBlacklist(c *gin.Context) {
	var request service.FriendBlacklistInput
	_ = c.ShouldBindJSON(&request)
	result, err := h.accountService.SetFriendBlacklist(c.Request.Context(), c.Param("wxid"), c.Param("targetWxid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40042, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) RefreshGroup(c *gin.Context) {
	result, err := h.accountService.RefreshGroup(c.Request.Context(), c.Param("wxid"), c.Param("qid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40047, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) GetGroupMembers(c *gin.Context) {
	result, err := h.accountService.GetGroupMembers(c.Request.Context(), c.Param("wxid"), c.Param("qid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40048, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) UpdateGroupName(c *gin.Context) {
	var request service.UpdateGroupInfoInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40049, "群名称参数无效")
		return
	}
	result, err := h.accountService.UpdateGroupName(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40050, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) UpdateGroupAnnouncement(c *gin.Context) {
	var request service.UpdateGroupInfoInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40051, "群公告参数无效")
		return
	}
	result, err := h.accountService.UpdateGroupAnnouncement(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40052, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) UpdateGroupRemark(c *gin.Context) {
	var request service.UpdateGroupInfoInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40053, "群备注参数无效")
		return
	}
	result, err := h.accountService.UpdateGroupRemark(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) SetGroupAddressBook(c *gin.Context) {
	var request service.GroupAddressBookInput
	_ = c.ShouldBindJSON(&request)
	result, err := h.accountService.SetGroupAddressBook(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40055, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) AddGroupMembers(c *gin.Context) {
	var request service.GroupMemberMutationInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40056, "群成员参数无效")
		return
	}
	result, err := h.accountService.AddGroupMembers(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40057, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) InviteGroupMembers(c *gin.Context) {
	var request service.GroupMemberMutationInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40058, "邀请成员参数无效")
		return
	}
	result, err := h.accountService.InviteGroupMembers(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40059, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) RemoveGroupMembers(c *gin.Context) {
	var request service.GroupMemberMutationInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40060, "移除成员参数无效")
		return
	}
	result, err := h.accountService.RemoveGroupMembers(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40061, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) OperateGroupAdmin(c *gin.Context) {
	var request service.GroupAdminInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40062, "群管理员操作参数无效")
		return
	}
	result, err := h.accountService.OperateGroupAdmin(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40063, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) AddGroupFriend(c *gin.Context) {
	var request service.GroupAddFriendInput
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, 40064, "添加群好友参数无效")
		return
	}
	result, err := h.accountService.AddGroupFriend(c.Request.Context(), c.Param("wxid"), c.Param("qid"), request)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40065, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AccountHandler) QuitGroup(c *gin.Context) {
	result, err := h.accountService.QuitGroup(c.Request.Context(), c.Param("wxid"), c.Param("qid"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40066, err.Error())
		return
	}
	response.Success(c, result)
}
