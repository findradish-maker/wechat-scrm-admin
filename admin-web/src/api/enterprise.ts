import { requestClient } from '#/api/request';

export type WechatAccount = {
  alias: string;
  avatar: string;
  createdAt?: string;
  deviceName: string;
  id: number;
  lastHeartbeatAt?: null | string;
  lastLoginAt?: null | string;
  lastSyncAt?: null | string;
  mobile: string;
  nickname: string;
  platform: string;
  signature: string;
  status: string;
  updatedAt?: string;
  wxid: string;
};

export type CreateLoginSessionInput = {
  deviceName?: string;
  platform?: string;
};

export type LoginSessionState = {
  account?: null | WechatAccount;
  expiresAt?: null | string;
  message: string;
  qrBase64: string;
  qrUrl: string;
  sessionId: string;
  status: string;
  upstream?: null | Record<string, any>;
  uuid: string;
};

export type DashboardSummary = {
  accountsActive24Hours: number;
  accountsWithSync: number;
  directContacts: number;
  groups: number;
  messages24Hours: number;
  offlineAccounts: number;
  officialAccounts: number;
  onlineAccounts: number;
  totalAccounts: number;
  totalContacts: number;
  totalMessages: number;
};

export type DashboardAccountRow = {
  avatar: string;
  directContactCount: number;
  groupCount: number;
  lastHeartbeatAt?: null | string;
  lastSyncAt?: null | string;
  messageCount: number;
  nickname: string;
  officialAccountCount: number;
  status: string;
  wxid: string;
};

export type DashboardOverview = {
  accounts: DashboardAccountRow[];
  summary: DashboardSummary;
};

export type ConversationSummary = {
  accountWxid: string;
  chatRoomOwner: string;
  conversationId: string;
  conversationType: 'group' | 'single';
  groupName: string;
  lastMessage: string;
  lastMessageTime: number;
  lastMessageType: string;
  memberCount: number;
  remark: string;
  targetAvatar: string;
  targetName: string;
  targetWxid: string;
  unreadCount: number;
};

export type ConversationListResult = {
  conversations: ConversationSummary[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type ConversationDetail = {
  accountWxid: string;
  announcement: string;
  chatRoomOwner: string;
  conversationId: string;
  conversationType: 'group' | 'single';
  groupMembers?: Array<{
    chatroomMemberFlag: number;
    inviterUserName: string;
    nickName: string;
    userName: string;
  }>;
  groupName: string;
  memberCount: number;
  remark: string;
  targetAvatar: string;
  targetName: string;
  targetWxid: string;
};

export type GroupMemberItem = {
  avatar: string;
  chatroomMemberFlag: number;
  displayName: string;
  inviterUserName: string;
  nickName: string;
  userName: string;
};

export type GroupMembersResult = {
  chatRoom: string;
  members: GroupMemberItem[];
};

export type GroupActionResult = {
  action: string;
  chatRoom: string;
  message: string;
};

export type GroupMemberMutationInput = {
  targets: string[];
};

export type GroupAdminInput = {
  targets: string[];
  val: number;
};

export type GroupAddressBookInput = {
  enabled: boolean;
};

export type UpdateGroupInfoInput = {
  content: string;
};

export type GroupAddFriendInput = {
  chatRoomAccessVerifyTicket?: string;
  getContactScene?: number;
  opcode?: number;
  scene?: number;
  targets: string[];
  verifyContent?: string;
};

export type GroupAddFriendItemResult = {
  message: string;
  success: boolean;
  targetWxid: string;
  username: string;
  v1: string;
  v2: string;
  verifyData?: null | Record<string, any>;
};

export type GroupAddFriendResult = {
  chatRoom: string;
  items: GroupAddFriendItemResult[];
  opcode: number;
  scene: number;
};

export type ConversationMessage = {
  chatType: 'group' | 'single';
  content: string;
  contentMeta?: null | Record<string, any>;
  conversationId: string;
  createdAt: number;
  fromWxid: string;
  isSelf: boolean;
  messageId: number;
  messageType: string;
  msgId: number;
  senderAvatar: string;
  senderName: string;
  senderWxid: string;
  status: string;
  toWxid: string;
};

export type SendConversationEmojiInput = {
  md5: string;
  totalLen: number;
};

export type EmojiCatalogItem = {
  height: number;
  label: string;
  lastUsedAt: number;
  md5: string;
  totalLen: number;
  width: number;
};

export type ShareConversationCardInput = {
  cardAlias?: string;
  cardNickname?: string;
  cardWxid: string;
};

export type ShareConversationLinkInput = {
  description?: string;
  thumbUrl?: string;
  title?: string;
  url: string;
};

export type ConversationMessagesResult = {
  messages: ConversationMessage[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type ConversationImageResult = {
  height: number;
  messageId: number;
  src: string;
  width: number;
};

export type AIProviderSummary = {
  defaultModel: string;
  key: string;
  label: string;
  models: string[];
  supportsCustomBaseUrl: boolean;
};

export type ConversationAISetting = {
  apiKey: string;
  apiBaseUrl: string;
  conversationId: string;
  enabled: boolean;
  keywordTriggerEnabled: boolean;
  model: string;
  provider: string;
  systemPrompt: string;
  triggerKeywords: string[];
};

export type UpdateConversationAISettingInput = {
  apiKey?: string;
  apiBaseUrl?: string;
  enabled: boolean;
  keywordTriggerEnabled: boolean;
  model?: string;
  provider: string;
  systemPrompt?: string;
  triggerKeywords: string[];
};

export type GenerateConversationAIDraftResult = {
  content: string;
  model: string;
  provider: string;
};

export type ContactSummary = {
  alias: string;
  announcement: string;
  avatar: string;
  chatRoomOwner: string;
  city: string;
  contactCategory: string;
  contactCategoryLabel: string;
  contactType: 'contact' | 'group' | 'official_account';
  displayName: string;
  lastSyncedAt?: null | string;
  memberCount: number;
  nickname: string;
  province: string;
  remark: string;
  sortKey: string;
  sortLetter: string;
  signature: string;
  verifyFlag: number;
  wxid: string;
};

export type ContactListResult = {
  contacts: ContactSummary[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type FriendSearchTarget = {
  alias: string;
  avatar: string;
  canAdd: boolean;
  city: string;
  displayName: string;
  matchType: number;
  nickname: string;
  province: string;
  signature: string;
  source: string;
  v1: string;
  v2: string;
  verifyFlag: number;
  wxid: string;
};

export type FriendSearchResult = {
  keyword: string;
  results: FriendSearchTarget[];
};

export type SendFriendRequestInput = {
  opcode?: number;
  scene?: number;
  targetWxid?: string;
  v1: string;
  v2: string;
  verifyContent?: string;
};

export type SendFriendRequestResult = {
  targetWxid: string;
  username: string;
};

export type FriendRelationResult = {
  errorMessage?: string;
  headImgUrl: string;
  nickName: string;
  relationCode: number;
  relationKey: string;
  relationLabel: string;
  sign: string;
  targetWxid: string;
  upstreamRet?: number;
};

export type FriendOperationResult = {
  action: string;
  message: string;
  targetWxid: string;
};

export type FriendRelationBatchResult = {
  failed: Array<{
    reason: string;
    targetWxid: string;
  }>;
  results: FriendRelationResult[];
};

export type MomentMediaItem = {
  height: string;
  thumbUrl: string;
  totalSize: string;
  type: string;
  url: string;
  videoDuration: string;
  width: string;
};

export type MomentCommentItem = {
  commentId: number;
  content: string;
  createTime: number;
  nickname: string;
  replyUsername: string;
  type: number;
  username: string;
};

export type MomentSummary = {
  authorAvatar: string;
  authorName: string;
  authorWxid: string;
  canEdit: boolean;
  commentCount: number;
  comments: MomentCommentItem[];
  content: string;
  createdAt: number;
  id: string;
  liked: boolean;
  likeCount: number;
  likes: MomentCommentItem[];
  media: MomentMediaItem[];
  mediaCount: number;
  mediaType: string;
  preview: string;
  rawXml: string;
  visibility: 'private' | 'public';
};

export type MomentListResult = {
  firstPageMd5: string;
  hasMore: boolean;
  items: MomentSummary[];
  nextMaxId: string;
};

export type MomentDetailResult = {
  item: MomentSummary;
};

export type PublishMomentInput = {
  blackList?: string[];
  content: string;
  withUserList?: string[];
};

export type PublishMomentResult = {
  id: string;
  message: string;
};

export type OperateMomentInput = {
  action: 'delete' | 'private' | 'public';
  commentId?: number;
};

export type OperateMomentResult = {
  action: string;
  id: string;
  message: string;
};

export type FavoriteMediaItem = {
  description: string;
  thumbUrl: string;
  title: string;
  type: string;
  url: string;
};

export type FavoriteSummary = {
  favId: number;
  kind: string;
  kindLabel: string;
  media: FavoriteMediaItem[];
  mediaCount: number;
  preview: string;
  rawXml: string;
  sourceDisplay: string;
  sourceWxid: string;
  title: string;
  updatedAt: number;
};

export type FavoriteListResult = {
  hasMore: boolean;
  items: FavoriteSummary[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type FavoriteDetailResult = {
  item: FavoriteSummary;
};

export type FavoriteDeleteResult = {
  favId: number;
  message: string;
};

export type FinderProfileResult = {
  actionType: number;
  authInfo?: null | {
    appName: string;
    authIconType: number;
    authProfession: string;
    detailLink: string;
    guarantorName: string;
    guarantorWxid: string;
    realName: string;
  };
  canUse: boolean;
  coverImgUrl: string;
  extInfo: {
    birthDay: number;
    birthMonth: number;
    birthYear: number;
    city: string;
    country: string;
    province: string;
    sex: number;
  };
  followFlag: number;
  followTime: number;
  headUrl: string;
  isNonresidentFinderacctLocation: boolean;
  isNonresidentRealtimeLocation: boolean;
  isNonresidentWxacctLocation: boolean;
  message: string;
  nickname: string;
  nicknameModifyWording: string;
  originalFlag: number;
  seq: number;
  signature: string;
  spamStatus: number;
  userFlag: number;
  username: string;
  verifyInfo?: null | {
    appName: string;
    bannerWording: string;
    errScene: number;
    headImgUrl: string;
    verifyLink: string;
    verifyNickname: string;
    verifyPrefix: string;
  };
};

export type SendFriendRequestBatchInput = {
  requests: SendFriendRequestInput[];
};

export type SendFriendRequestBatchResult = {
  failed: Array<{
    reason: string;
    targetWxid: string;
  }>;
  results: SendFriendRequestResult[];
};

export type MessageArchiveItem = {
  article?: {
    cover?: string;
    extraCount?: number;
    publishTime?: number;
    publisher?: string;
    summary?: string;
    title?: string;
    url?: string;
  } | null;
  card?: {
    alias?: string;
    nickname?: string;
    wxid?: string;
  } | null;
  chatDisplay: string;
  chatWxid: string;
  content: string;
  conversationType: string;
  createTime: number;
  emoji?: {
    height?: number;
    length?: number;
    md5?: string;
    width?: number;
  } | null;
  fromWxid: string;
  image?: {
    base64?: string;
    height?: number;
    thumbUrl?: string;
    url?: string;
    width?: number;
  } | null;
  isSelf: boolean;
  kind: string;
  msgId: number;
  msgSeq: number;
  msgType: number;
  newMsgId: number;
  preview: string;
  quote?: {
    quotedBy?: string;
    quotedTitle?: string;
    title?: string;
  } | null;
  senderDisplay: string;
  senderWxid: string;
  system?: {
    action?: string;
    summary?: string;
    title?: string;
  } | null;
  toWxid: string;
  video?: {
    durationSec?: number;
    height?: number;
    length?: number;
    width?: number;
  } | null;
  voice?: {
    durationMs?: number;
    length?: number;
  } | null;
};

export type MessageArchiveResult = {
  messages: MessageArchiveItem[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type SyncMessagesResult = {
  continueFlag: number;
  keyBuf: string;
  message: string;
  messages: MessageArchiveItem[];
  status: number;
  syncedCount: number;
};

export async function getDashboardOverviewApi() {
  return requestClient.get<DashboardOverview>('/dashboard/overview');
}

export async function listAccountsApi() {
  return requestClient.get<WechatAccount[]>('/accounts');
}

export async function listAIProvidersApi() {
  return requestClient.get<AIProviderSummary[]>('/ai/providers');
}

export async function createAccountLoginSessionApi(
  payload: CreateLoginSessionInput = {},
) {
  return requestClient.post<LoginSessionState>('/accounts/login-sessions', payload);
}

export async function pollAccountLoginSessionApi(sessionId: string) {
  return requestClient.get<LoginSessionState>(
    `/accounts/login-sessions/${encodeURIComponent(sessionId)}`,
  );
}

export async function createAccountAwakenLoginSessionApi(wxid: string) {
  return requestClient.post<LoginSessionState>(
    `/accounts/${encodeURIComponent(wxid)}/awaken-login-sessions`,
  );
}

export async function bootstrapAccountApi(wxid: string) {
  return requestClient.post(`/accounts/${encodeURIComponent(wxid)}/bootstrap`);
}

export async function startAccountHeartbeatApi(wxid: string) {
  return requestClient.post(`/accounts/${encodeURIComponent(wxid)}/heartbeat/start`);
}

export async function stopAccountHeartbeatApi(wxid: string) {
  return requestClient.post(`/accounts/${encodeURIComponent(wxid)}/heartbeat/stop`);
}

export async function logoutAccountApi(wxid: string) {
  return requestClient.post(`/accounts/${encodeURIComponent(wxid)}/logout`);
}

export async function refreshGroupApi(wxid: string, qid: string) {
  return requestClient.post<ConversationDetail>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/refresh`,
  );
}

export async function listGroupMembersApi(wxid: string, qid: string) {
  return requestClient.get<GroupMembersResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/members`,
  );
}

export async function updateGroupNameApi(
  wxid: string,
  qid: string,
  payload: UpdateGroupInfoInput,
) {
  return requestClient.put<ConversationDetail>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/name`,
    payload,
  );
}

export async function updateGroupAnnouncementApi(
  wxid: string,
  qid: string,
  payload: UpdateGroupInfoInput,
) {
  return requestClient.put<ConversationDetail>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/announcement`,
    payload,
  );
}

export async function updateGroupRemarkApi(
  wxid: string,
  qid: string,
  payload: UpdateGroupInfoInput,
) {
  return requestClient.put<ConversationDetail>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/remark`,
    payload,
  );
}

export async function setGroupAddressBookApi(
  wxid: string,
  qid: string,
  payload: GroupAddressBookInput,
) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/address-book`,
    payload,
  );
}

export async function addGroupMembersApi(
  wxid: string,
  qid: string,
  payload: GroupMemberMutationInput,
) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/members/add`,
    payload,
  );
}

export async function inviteGroupMembersApi(
  wxid: string,
  qid: string,
  payload: GroupMemberMutationInput,
) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/members/invite`,
    payload,
  );
}

export async function removeGroupMembersApi(
  wxid: string,
  qid: string,
  payload: GroupMemberMutationInput,
) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/members/remove`,
    payload,
  );
}

export async function operateGroupAdminApi(
  wxid: string,
  qid: string,
  payload: GroupAdminInput,
) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/admin`,
    payload,
  );
}

export async function addGroupFriendApi(
  wxid: string,
  qid: string,
  payload: GroupAddFriendInput,
) {
  return requestClient.post<GroupAddFriendResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/add-friend`,
    payload,
  );
}

export async function quitGroupApi(wxid: string, qid: string) {
  return requestClient.post<GroupActionResult>(
    `/accounts/${encodeURIComponent(wxid)}/groups/${encodeURIComponent(qid)}/quit`,
  );
}

export async function listConversationsApi(
  wxid: string,
  params: {
    keyword?: string;
    page?: number;
    pageSize?: number;
  } = {},
) {
  return requestClient.get<ConversationListResult>(
    `/accounts/${encodeURIComponent(wxid)}/conversations`,
    { params },
  );
}

export async function getConversationDetailApi(wxid: string, conversationId: string) {
  return requestClient.get<ConversationDetail>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}`,
  );
}

export async function getConversationAISettingApi(wxid: string, conversationId: string) {
  return requestClient.get<ConversationAISetting>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/ai-setting`,
  );
}

export async function updateConversationAISettingApi(
  wxid: string,
  conversationId: string,
  payload: UpdateConversationAISettingInput,
) {
  return requestClient.put<ConversationAISetting>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/ai-setting`,
    payload,
  );
}

export async function generateConversationAIDraftApi(
  wxid: string,
  conversationId: string,
  instruction = '',
) {
  return requestClient.post<GenerateConversationAIDraftResult>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/ai-draft`,
    { instruction },
  );
}

export async function deleteConversationApi(wxid: string, conversationId: string) {
  return requestClient.delete(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}`,
  );
}

export async function listConversationMessagesApi(
  wxid: string,
  conversationId: string,
  params: {
    page?: number;
    pageSize?: number;
  } = {},
) {
  return requestClient.get<ConversationMessagesResult>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages`,
    { params },
  );
}

export async function downloadConversationImageApi(
  wxid: string,
  conversationId: string,
  messageId: number,
) {
  return requestClient.get<ConversationImageResult>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/${messageId}/image`,
  );
}

export async function sendConversationTextApi(
  wxid: string,
  conversationId: string,
  content: string,
) {
  return requestClient.post<ConversationMessage>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/text`,
    { content },
  );
}

export async function sendConversationImageApi(
  wxid: string,
  conversationId: string,
  base64: string,
) {
  return requestClient.post<ConversationMessage>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/image`,
    { base64 },
  );
}

export async function sendConversationEmojiApi(
  wxid: string,
  conversationId: string,
  payload: SendConversationEmojiInput,
) {
  return requestClient.post<ConversationMessage>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/emoji`,
    payload,
  );
}

export async function listRecentEmojisApi(wxid: string, limit = 48) {
  return requestClient.get<EmojiCatalogItem[]>(
    `/accounts/${encodeURIComponent(wxid)}/messages/emojis`,
    { params: { limit } },
  );
}

export async function shareConversationCardApi(
  wxid: string,
  conversationId: string,
  payload: ShareConversationCardInput,
) {
  return requestClient.post<ConversationMessage>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/card`,
    payload,
  );
}

export async function shareConversationLinkApi(
  wxid: string,
  conversationId: string,
  payload: ShareConversationLinkInput,
) {
  return requestClient.post<ConversationMessage>(
    `/accounts/${encodeURIComponent(wxid)}/conversations/${encodeURIComponent(conversationId)}/messages/link`,
    payload,
  );
}

export async function listContactsApi(
  wxid: string,
  params: {
    category?: string;
    contactType?: string;
    keyword?: string;
    page?: number;
    pageSize?: number;
    refresh?: boolean;
  } = {},
) {
  return requestClient.get<ContactListResult>(
    `/accounts/${encodeURIComponent(wxid)}/contacts`,
    { params },
  );
}

export async function reloadContactsApi(wxid: string) {
  return requestClient.post<{ message: string }>(
    `/accounts/${encodeURIComponent(wxid)}/contacts/reload`,
  );
}

export async function getFinderProfileApi(wxid: string) {
  return requestClient.get<FinderProfileResult>(
    `/accounts/${encodeURIComponent(wxid)}/finder/profile`,
  );
}

export async function listMomentsApi(
  wxid: string,
  params: {
    firstPageMd5?: string;
    maxId?: string;
  } = {},
) {
  return requestClient.get<MomentListResult>(
    `/accounts/${encodeURIComponent(wxid)}/moments`,
    { params },
  );
}

export async function getMomentDetailApi(
  wxid: string,
  momentId: string,
  authorWxid?: string,
) {
  return requestClient.get<MomentDetailResult>(
    `/accounts/${encodeURIComponent(wxid)}/moments/${encodeURIComponent(momentId)}`,
    { params: { authorWxid } },
  );
}

export async function publishMomentApi(
  wxid: string,
  payload: PublishMomentInput,
) {
  return requestClient.post<PublishMomentResult>(
    `/accounts/${encodeURIComponent(wxid)}/moments`,
    payload,
  );
}

export async function operateMomentApi(
  wxid: string,
  momentId: string,
  payload: OperateMomentInput,
) {
  return requestClient.post<OperateMomentResult>(
    `/accounts/${encodeURIComponent(wxid)}/moments/${encodeURIComponent(momentId)}/operation`,
    payload,
  );
}

export async function listFavoritesApi(
  wxid: string,
  params: {
    page?: number;
    pageSize?: number;
  } = {},
) {
  return requestClient.get<FavoriteListResult>(
    `/accounts/${encodeURIComponent(wxid)}/favorites`,
    { params },
  );
}

export async function getFavoriteDetailApi(wxid: string, favId: number) {
  return requestClient.get<FavoriteDetailResult>(
    `/accounts/${encodeURIComponent(wxid)}/favorites/${encodeURIComponent(String(favId))}`,
  );
}

export async function deleteFavoriteApi(wxid: string, favId: number) {
  return requestClient.post<FavoriteDeleteResult>(
    `/accounts/${encodeURIComponent(wxid)}/favorites/${encodeURIComponent(String(favId))}/delete`,
  );
}

export async function searchFriendCandidatesApi(
  wxid: string,
  keyword: string,
  payload: {
    fromScene?: number;
    searchScene?: number;
  } = {},
) {
  return requestClient.post<FriendSearchResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/search`,
    {
      fromScene: payload.fromScene ?? 0,
      keyword,
      searchScene: payload.searchScene ?? 1,
    },
  );
}

export async function sendFriendRequestApi(
  wxid: string,
  payload: SendFriendRequestInput,
) {
  return requestClient.post<SendFriendRequestResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/request`,
    payload,
  );
}

export async function sendFriendRequestBatchApi(
  wxid: string,
  payload: SendFriendRequestBatchInput,
) {
  return requestClient.post<SendFriendRequestBatchResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/request/batch`,
    payload,
  );
}

export async function checkFriendRelationApi(wxid: string, targetWxid: string) {
  return requestClient.post<FriendRelationResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/${encodeURIComponent(targetWxid)}/relation`,
  );
}

export async function checkFriendRelationBatchApi(wxid: string, userNames: string[]) {
  return requestClient.post<FriendRelationBatchResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/relation/batch`,
    { userNames },
  );
}

export async function deleteFriendApi(wxid: string, targetWxid: string) {
  return requestClient.post<FriendOperationResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/${encodeURIComponent(targetWxid)}/delete`,
  );
}

export async function setFriendBlacklistApi(
  wxid: string,
  targetWxid: string,
  action: 'add' | 'remove' = 'add',
) {
  return requestClient.post<FriendOperationResult>(
    `/accounts/${encodeURIComponent(wxid)}/friends/${encodeURIComponent(targetWxid)}/blacklist`,
    { action },
  );
}

export async function listMessageArchiveApi(
  wxid: string,
  params: {
    conversationType?: string;
    keyword?: string;
    kind?: string;
    page?: number;
    pageSize?: number;
  } = {},
) {
  return requestClient.get<MessageArchiveResult>(
    `/accounts/${encodeURIComponent(wxid)}/messages`,
    { params },
  );
}

export async function syncMessagesApi(wxid: string) {
  return requestClient.post<SyncMessagesResult>(
    `/accounts/${encodeURIComponent(wxid)}/messages/sync`,
  );
}
