<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Badge,
  Button,
  Descriptions,
  Divider,
  Dropdown,
  Drawer,
  Empty,
  Input,
  Menu,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import {
  downloadConversationImageApi,
  deleteConversationApi,
  generateConversationAIDraftApi,
  getConversationAISettingApi,
  getConversationDetailApi,
  listAIProvidersApi,
  listRecentEmojisApi,
  listConversationMessagesApi,
  listConversationsApi,
  sendConversationEmojiApi,
  sendConversationImageApi,
  sendConversationTextApi,
  shareConversationCardApi,
  shareConversationLinkApi,
  syncMessagesApi,
  updateConversationAISettingApi,
  type AIProviderSummary,
  type ConversationAISetting,
  type ConversationDetail,
  type ConversationMessage,
  type ConversationSummary,
  type EmojiCatalogItem,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

type ConversationCacheEntry = {
  items: ConversationSummary[];
  loadedPages: Set<number>;
  total: number;
  totalPages: number;
};

type MessageCacheEntry = {
  items: ConversationMessage[];
  loadedPages: Set<number>;
  total: number;
  totalPages: number;
};

type RealtimeInboundMessage = {
  category?: number;
  conversationId?: string;
  content?: string;
  msgId?: number;
  sender?: {
    id?: string;
    nickname?: string;
  };
  timestamp?: string;
};

const CONVERSATION_PAGE_SIZE = 20;
const MESSAGE_PAGE_SIZE = 30;
const DETAIL_HYDRATION_CONCURRENCY = 2;
const COMMON_EMOJI_CHOICES = ['😀', '😁', '😂', '🤣', '😊', '😍', '😘', '😭', '😡', '👍', '🙏', '🎉', '❤️', '🔥', '👀', '🥺'];
const WECHAT_EMOJI_MAP: Record<string, string> = {
  微笑: '🙂',
  撇嘴: '😒',
  色: '😍',
  发呆: '😳',
  得意: '😎',
  流泪: '😭',
  害羞: '😊',
  闭嘴: '🤐',
  睡: '😴',
  大哭: '😭',
  尴尬: '😅',
  发怒: '😠',
  调皮: '😜',
  呲牙: '😁',
  惊讶: '😮',
  难过: '😔',
  酷: '😎',
  冷汗: '😅',
  抓狂: '😫',
  吐: '🤮',
  偷笑: '🤭',
  愉快: '😄',
  白眼: '🙄',
  傲慢: '😏',
  困: '🥱',
  惊恐: '😱',
  流汗: '😓',
  憨笑: '😄',
  悠闲: '😌',
  奋斗: '💪',
  咒骂: '🤬',
  疑问: '❓',
  嘘: '🤫',
  晕: '😵',
  衰: '😞',
  骷髅: '💀',
  敲打: '👊',
  再见: '👋',
  擦汗: '😓',
  抠鼻: '🤥',
  鼓掌: '👏',
  坏笑: '😏',
  左哼哼: '😤',
  右哼哼: '😤',
  哈欠: '🥱',
  鄙视: '😒',
  委屈: '🥺',
  快哭了: '🥹',
  阴险: '😈',
  亲亲: '😘',
  吓: '😱',
  可怜: '🥺',
  菜刀: '🔪',
  西瓜: '🍉',
  啤酒: '🍺',
  咖啡: '☕',
  猪头: '🐷',
  玫瑰: '🌹',
  凋谢: '🥀',
  嘴唇: '💋',
  爱心: '❤️',
  心碎: '💔',
  蛋糕: '🎂',
  炸弹: '💣',
  便便: '💩',
  月亮: '🌙',
  太阳: '☀️',
  拥抱: '🤗',
  强: '👍',
  弱: '👎',
  握手: '🤝',
  胜利: '✌️',
  抱拳: '🙏',
  勾引: '👉',
  OK: '👌',
  跳跳: '💃',
  发抖: '🥶',
  怄火: '😤',
  转圈: '🌀',
  磕头: '🙇',
  回头: '↩️',
  跳绳: '🪢',
  激动: '🤩',
  街舞: '🕺',
  献吻: '😘',
  左太极: '☯️',
  右太极: '☯️',
  捂脸: '🤦',
  奸笑: '😏',
  机智: '🤓',
  皱眉: '😣',
  耶: '✌️',
  吃瓜: '🍉',
  加油: '⛽',
  汗: '😓',
  天啊: '😱',
  Emm: '😶',
  社会社会: '😎',
  旺柴: '🐶',
  好的: '👌',
  打脸: '🤦',
  哇: '😮',
  翻白眼: '🙄',
  666: '🔥',
  让我看看: '👀',
  叹气: '😮‍💨',
  苦涩: '😖',
  裂开: '😵‍💫',
};

const route = useRoute();
const router = useRouter();

const draft = ref('');
const searchInput = ref('');
const keyword = ref('');
const fileInputRef = ref<HTMLInputElement | null>(null);
const conversationPaneRef = ref<HTMLElement | null>(null);
const messagePaneRef = ref<HTMLElement | null>(null);

const detailDrawerOpen = ref(false);
const emojiModalOpen = ref(false);
const cardModalOpen = ref(false);
const linkModalOpen = ref(false);
const imagePreviewOpen = ref(false);
const imagePreviewSrc = ref('');
const imagePreviewLoading = ref(false);
const hdImageCache = ref<Record<number, string>>({});
const aiSettingModalOpen = ref(false);
const aiProviders = ref<AIProviderSummary[]>([]);
const aiKeywordsInput = ref('');
const emojiCatalog = ref<EmojiCatalogItem[]>([]);
const selectedEmojiKey = ref('');

const cardForm = reactive({
  cardAlias: '',
  cardNickname: '',
  cardWxid: '',
});
const linkForm = reactive({
  description: '',
  thumbUrl: '',
  title: '',
  url: '',
});
const aiSetting = reactive<ConversationAISetting>({
  apiKey: '',
  apiBaseUrl: '',
  conversationId: '',
  enabled: false,
  keywordTriggerEnabled: false,
  model: '',
  provider: 'deepseek',
  systemPrompt: '',
  triggerKeywords: [],
});

const loading = reactive({
  aiDraft: false,
  aiProviders: false,
  aiSetting: false,
  aiSettingSave: false,
  accounts: false,
  emojiCatalog: false,
  syncing: false,
  conversationDetail: false,
  conversations: false,
  conversationsMore: false,
  messages: false,
  messagesMore: false,
  sendCard: false,
  sendEmoji: false,
  sendImage: false,
  sendLink: false,
  sendText: false,
});

const realtime = reactive({
  connectError: '',
  connected: false,
  connecting: false,
  lastEventAt: 0,
  lastEventPreview: '',
  reconnectAttempt: 0,
  status: 'idle' as 'connected' | 'connecting' | 'disconnected' | 'idle' | 'reconnecting',
});

const accountStore = useEnterpriseAccountStore();
const accounts = computed(() => accountStore.accounts);
const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});

const conversationItems = ref<ConversationSummary[]>([]);
const selectedConversationId = ref('');
const conversationPage = ref(1);
const conversationHasMore = ref(false);

const conversationDetail = ref<ConversationDetail | null>(null);
const messageItems = ref<ConversationMessage[]>([]);
const messagePage = ref(1);
const messageHasMore = ref(false);

const conversationCache = ref<Record<string, ConversationCacheEntry>>({});
const messageCache = ref<Record<string, MessageCacheEntry>>({});
const detailCache = ref<Record<string, ConversationDetail>>({});
const localUnreadState = ref<Record<string, number>>({});
const detailHydrationQueue = ref<Array<{ conversationId: string; wxid: string }>>([]);
const detailHydrationPending = new Set<string>();
let activeDetailHydrationCount = 0;
let realtimeSocket: WebSocket | null = null;
let realtimeSocketNonce = 0;
let realtimeRefreshPromise: null | Promise<void> = null;
let realtimeRefreshQueued = false;
let reconnectTimer: null | ReturnType<typeof window.setTimeout> = null;
let refreshTimer: null | ReturnType<typeof window.setTimeout> = null;
let workspaceActivationNonce = 0;
let aiSettingLoadNonce = 0;

const currentConversation = computed(() =>
  conversationItems.value.find((item) => item.conversationId === selectedConversationId.value),
);

const currentAccount = computed(() => accountStore.currentAccount);

const realtimeStatusLabel = computed(() => {
  switch (realtime.status) {
    case 'connected': {
      return '实时已连接';
    }
    case 'connecting': {
      return '实时连接中';
    }
    case 'reconnecting': {
      return '实时重连中';
    }
    case 'disconnected': {
      return '实时已断开';
    }
    default: {
      return '实时未连接';
    }
  }
});

const realtimeStatusColor = computed(() => {
  switch (realtime.status) {
    case 'connected': {
      return 'success';
    }
    case 'connecting':
    case 'reconnecting': {
      return 'processing';
    }
    case 'disconnected': {
      return 'warning';
    }
    default: {
      return 'default';
    }
  }
});

const accountOptions = computed(() => accountStore.accountOptions);
const aiProviderOptions = computed(() =>
  aiProviders.value.map((item) => ({
    label: item.label,
    value: item.key,
  })),
);
const selectedAIProvider = computed(() =>
  aiProviders.value.find((item) => item.key === aiSetting.provider) || null,
);
const aiModelOptions = computed(() =>
  (selectedAIProvider.value?.models || []).map((item) => ({
    label: item,
    value: item,
  })),
);
const aiAdvancedDisabled = computed(() => !aiSetting.enabled);
const aiCustomBaseURLDisabled = computed(
  () => aiAdvancedDisabled.value || !selectedAIProvider.value?.supportsCustomBaseUrl,
);
const selectedEmojiItem = computed(
  () => emojiCatalog.value.find((item) => `${item.md5}:${item.totalLen}` === selectedEmojiKey.value) || null,
);

const groupMembers = computed(() => conversationDetail.value?.groupMembers ?? []);

function getConversationCacheKey(wxid: string, nextKeyword: string) {
  return `${wxid}::${nextKeyword.trim().toLowerCase()}`;
}

function getMessageCacheKey(wxid: string, conversationId: string) {
  return `${wxid}::${conversationId}`;
}

function toMillis(timestamp: number) {
  if (!timestamp) return 0;
  return timestamp > 10_000_000_000 ? timestamp : timestamp * 1000;
}

function formatClock(timestamp: number) {
  if (!timestamp) return '';
  const date = new Date(toMillis(timestamp));
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatConversationTime(timestamp: number) {
  if (!timestamp) return '';
  const date = new Date(toMillis(timestamp));
  const now = new Date();
  if (date.toDateString() === now.toDateString()) {
    return formatClock(timestamp);
  }
  return date.toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
  });
}

function parseWechatEmojiText(content: string) {
  if (!content) return '';
  return content.replace(/\[([^[\]]+)\]/g, (full, name) => {
    const emoji = WECHAT_EMOJI_MAP[String(name).trim()];
    return emoji ? `${emoji}` : full;
  });
}

function mergeConversations(existing: ConversationSummary[], incoming: ConversationSummary[]) {
  const map = new Map<string, ConversationSummary>();
  for (const item of existing) map.set(item.conversationId, item);
  for (const item of incoming) {
    const current = map.get(item.conversationId);
    const localUnread = localUnreadState.value[item.conversationId] || 0;
    map.set(item.conversationId, {
      ...current,
      ...item,
      lastMessage: parseWechatEmojiText(item.lastMessage || current?.lastMessage || ''),
      unreadCount: Math.max(item.unreadCount || 0, current?.unreadCount || 0, localUnread),
    });
  }
  return [...map.values()].sort((a, b) => b.lastMessageTime - a.lastMessageTime);
}

function mergeMessages(existing: ConversationMessage[], incoming: ConversationMessage[]) {
  const seen = new Set<string>();
  const merged: ConversationMessage[] = [];
  for (const item of [...incoming, ...existing]) {
    const key = `${item.messageId}-${item.msgId}-${item.createdAt}`;
    if (seen.has(key)) continue;
    seen.add(key);
    merged.push(item);
  }
  return merged.sort((a, b) => {
    if (a.createdAt === b.createdAt) {
      return a.messageId - b.messageId;
    }
    return a.createdAt - b.createdAt;
  });
}

function mergeConversationSummaryDetail(
  summary: ConversationSummary,
  detail: ConversationDetail,
): ConversationSummary {
  return {
    ...summary,
    chatRoomOwner: detail.chatRoomOwner || summary.chatRoomOwner,
    groupName: detail.groupName || summary.groupName || detail.targetName || summary.targetName,
    memberCount: detail.memberCount || summary.memberCount,
    remark: detail.remark || summary.remark,
    targetAvatar: detail.targetAvatar || summary.targetAvatar,
    targetName: detail.targetName || summary.targetName,
    targetWxid: detail.targetWxid || summary.targetWxid,
  };
}

function updateConversationSummaryInStores(
  wxid: string,
  conversationId: string,
  updater: (item: ConversationSummary) => ConversationSummary,
) {
  conversationItems.value = conversationItems.value.map((item) =>
    item.conversationId === conversationId ? updater(item) : item,
  );

  for (const key of Object.keys(conversationCache.value)) {
    if (!key.startsWith(`${wxid}::`)) continue;
    const entry = conversationCache.value[key];
    if (!entry) continue;
    conversationCache.value[key] = {
      ...entry,
      items: entry.items.map((item) =>
        item.conversationId === conversationId ? updater(item) : item,
      ),
    };
  }
}

function applyConversationDetailToStores(wxid: string, detail: ConversationDetail) {
  const key = getMessageCacheKey(wxid, detail.conversationId);
  detailCache.value[key] = detail;
  updateConversationSummaryInStores(wxid, detail.conversationId, (item) =>
    mergeConversationSummaryDetail(item, detail),
  );

  if (selectedWxid.value === wxid && selectedConversationId.value === detail.conversationId) {
    conversationDetail.value = detail;
  }
}

function asRecord(value: unknown) {
  if (!value || typeof value !== 'object') return null;
  return value as Record<string, any>;
}

function imageFromMeta(item: ConversationMessage) {
  const meta = asRecord(item.contentMeta);
  if (!meta) return '';
  const base64 = String(meta.base64 || '').trim();
  if (base64) {
    if (base64.startsWith('data:')) {
      return base64;
    }
    return `data:image/jpeg;base64,${base64}`;
  }

  const thumbUrl = String(meta.thumbUrl || '').trim();
  if (thumbUrl.startsWith('data:') || thumbUrl.startsWith('http://') || thumbUrl.startsWith('https://')) {
    return thumbUrl;
  }

  const url = String(meta.url || '').trim();
  if (url.startsWith('data:') || url.startsWith('http://') || url.startsWith('https://')) {
    return url;
  }

  return '';
}

async function openImagePreview(item: ConversationMessage) {
  const src = imageFromMeta(item);
  if (!src) return;
  imagePreviewSrc.value = src;
  imagePreviewOpen.value = true;

  const cached = hdImageCache.value[item.messageId];
  if (cached) {
    imagePreviewSrc.value = cached;
    return;
  }
  if (!selectedWxid.value || !item.conversationId || !item.messageId) return;

  imagePreviewLoading.value = true;
  try {
    const response = await downloadConversationImageApi(
      selectedWxid.value,
      item.conversationId,
      item.messageId,
    );
    if (response?.src) {
      hdImageCache.value[item.messageId] = response.src;
      imagePreviewSrc.value = response.src;
    }
  } catch (error) {
    console.error(error);
  } finally {
    imagePreviewLoading.value = false;
  }
}

function linkFromMeta(item: ConversationMessage) {
  const meta = asRecord(item.contentMeta);
  if (!meta) return null;
  return {
    cover: String(meta.cover || meta.thumbUrl || ''),
    summary: String(meta.summary || meta.description || ''),
    title: String(meta.title || ''),
    url: String(meta.url || ''),
  };
}

function cardFromMeta(item: ConversationMessage) {
  const meta = asRecord(item.contentMeta);
  if (!meta) return null;
  return {
    alias: String(meta.alias || ''),
    nickname: String(meta.nickname || ''),
    wxid: String(meta.wxid || ''),
  };
}

function emojiFromMeta(item: ConversationMessage) {
  const meta = asRecord(item.contentMeta);
  if (!meta) return null;
  return {
    length: Number(meta.length || 0),
    md5: String(meta.md5 || ''),
  };
}

function previewFromMessageType(item: ConversationMessage) {
  if (item.messageType === 'text') {
    return parseWechatEmojiText(item.content || '文本消息');
  }
  if (item.messageType === 'image') return '[图片]';
  if (item.messageType === 'emoji') return '[表情]';
  if (item.messageType === 'card') return '[名片]';
  if (item.messageType === 'link') return '[链接]';
  if (item.messageType === 'video') return '[视频]';
  if (item.messageType === 'voice') return '[语音]';
  return parseWechatEmojiText(item.content || '[消息]');
}

function getRealtimeWsCandidates(wxid: string) {
  if (typeof window === 'undefined' || !wxid) return [];
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const encodedWxid = encodeURIComponent(wxid);
  const sameOriginUrl = `${protocol}//${window.location.host}/ws/${encodedWxid}`;
  const fallbackUrl = `${protocol}//${window.location.hostname}:8080/ws/${encodedWxid}`;
  return [...new Set([sameOriginUrl, fallbackUrl])];
}

function parseRealtimePayload(payload: unknown) {
  const text = String(payload ?? '').trim();
  if (!text) return [] as RealtimeInboundMessage[];

  return text
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      try {
        return JSON.parse(item) as RealtimeInboundMessage;
      } catch {
        return {
          content: item,
        } as RealtimeInboundMessage;
      }
    });
}

function formatRealtimePreview(events: RealtimeInboundMessage[]) {
  const latest = events[events.length - 1];
  if (!latest) return '';
  const sender = latest.sender?.nickname || latest.sender?.id || '新消息';
  const content = parseWechatEmojiText(String(latest.content || '').trim());
  if (!content) return sender;
  return `${sender}: ${content}`;
}

function cleanupRealtimeTimers() {
  if (reconnectTimer) {
    window.clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
    refreshTimer = null;
  }
}

function closeRealtimeSocket(options?: { preserveStatus?: boolean }) {
  cleanupRealtimeTimers();
  realtimeSocketNonce += 1;
  if (realtimeSocket) {
    realtimeSocket.onclose = null;
    realtimeSocket.onerror = null;
    realtimeSocket.onmessage = null;
    realtimeSocket.onopen = null;
    realtimeSocket.close();
    realtimeSocket = null;
  }
  realtime.connected = false;
  realtime.connecting = false;
  if (!options?.preserveStatus) {
    realtime.connectError = '';
    realtime.reconnectAttempt = 0;
    realtime.status = 'idle';
  }
}

function setConversationUnreadCount(wxid: string, conversationId: string, count: number) {
  localUnreadState.value = {
    ...localUnreadState.value,
    [conversationId]: Math.max(0, count),
  };
  updateConversationSummaryInStores(wxid, conversationId, (item) => ({
    ...item,
    unreadCount: Math.max(0, count),
  }));
}

function increaseConversationUnread(wxid: string, conversationId: string) {
  const nextUnread = Math.max(0, (localUnreadState.value[conversationId] || 0) + 1);
  localUnreadState.value = {
    ...localUnreadState.value,
    [conversationId]: nextUnread,
  };
  updateConversationSummaryInStores(wxid, conversationId, (item) => ({
    ...item,
    unreadCount: Math.max(nextUnread, (item.unreadCount || 0) + 1),
  }));
}

function resetMessageComposer() {
  draft.value = '';
}

function scrollMessagesToBottom() {
  nextTick(() => {
    if (!messagePaneRef.value) return;
    messagePaneRef.value.scrollTop = messagePaneRef.value.scrollHeight;
  });
}

function syncConversationCaches(wxid: string, conversation: ConversationSummary) {
  updateConversationSummaryInStores(wxid, conversation.conversationId, () => conversation);
}

function upsertSentMessage(sent: ConversationMessage) {
  const cacheKey = getMessageCacheKey(selectedWxid.value, sent.conversationId);
  const currentCache = messageCache.value[cacheKey];
  const nextItems = mergeMessages(currentCache?.items ?? messageItems.value, [sent]);
  const total = Math.max(currentCache?.total ?? 0, nextItems.length);
  const nextCache: MessageCacheEntry = {
    items: nextItems,
    loadedPages: new Set([...(currentCache?.loadedPages ?? []), 1]),
    total,
    totalPages: Math.max(1, Math.ceil(total / MESSAGE_PAGE_SIZE)),
  };

  messageCache.value[cacheKey] = nextCache;
  messageItems.value = nextItems;

  const nextConversation = (currentConversation.value
    ? {
        ...currentConversation.value,
        lastMessage: previewFromMessageType(sent),
        lastMessageTime: Math.floor(Date.now() / 1000),
        lastMessageType: sent.messageType,
        unreadCount: 0,
      }
    : null) as ConversationSummary | null;

  if (nextConversation) {
    conversationItems.value = mergeConversations(conversationItems.value, [nextConversation]);
    syncConversationCaches(selectedWxid.value, nextConversation);
  }

  scrollMessagesToBottom();
}

async function loadAccounts() {
  loading.accounts = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loading.accounts = false;
  }
}

async function loadAIProviders() {
  if (loading.aiProviders || aiProviders.value.length) return;
  loading.aiProviders = true;
  try {
    aiProviders.value = await listAIProvidersApi();
  } finally {
    loading.aiProviders = false;
  }
}

function applyConversationAISetting(setting: ConversationAISetting) {
  aiSetting.apiKey = setting.apiKey || '';
  aiSetting.apiBaseUrl = setting.apiBaseUrl || '';
  aiSetting.conversationId = setting.conversationId;
  aiSetting.enabled = setting.enabled;
  aiSetting.keywordTriggerEnabled = setting.keywordTriggerEnabled;
  aiSetting.model = setting.model;
  aiSetting.provider = setting.provider;
  aiSetting.systemPrompt = setting.systemPrompt;
  aiSetting.triggerKeywords = [...setting.triggerKeywords];
  aiKeywordsInput.value = setting.triggerKeywords.join('\n');
}

async function loadCurrentConversationAISetting(wxid: string, conversationId: string) {
  if (!wxid || !conversationId) return;
  const nonce = ++aiSettingLoadNonce;
  await loadAIProviders();
  loading.aiSetting = true;
  try {
    const setting = await getConversationAISettingApi(wxid, conversationId);
    if (
      nonce !== aiSettingLoadNonce
      || selectedWxid.value !== wxid
      || selectedConversationId.value !== conversationId
    ) {
      return;
    }
    applyConversationAISetting(setting);
  } finally {
    if (nonce === aiSettingLoadNonce) {
      loading.aiSetting = false;
    }
  }
}

async function syncWorkspaceRoute(wxid: string, conversationId = '') {
  const currentWxid = typeof route.query.wxid === 'string' ? route.query.wxid : '';
  const currentConversation =
    typeof route.query.conversation === 'string' ? route.query.conversation : '';
  if (
    route.name === 'EnterpriseMessages'
    && currentWxid === wxid
    && currentConversation === conversationId
  ) {
    return;
  }
  const targetPath = route.path === '/dashboard/messages'
    ? '/dashboard/messages'
    : '/enterprise/messages';
  await router.replace({
    path: targetPath,
    query: {
      ...route.query,
      conversation: conversationId || undefined,
      wxid: wxid || undefined,
    },
  });
}

async function ensureConversationVisible(wxid: string, conversationId: string) {
  if (!wxid || !conversationId) return;
  if (conversationItems.value.some((item) => item.conversationId === conversationId)) return;

  const detail = await getConversationDetailApi(wxid, conversationId);
  applyConversationDetailToStores(wxid, detail);

  const placeholder: ConversationSummary = {
    accountWxid: wxid,
    chatRoomOwner: detail.chatRoomOwner || '',
    conversationId: detail.conversationId,
    conversationType: detail.conversationType,
    groupName: detail.groupName || '',
    lastMessage: '',
    lastMessageTime: 0,
    lastMessageType: 'text',
    memberCount: detail.memberCount || 0,
    remark: detail.remark || '',
    targetAvatar: detail.targetAvatar || '',
    targetName: detail.targetName || detail.groupName || detail.targetWxid,
    targetWxid: detail.targetWxid,
    unreadCount: 0,
  };

  const nextItems = mergeConversations([placeholder], conversationItems.value);
  conversationItems.value = nextItems;

  const cacheKey = getConversationCacheKey(wxid, keyword.value);
  const cache = conversationCache.value[cacheKey];
  if (cache) {
    conversationCache.value[cacheKey] = {
      ...cache,
      items: mergeConversations([placeholder], cache.items),
      total: Math.max(cache.total, nextItems.length),
      totalPages: Math.max(cache.totalPages, Math.ceil(Math.max(cache.total, nextItems.length) / CONVERSATION_PAGE_SIZE)),
    };
  }
}

function resetConversationWorkspace() {
  closeRealtimeSocket();
  detailHydrationQueue.value = [];
  detailHydrationPending.clear();
  activeDetailHydrationCount = 0;
  conversationItems.value = [];
  selectedConversationId.value = '';
  conversationDetail.value = null;
  messageItems.value = [];
  conversationPage.value = 1;
  conversationHasMore.value = false;
  messagePage.value = 1;
  messageHasMore.value = false;
  realtime.lastEventAt = 0;
  realtime.lastEventPreview = '';
}

async function syncMessagesBeforeLoading(wxid: string, quiet = false) {
  if (!wxid) return;
  loading.syncing = true;
  try {
    await syncMessagesApi(wxid);
  } catch (error) {
    console.error(error);
    if (!quiet) {
      message.warning('消息同步失败，已回退到本地缓存继续加载');
    }
  } finally {
    loading.syncing = false;
  }
}

async function syncMessagesInBackground(
  wxid: string,
  options?: {
    quiet?: boolean;
    refreshCurrentConversationId?: string;
  },
) {
  if (!wxid) return;
  try {
    await syncMessagesBeforeLoading(wxid, options?.quiet ?? true);
    if (selectedWxid.value !== wxid) return;
    await refreshConversationPageOne(wxid);
    const activeConversationId =
      options?.refreshCurrentConversationId || selectedConversationId.value || conversationItems.value[0]?.conversationId || '';
    if (!selectedConversationId.value && activeConversationId) {
      await selectConversation(activeConversationId, false);
      await syncWorkspaceRoute(wxid, activeConversationId);
      return;
    }
    if (activeConversationId) {
      await refreshCurrentConversationMessages(wxid, activeConversationId);
    }
  } catch (error) {
    console.error(error);
  }
}

async function refreshConversationPageOne(wxid: string) {
  if (!wxid || selectedWxid.value !== wxid) return;
  const data = await listConversationsApi(wxid, {
    keyword: keyword.value.trim() || undefined,
    page: 1,
    pageSize: CONVERSATION_PAGE_SIZE,
  });
  const cacheKey = getConversationCacheKey(wxid, keyword.value);
  const previous = conversationCache.value[cacheKey];
  const nextEntry: ConversationCacheEntry = {
    items: mergeConversations(previous?.items ?? [], data.conversations),
    loadedPages: new Set([...(previous?.loadedPages ?? []), 1]),
    total: data.total,
    totalPages: data.totalPages,
  };
  conversationCache.value[cacheKey] = nextEntry;
  conversationItems.value = nextEntry.items;
  conversationPage.value = Math.max(...nextEntry.loadedPages);
  conversationHasMore.value = conversationPage.value < nextEntry.totalPages;
  queueConversationDetailHydration(wxid, data.conversations);
}

async function refreshCurrentConversationMessages(wxid: string, conversationId: string) {
  if (!wxid || !conversationId || selectedWxid.value !== wxid) return;

  const [detail, data] = await Promise.all([
    getConversationDetailApi(wxid, conversationId),
    listConversationMessagesApi(wxid, conversationId, {
      page: 1,
      pageSize: MESSAGE_PAGE_SIZE,
    }),
  ]);

  applyConversationDetailToStores(wxid, detail);

  const cacheKey = getMessageCacheKey(wxid, conversationId);
  const previous = messageCache.value[cacheKey];
  const nextItems = mergeMessages(previous?.items ?? [], data.messages);
  const nextEntry: MessageCacheEntry = {
    items: nextItems,
    loadedPages: new Set([...(previous?.loadedPages ?? []), 1]),
    total: Math.max(data.total, nextItems.length),
    totalPages: Math.max(data.totalPages, Math.ceil(nextItems.length / MESSAGE_PAGE_SIZE)),
  };

  messageCache.value[cacheKey] = nextEntry;
  if (selectedConversationId.value === conversationId) {
    messageItems.value = nextItems;
    messagePage.value = Math.max(...nextEntry.loadedPages);
    messageHasMore.value = messagePage.value < nextEntry.totalPages;
    scrollMessagesToBottom();
  }
}

async function refreshAfterRealtimeMessage(
  wxid: string,
  options?: { refreshCurrentConversation?: boolean },
) {
  if (!wxid) return;
  if (realtimeRefreshPromise) {
    realtimeRefreshQueued = true;
    return realtimeRefreshPromise;
  }

  realtimeRefreshPromise = (async () => {
    try {
      await refreshConversationPageOne(wxid);
      if (options?.refreshCurrentConversation && selectedConversationId.value) {
        await refreshCurrentConversationMessages(wxid, selectedConversationId.value);
      }
    } finally {
      realtimeRefreshPromise = null;
      if (realtimeRefreshQueued) {
        realtimeRefreshQueued = false;
        if (selectedWxid.value === wxid) {
          await refreshAfterRealtimeMessage(wxid, options);
        }
      }
    }
  })();

  return realtimeRefreshPromise;
}

function scheduleRealtimeRefresh(
  wxid: string,
  options?: { refreshCurrentConversation?: boolean },
) {
  if (!wxid) return;
  if (refreshTimer) {
    window.clearTimeout(refreshTimer);
  }
  refreshTimer = window.setTimeout(() => {
    refreshTimer = null;
    void refreshAfterRealtimeMessage(wxid, options);
  }, 450);
}

function scheduleRealtimeReconnect(wxid: string) {
  if (!wxid || selectedWxid.value !== wxid) return;
  if (reconnectTimer) return;
  realtime.status = 'reconnecting';
  realtime.reconnectAttempt += 1;
  const delay = Math.min(10_000, Math.max(1500, realtime.reconnectAttempt * 1500));
  reconnectTimer = window.setTimeout(() => {
    reconnectTimer = null;
    if (selectedWxid.value === wxid) {
      connectRealtimeSocket(wxid);
    }
  }, delay);
}

function connectRealtimeSocket(wxid: string) {
  if (!wxid) return;
  closeRealtimeSocket({ preserveStatus: true });
  const candidates = getRealtimeWsCandidates(wxid);
  if (!candidates.length) return;

  const nonce = ++realtimeSocketNonce;
  realtime.connectError = '';
  realtime.connecting = true;
  realtime.connected = false;
  realtime.status = 'connecting';

  const tryConnect = (index: number) => {
    if (nonce !== realtimeSocketNonce || selectedWxid.value !== wxid) return;
    if (index >= candidates.length) {
      realtime.connecting = false;
      realtime.connected = false;
      realtime.status = 'disconnected';
      scheduleRealtimeReconnect(wxid);
      return;
    }

    const socket = new WebSocket(candidates[index]!);
    realtimeSocket = socket;
    let opened = false;

    socket.onopen = () => {
      if (nonce !== realtimeSocketNonce || selectedWxid.value !== wxid) {
        socket.close();
        return;
      }
      opened = true;
      realtime.connectError = '';
      realtime.connected = true;
      realtime.connecting = false;
      realtime.reconnectAttempt = 0;
      realtime.status = 'connected';
    };

    socket.onmessage = (event) => {
      if (nonce !== realtimeSocketNonce || selectedWxid.value !== wxid) return;
      const events = parseRealtimePayload(event.data);
      realtime.lastEventAt = Date.now();
      realtime.lastEventPreview = formatRealtimePreview(events);
      for (const item of events) {
        if (!item.conversationId) continue;
        if (item.conversationId === selectedConversationId.value) {
          setConversationUnreadCount(wxid, item.conversationId, 0);
          continue;
        }
        increaseConversationUnread(wxid, item.conversationId);
      }
      const refreshCurrentConversation = Boolean(
        selectedConversationId.value
        && events.some((item) => item.conversationId === selectedConversationId.value),
      );
      scheduleRealtimeRefresh(wxid, {
        refreshCurrentConversation,
      });
    };

    socket.onerror = () => {
      realtime.connectError = '实时连接失败';
    };

    socket.onclose = () => {
      if (nonce !== realtimeSocketNonce) return;
      realtime.connected = false;
      realtime.connecting = false;
      realtimeSocket = null;
      if (!opened) {
        tryConnect(index + 1);
        return;
      }
      realtime.status = 'disconnected';
      scheduleRealtimeReconnect(wxid);
    };
  };

  tryConnect(0);
}

function queueConversationDetailHydration(
  wxid: string,
  conversations: ConversationSummary[],
) {
  for (const item of conversations) {
    const key = getMessageCacheKey(wxid, item.conversationId);
    if (detailCache.value[key] || detailHydrationPending.has(key)) continue;
    detailHydrationPending.add(key);
    detailHydrationQueue.value.push({
      conversationId: item.conversationId,
      wxid,
    });
  }
  void pumpConversationDetailHydration();
}

async function pumpConversationDetailHydration() {
  while (
    activeDetailHydrationCount < DETAIL_HYDRATION_CONCURRENCY
    && detailHydrationQueue.value.length
  ) {
    const task = detailHydrationQueue.value.shift();
    if (!task) return;

    activeDetailHydrationCount += 1;
    const key = getMessageCacheKey(task.wxid, task.conversationId);

    void getConversationDetailApi(task.wxid, task.conversationId)
      .then((detail) => {
        applyConversationDetailToStores(task.wxid, detail);
      })
      .catch((error) => {
        console.error(error);
      })
      .finally(() => {
        detailHydrationPending.delete(key);
        activeDetailHydrationCount = Math.max(0, activeDetailHydrationCount - 1);
        if (detailHydrationQueue.value.length) {
          void pumpConversationDetailHydration();
        }
      });
  }
}

async function loadConversations(options?: {
  append?: boolean;
  force?: boolean;
  keyword?: string;
  page?: number;
  wxid?: string;
}) {
  const wxid = options?.wxid ?? selectedWxid.value;
  if (!wxid) return;

  const page = options?.page ?? 1;
  const append = options?.append ?? false;
  const force = options?.force ?? false;
  const nextKeyword = options?.keyword ?? keyword.value;
  const cacheKey = getConversationCacheKey(wxid, nextKeyword);
  const cache = conversationCache.value[cacheKey];

  if (!force && cache?.loadedPages.has(page)) {
    conversationItems.value = cache.items;
    conversationPage.value = Math.max(...cache.loadedPages);
    conversationHasMore.value = conversationPage.value < cache.totalPages;
    const start = (page - 1) * CONVERSATION_PAGE_SIZE;
    const end = start + CONVERSATION_PAGE_SIZE;
    queueConversationDetailHydration(wxid, cache.items.slice(start, end));
    return;
  }

  if (append) loading.conversationsMore = true;
  else loading.conversations = true;

  try {
    const data = await listConversationsApi(wxid, {
      keyword: nextKeyword.trim() || undefined,
      page,
      pageSize: CONVERSATION_PAGE_SIZE,
    });
    const previous = force && page === 1 ? undefined : conversationCache.value[cacheKey];
    const nextEntry: ConversationCacheEntry = {
      items: mergeConversations(previous?.items ?? [], data.conversations),
      loadedPages: new Set([...(previous?.loadedPages ?? []), page]),
      total: data.total,
      totalPages: data.totalPages,
    };
    conversationCache.value[cacheKey] = nextEntry;
    conversationItems.value = nextEntry.items;
    conversationPage.value = Math.max(...nextEntry.loadedPages);
    conversationHasMore.value = conversationPage.value < nextEntry.totalPages;
    queueConversationDetailHydration(wxid, data.conversations);
  } finally {
    loading.conversations = false;
    loading.conversationsMore = false;
  }
}

async function loadConversationDetail(wxid: string, conversationId: string, force = false) {
  if (!wxid || !conversationId) return;
  const key = getMessageCacheKey(wxid, conversationId);
  if (!force && detailCache.value[key]) {
    conversationDetail.value = detailCache.value[key];
    return;
  }
  loading.conversationDetail = true;
  try {
    const detail = await getConversationDetailApi(wxid, conversationId);
    detailCache.value[key] = detail;
    conversationDetail.value = detail;
  } finally {
    loading.conversationDetail = false;
  }
}

async function loadMessages(options?: {
  append?: boolean;
  conversationId?: string;
  force?: boolean;
  page?: number;
  wxid?: string;
}) {
  const wxid = options?.wxid ?? selectedWxid.value;
  const conversationId = options?.conversationId ?? selectedConversationId.value;
  if (!wxid || !conversationId) return;

  const page = options?.page ?? 1;
  const append = options?.append ?? false;
  const force = options?.force ?? false;
  const cacheKey = getMessageCacheKey(wxid, conversationId);
  const cache = messageCache.value[cacheKey];

  if (!force && cache?.loadedPages.has(page)) {
    messageItems.value = cache.items;
    messagePage.value = Math.max(...cache.loadedPages);
    messageHasMore.value = messagePage.value < cache.totalPages;
    if (!append) scrollMessagesToBottom();
    return;
  }

  let beforeHeight = 0;
  let beforeTop = 0;
  if (append && messagePaneRef.value) {
    beforeHeight = messagePaneRef.value.scrollHeight;
    beforeTop = messagePaneRef.value.scrollTop;
  }

  if (append) loading.messagesMore = true;
  else loading.messages = true;

  try {
    const data = await listConversationMessagesApi(wxid, conversationId, {
      page,
      pageSize: MESSAGE_PAGE_SIZE,
    });
    const previous = force && page === 1 ? undefined : messageCache.value[cacheKey];
    const nextEntry: MessageCacheEntry = {
      items: mergeMessages(previous?.items ?? [], data.messages),
      loadedPages: new Set([...(previous?.loadedPages ?? []), page]),
      total: data.total,
      totalPages: data.totalPages,
    };
    messageCache.value[cacheKey] = nextEntry;
    messageItems.value = nextEntry.items;
    messagePage.value = Math.max(...nextEntry.loadedPages);
    messageHasMore.value = messagePage.value < nextEntry.totalPages;

    if (append && messagePaneRef.value) {
      nextTick(() => {
        if (!messagePaneRef.value) return;
        const afterHeight = messagePaneRef.value.scrollHeight;
        messagePaneRef.value.scrollTop = afterHeight - beforeHeight + beforeTop;
      });
    } else {
      scrollMessagesToBottom();
    }
  } finally {
    loading.messages = false;
    loading.messagesMore = false;
  }
}

async function selectConversation(conversationId: string, syncRoute = true) {
  if (!selectedWxid.value || !conversationId) return;
  if (selectedConversationId.value !== conversationId) {
    selectedConversationId.value = conversationId;
  }
  setConversationUnreadCount(selectedWxid.value, conversationId, 0);

  if (syncRoute) {
    await syncWorkspaceRoute(selectedWxid.value, conversationId);
  }

  await Promise.all([
    loadConversationDetail(selectedWxid.value, conversationId),
    loadCurrentConversationAISetting(selectedWxid.value, conversationId),
    loadMessages({
      conversationId,
      force: true,
      page: 1,
      wxid: selectedWxid.value,
    }),
  ]);
}

async function activateAccountWorkspace(
  wxid: string,
  options?: {
    preferredConversationId?: string;
    quietSync?: boolean;
  },
) {
  const activationNonce = ++workspaceActivationNonce;
  if (!wxid) {
    resetConversationWorkspace();
    await syncWorkspaceRoute('', '');
    return;
  }

  resetConversationWorkspace();
  await loadConversations({
    force: true,
    keyword: keyword.value,
    page: 1,
    wxid,
  });
  if (activationNonce !== workspaceActivationNonce || selectedWxid.value !== wxid) return;

  const preferredConversationId = options?.preferredConversationId || '';
  const selectedFromUrl =
    preferredConversationId
    && conversationItems.value.some((item) => item.conversationId === preferredConversationId)
      ? preferredConversationId
      : '';
  if (preferredConversationId && !selectedFromUrl) {
    await ensureConversationVisible(wxid, preferredConversationId);
  }
  const selectedFromPrepared =
    preferredConversationId
    && conversationItems.value.some((item) => item.conversationId === preferredConversationId)
      ? preferredConversationId
      : '';
  const fallback = conversationItems.value[0]?.conversationId || '';
  const targetConversationId = selectedFromUrl || selectedFromPrepared || fallback;

  if (targetConversationId) {
    await selectConversation(targetConversationId, false);
    await syncWorkspaceRoute(wxid, targetConversationId);
  } else {
    await syncWorkspaceRoute(wxid, '');
  }

  if (activationNonce !== workspaceActivationNonce || selectedWxid.value !== wxid) return;
  connectRealtimeSocket(wxid);
  void syncMessagesInBackground(wxid, {
    quiet: options?.quietSync ?? true,
    refreshCurrentConversationId: targetConversationId || undefined,
  });
}

async function searchConversations() {
  keyword.value = searchInput.value.trim();
  selectedConversationId.value = '';
  conversationDetail.value = null;
  messageItems.value = [];
  await loadConversations({
    force: true,
    keyword: keyword.value,
    page: 1,
  });
  const firstConversation = conversationItems.value[0]?.conversationId;
  if (firstConversation) {
    await selectConversation(firstConversation);
  }
}

async function loadMoreConversations() {
  if (!selectedWxid.value || !conversationHasMore.value || loading.conversationsMore || loading.conversations) return;
  await loadConversations({
    append: true,
    keyword: keyword.value,
    page: conversationPage.value + 1,
  });
}

async function loadOlderMessages() {
  if (!selectedWxid.value || !selectedConversationId.value || !messageHasMore.value || loading.messagesMore || loading.messages) return;
  await loadMessages({
    append: true,
    conversationId: selectedConversationId.value,
    page: messagePage.value + 1,
    wxid: selectedWxid.value,
  });
}

function onConversationScroll(event: Event) {
  const target = event.target as HTMLElement;
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 80) {
    void loadMoreConversations();
  }
}

function onMessageScroll(event: Event) {
  const target = event.target as HTMLElement;
  if (target.scrollTop <= 32) {
    void loadOlderMessages();
  }
}

function triggerImageSelect() {
  fileInputRef.value?.click();
}

async function onImagePicked(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file || !selectedWxid.value || !selectedConversationId.value) return;

  if (!file.type.startsWith('image/')) {
    message.warning('请选择图片文件');
    target.value = '';
    return;
  }

  const reader = new FileReader();
  reader.onload = async () => {
    const dataUrl = String(reader.result || '');
    if (!dataUrl) return;
    loading.sendImage = true;
    try {
      const sent = await sendConversationImageApi(
        selectedWxid.value,
        selectedConversationId.value,
        dataUrl,
      );
      upsertSentMessage(sent);
      message.success('图片发送成功');
    } catch (error) {
      console.error(error);
    } finally {
      loading.sendImage = false;
      target.value = '';
    }
  };
  reader.onerror = () => {
    message.error('读取图片失败');
    target.value = '';
  };
  reader.readAsDataURL(file);
}

async function sendText() {
  const content = draft.value.trim();
  if (!content || !selectedWxid.value || !selectedConversationId.value) return;
  loading.sendText = true;
  try {
    const sent = await sendConversationTextApi(
      selectedWxid.value,
      selectedConversationId.value,
      content,
    );
    upsertSentMessage(sent);
    resetMessageComposer();
    message.success('消息发送成功');
  } catch (error) {
    console.error(error);
  } finally {
    loading.sendText = false;
  }
}

async function openEmojiModal() {
  emojiModalOpen.value = true;
  if (!selectedWxid.value) return;
  loading.emojiCatalog = true;
  try {
    const items = await listRecentEmojisApi(selectedWxid.value, 48);
    emojiCatalog.value = items || [];
    if (!selectedEmojiKey.value && emojiCatalog.value.length) {
      const first = emojiCatalog.value[0];
      selectedEmojiKey.value = `${first.md5}:${first.totalLen}`;
    }
  } catch (error) {
    console.error(error);
  } finally {
    loading.emojiCatalog = false;
  }
}

function appendUnicodeEmoji(symbol: string) {
  draft.value = `${draft.value}${symbol}`;
}

async function sendEmoji() {
  if (!selectedWxid.value || !selectedConversationId.value) return;
  if (!selectedEmojiItem.value) {
    message.warning('请选择一个微信表情');
    return;
  }
  loading.sendEmoji = true;
  try {
    const sent = await sendConversationEmojiApi(selectedWxid.value, selectedConversationId.value, {
      md5: selectedEmojiItem.value.md5,
      totalLen: Number(selectedEmojiItem.value.totalLen) || 0,
    });
    upsertSentMessage(sent);
    emojiModalOpen.value = false;
    message.success('表情发送成功');
  } catch (error) {
    console.error(error);
  } finally {
    loading.sendEmoji = false;
  }
}

async function sendCard() {
  if (!selectedWxid.value || !selectedConversationId.value) return;
  if (!cardForm.cardWxid.trim()) {
    message.warning('请输入名片 wxid');
    return;
  }
  loading.sendCard = true;
  try {
    const sent = await shareConversationCardApi(selectedWxid.value, selectedConversationId.value, {
      cardAlias: cardForm.cardAlias.trim() || undefined,
      cardNickname: cardForm.cardNickname.trim() || undefined,
      cardWxid: cardForm.cardWxid.trim(),
    });
    upsertSentMessage(sent);
    cardModalOpen.value = false;
    cardForm.cardAlias = '';
    cardForm.cardNickname = '';
    cardForm.cardWxid = '';
    message.success('名片发送成功');
  } catch (error) {
    console.error(error);
  } finally {
    loading.sendCard = false;
  }
}

async function sendLink() {
  if (!selectedWxid.value || !selectedConversationId.value) return;
  if (!linkForm.url.trim()) {
    message.warning('请输入分享链接');
    return;
  }
  loading.sendLink = true;
  try {
    const sent = await shareConversationLinkApi(selectedWxid.value, selectedConversationId.value, {
      description: linkForm.description.trim() || undefined,
      thumbUrl: linkForm.thumbUrl.trim() || undefined,
      title: linkForm.title.trim() || undefined,
      url: linkForm.url.trim(),
    });
    upsertSentMessage(sent);
    linkModalOpen.value = false;
    linkForm.description = '';
    linkForm.thumbUrl = '';
    linkForm.title = '';
    linkForm.url = '';
    message.success('链接发送成功');
  } catch (error) {
    console.error(error);
  } finally {
    loading.sendLink = false;
  }
}

async function saveAISetting() {
  if (!selectedWxid.value || !selectedConversationId.value) return;
  aiSettingLoadNonce += 1;
  loading.aiSettingSave = true;
  try {
    const saved = await updateConversationAISettingApi(selectedWxid.value, selectedConversationId.value, {
      apiKey: aiSetting.apiKey.trim() || undefined,
      apiBaseUrl: aiSetting.apiBaseUrl.trim() || undefined,
      enabled: aiSetting.enabled,
      keywordTriggerEnabled: aiSetting.keywordTriggerEnabled,
      model: aiSetting.model || undefined,
      provider: aiSetting.provider,
      systemPrompt: aiSetting.systemPrompt.trim(),
      triggerKeywords: aiKeywordsInput.value
        .split(/\n|,|，|;|；/)
        .map((item) => item.trim())
        .filter(Boolean),
    });
    applyConversationAISetting(saved);
    aiSettingModalOpen.value = false;
    message.success('AI 设置已保存');
  } catch (error) {
    console.error(error);
  } finally {
    loading.aiSettingSave = false;
  }
}

async function generateAIDraft() {
  if (!selectedWxid.value || !selectedConversationId.value) return;
  loading.aiDraft = true;
  try {
    const result = await generateConversationAIDraftApi(
      selectedWxid.value,
      selectedConversationId.value,
      draft.value.trim(),
    );
    draft.value = result.content;
    message.success(`已生成 ${result.provider} 草稿`);
  } catch (error) {
    console.error(error);
  } finally {
    loading.aiDraft = false;
  }
}

async function removeConversation(conversationId: string) {
  const wxid = selectedWxid.value;
  if (!wxid || !conversationId) return;

  await deleteConversationApi(wxid, conversationId);

  const removedWasSelected = selectedConversationId.value === conversationId;
  conversationItems.value = conversationItems.value.filter(
    (item) => item.conversationId !== conversationId,
  );

  const unreadSnapshot = { ...localUnreadState.value };
  delete unreadSnapshot[conversationId];
  localUnreadState.value = unreadSnapshot;

  delete detailCache.value[getMessageCacheKey(wxid, conversationId)];
  delete messageCache.value[getMessageCacheKey(wxid, conversationId)];
  for (const key of Object.keys(conversationCache.value)) {
    if (!key.startsWith(`${wxid}::`)) continue;
    const entry = conversationCache.value[key];
    if (!entry) continue;
    conversationCache.value[key] = {
      ...entry,
      items: entry.items.filter((item) => item.conversationId !== conversationId),
      total: Math.max(0, entry.total - 1),
      totalPages: Math.ceil(Math.max(0, entry.total-1) / CONVERSATION_PAGE_SIZE),
    };
  }

  if (!removedWasSelected) {
    message.success('会话已删除');
    return;
  }

  conversationDetail.value = null;
  messageItems.value = [];
  selectedConversationId.value = '';

  const nextConversationId = conversationItems.value[0]?.conversationId || '';
  if (nextConversationId) {
    await selectConversation(nextConversationId);
  } else {
    await syncWorkspaceRoute(wxid, '');
  }
  message.success('会话已删除');
}

function confirmDeleteConversation(item: ConversationSummary) {
  Modal.confirm({
    cancelText: '取消',
    centered: true,
    content: `将从当前账号的本地消息中心移除会话“${item.targetName || item.targetWxid}”。`,
    okButtonProps: {
      danger: true,
    },
    okText: '删除',
    onOk: async () => {
      await removeConversation(item.conversationId);
    },
    title: '删除会话',
  });
}

function onComposerKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.shiftKey) return;
  event.preventDefault();
  if (loading.sendText) return;
  void sendText();
}

watch(selectedWxid, async (value, previous) => {
  emojiCatalog.value = [];
  selectedEmojiKey.value = '';
  if (value === previous && value) return;
  if (!value) {
    await activateAccountWorkspace('');
    return;
  }
  await activateAccountWorkspace(value, {
    preferredConversationId:
      typeof route.query.conversation === 'string' ? route.query.conversation : '',
  });
});

watch(
  () => aiSetting.provider,
  (value) => {
    const provider = aiProviders.value.find((item) => item.key === value);
    if (!provider) return;
    if (!provider.models.includes(aiSetting.model)) {
      aiSetting.model = provider.defaultModel;
    }
    if (!provider.supportsCustomBaseUrl) {
      aiSetting.apiBaseUrl = '';
    }
  },
);

watch(
  () => route.query.wxid,
  async (value) => {
    const nextWxid = typeof value === 'string' ? value : '';
    if (!nextWxid || nextWxid === selectedWxid.value) return;
    await loadAccounts();
    if (!accounts.value.some((item) => item.wxid === nextWxid)) return;
    selectedWxid.value = nextWxid;
  },
);

watch(
  () => route.query.conversation,
  async (value) => {
    const nextId = typeof value === 'string' ? value : '';
    if (!nextId || !selectedWxid.value || nextId === selectedConversationId.value) return;
    await selectConversation(nextId, false);
  },
);

onMounted(async () => {
  const routeWxid = typeof route.query.wxid === 'string' ? route.query.wxid : '';
  const routeConversation =
    typeof route.query.conversation === 'string' ? route.query.conversation : '';
  await loadAccounts();
  const targetWxid =
    routeWxid && accounts.value.some((item) => item.wxid === routeWxid)
      ? routeWxid
      : selectedWxid.value;
  if (targetWxid && targetWxid !== selectedWxid.value) {
    selectedWxid.value = targetWxid;
    return;
  }
  if (targetWxid) {
    await activateAccountWorkspace(targetWxid, {
      preferredConversationId: routeConversation,
      quietSync: true,
    });
  }
});

onBeforeUnmount(() => {
  closeRealtimeSocket();
});
</script>

<template>
  <Page>
    <div class="mb-4 flex flex-wrap items-center gap-3">
      <Typography.Text type="secondary">当前账号</Typography.Text>
      <Select
        v-model:value="selectedWxid"
        :loading="loading.accounts"
        :options="accountOptions"
        class="!w-[280px]"
        placeholder="请选择账号"
      />
      <Tag v-if="currentAccount" :color="currentAccount.status === 'online' ? 'success' : 'error'">
        {{ currentAccount.status === 'online' ? '在线' : '离线' }}
      </Tag>
      <Tag v-if="loading.syncing" color="processing">同步中</Tag>
      <Tag :color="realtimeStatusColor">
        {{ realtimeStatusLabel }}
      </Tag>
      <Typography.Text
        v-if="realtime.lastEventPreview"
        class="max-w-[240px] truncate !text-xs !text-slate-500"
      >
        {{ realtime.lastEventPreview }}
      </Typography.Text>
    </div>

    <div class="h-[calc(100vh-220px)] min-h-[640px] overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div class="flex h-full">
        <aside class="flex h-full w-[340px] min-w-[320px] flex-col border-r border-slate-200 bg-slate-50/50 p-4">
          <Input.Search
            v-model:value="searchInput"
            allow-clear
            placeholder="搜索会话 / 群聊 / 公众号"
            @search="searchConversations"
          />
          <div
            class="mt-3 min-h-0 flex-1 overflow-auto pr-1"
            ref="conversationPaneRef"
            @scroll="onConversationScroll"
          >
            <Spin :spinning="loading.conversations || loading.syncing">
              <Space class="w-full" direction="vertical" size="small">
                <Dropdown
                  v-for="item in conversationItems"
                  :key="item.conversationId"
                  :trigger="['contextmenu']"
                >
                  <button
                    :class="[
                      'w-full rounded-xl border p-3 text-left transition',
                      selectedConversationId === item.conversationId
                        ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))]'
                        : 'border-transparent bg-white hover:border-slate-200',
                    ]"
                    type="button"
                    @click="selectConversation(item.conversationId)"
                  >
                    <div class="flex items-start gap-3">
                      <Badge :count="item.unreadCount || 0" :offset="[-3, 4]">
                        <Avatar :src="item.targetAvatar || undefined">
                          {{ (item.targetName || item.targetWxid).slice(0, 1) }}
                        </Avatar>
                      </Badge>
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2">
                          <Typography.Text class="truncate" strong>
                            {{ item.targetName || item.targetWxid }}
                          </Typography.Text>
                          <Tag :bordered="false" :color="item.conversationType === 'group' ? 'purple' : 'blue'">
                            {{ item.conversationType === 'group' ? '群聊' : '单聊' }}
                          </Tag>
                          <Badge v-if="item.unreadCount > 0" dot color="#ff4d4f" />
                        </div>
                        <div class="mt-1 flex items-center justify-between gap-2">
                          <Typography.Text
                            :class="item.unreadCount > 0 ? '!font-medium !text-slate-900' : ''"
                            class="truncate !text-xs"
                            type="secondary"
                          >
                            {{ parseWechatEmojiText(item.lastMessage || '暂无消息') }}
                          </Typography.Text>
                          <Typography.Text class="shrink-0 !text-xs" type="secondary">
                            {{ formatConversationTime(item.lastMessageTime) }}
                          </Typography.Text>
                        </div>
                      </div>
                    </div>
                  </button>
                  <template #overlay>
                    <Menu>
                      <Menu.Item key="delete" danger @click="confirmDeleteConversation(item)">
                        删除会话
                      </Menu.Item>
                    </Menu>
                  </template>
                </Dropdown>
              </Space>

              <div class="py-3 text-center">
                <Button
                  v-if="conversationHasMore"
                  :loading="loading.conversationsMore"
                  size="small"
                  @click="loadMoreConversations"
                >
                  加载更多会话
                </Button>
              </div>

              <Empty
                v-if="!conversationItems.length && !loading.conversations"
                description="暂无会话数据"
              />
            </Spin>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col bg-slate-50">
          <template v-if="selectedConversationId">
            <header class="flex items-center justify-between border-b border-slate-200 bg-white px-5 py-3">
              <Space>
                <Avatar :size="42" :src="conversationDetail?.targetAvatar || currentConversation?.targetAvatar || undefined">
                  {{ (conversationDetail?.targetName || currentConversation?.targetName || '').slice(0, 1) }}
                </Avatar>
                <div class="min-w-0">
                  <Typography.Title :level="5" class="!mb-0 truncate">
                    {{ conversationDetail?.targetName || currentConversation?.targetName || selectedConversationId }}
                  </Typography.Title>
                  <Typography.Text type="secondary">
                    {{ selectedConversationId }}
                  </Typography.Text>
                </div>
              </Space>
              <Space>
                <Tag :color="conversationDetail?.conversationType === 'group' ? 'purple' : 'blue'">
                  {{ conversationDetail?.conversationType === 'group' ? '群聊' : '联系人' }}
                </Tag>
                <Tag v-if="conversationDetail?.conversationType === 'group'" color="processing">
                  {{ conversationDetail.memberCount }} 人
                </Tag>
                <Button type="text" @click="detailDrawerOpen = true">会话详情</Button>
              </Space>
            </header>

            <div
              class="min-h-0 flex-1 overflow-auto px-5 py-4"
              ref="messagePaneRef"
              @scroll="onMessageScroll"
            >
              <div class="pb-3 text-center">
                <Button
                  v-if="messageHasMore"
                  :loading="loading.messagesMore"
                  size="small"
                  @click="loadOlderMessages"
                >
                  加载更早消息
                </Button>
              </div>

              <Spin :spinning="loading.messages || loading.conversationDetail">
                <Space class="w-full" direction="vertical" size="middle">
                  <div
                    v-for="item in messageItems"
                    :key="`${item.messageId}-${item.msgId}-${item.createdAt}`"
                    :class="['flex w-full gap-2', item.isSelf ? 'justify-end' : 'justify-start']"
                  >
                    <Avatar
                      v-if="!item.isSelf"
                      :src="item.senderAvatar || undefined"
                      class="!self-end"
                    >
                      {{ (item.senderName || item.senderWxid).slice(0, 1) }}
                    </Avatar>

                    <div
                      :class="[
                        'max-w-[76%] rounded-xl px-3 py-2 shadow-sm',
                        item.isSelf
                          ? 'bg-blue-600 text-white'
                          : 'border border-slate-200 bg-white text-slate-900',
                      ]"
                    >
                      <Typography.Text
                        v-if="!item.isSelf && item.chatType === 'group'"
                        class="!mb-1 !block !text-xs"
                        type="secondary"
                      >
                        {{ item.senderName || item.senderWxid }}
                      </Typography.Text>

                      <div v-if="item.messageType === 'text'" class="whitespace-pre-wrap break-words">
                        {{ parseWechatEmojiText(item.content) }}
                      </div>

                      <div v-else-if="item.messageType === 'image'" class="space-y-2">
                        <img
                          v-if="imageFromMeta(item)"
                          :src="imageFromMeta(item)"
                          alt="image"
                          class="max-h-[240px] max-w-full cursor-zoom-in rounded-lg object-cover transition hover:opacity-90"
                          @click="openImagePreview(item)"
                        />
                        <div v-else>[图片消息]</div>
                      </div>

                      <div v-else-if="item.messageType === 'emoji'" class="space-y-1">
                        <div>[表情消息]</div>
                        <Typography.Text
                          class="!block !text-xs"
                          :class="item.isSelf ? '!text-white/80' : '!text-slate-500'"
                        >
                          MD5: {{ emojiFromMeta(item)?.md5 || '-' }}
                        </Typography.Text>
                      </div>

                      <div v-else-if="item.messageType === 'card'" class="space-y-1">
                        <div class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-slate-900">
                          <div class="font-medium">
                            {{ cardFromMeta(item)?.nickname || cardFromMeta(item)?.alias || '名片' }}
                          </div>
                          <div class="text-xs text-slate-500">
                            {{ cardFromMeta(item)?.wxid || item.content }}
                          </div>
                        </div>
                      </div>

                      <div v-else-if="item.messageType === 'link'" class="space-y-2">
                        <a
                          v-if="linkFromMeta(item)?.url"
                          :href="linkFromMeta(item)?.url"
                          class="block rounded-lg border border-slate-200 bg-slate-50 p-3 text-slate-900 no-underline transition hover:border-slate-300"
                          rel="noopener noreferrer"
                          target="_blank"
                        >
                          <div class="font-medium">{{ linkFromMeta(item)?.title || '链接消息' }}</div>
                          <div class="mt-1 text-xs text-slate-500 line-clamp-2">
                            {{ linkFromMeta(item)?.summary || linkFromMeta(item)?.url }}
                          </div>
                        </a>
                        <div v-else>{{ item.content || '[链接消息]' }}</div>
                      </div>

                      <div v-else-if="item.messageType === 'video'">[视频消息]</div>
                      <div v-else-if="item.messageType === 'voice'">[语音消息]</div>
                      <div v-else-if="item.messageType === 'system'">[系统消息] {{ parseWechatEmojiText(item.content) }}</div>
                      <div v-else>{{ parseWechatEmojiText(item.content || `[${item.messageType}]`) }}</div>

                      <div
                        :class="[
                          'mt-1 block text-right text-[11px]',
                          item.isSelf ? 'text-white/80' : 'text-slate-400',
                        ]"
                      >
                        {{ formatClock(item.createdAt) }}
                      </div>
                    </div>

                    <Avatar
                      v-if="item.isSelf"
                      :src="currentAccount?.avatar || undefined"
                      class="!self-end"
                    >
                      {{ (currentAccount?.nickname || currentAccount?.alias || currentAccount?.wxid || '').slice(0, 1) }}
                    </Avatar>
                  </div>
                </Space>
              </Spin>

              <Empty v-if="!loading.messages && !messageItems.length" description="暂无消息记录" />
            </div>

            <footer class="border-t border-slate-200 bg-white px-5 py-4">
              <input
                ref="fileInputRef"
                accept="image/*"
                class="hidden"
                type="file"
                @change="onImagePicked"
              />
              <div class="rounded-2xl border border-slate-200 bg-slate-50/90 p-3 shadow-sm">
                <div class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200/80 pb-3">
                  <div class="flex flex-wrap items-center gap-2">
                  <Button
                    :loading="loading.sendImage"
                    class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                    size="small"
                    @click="triggerImageSelect"
                  >
                    图片
                  </Button>
                  <Button
                    class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                    size="small"
                    @click="openEmojiModal"
                  >
                    表情
                  </Button>
                  <Button
                    class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                    size="small"
                    @click="cardModalOpen = true"
                  >
                    名片
                  </Button>
                  <Button
                    class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                    size="small"
                    @click="linkModalOpen = true"
                  >
                    链接
                  </Button>
                  </div>

                  <div class="flex flex-wrap items-center gap-2">
                    <Tag :color="aiSetting.enabled ? 'success' : 'default'">
                      AI {{ aiSetting.enabled ? '已开启' : '未开启' }}
                    </Tag>
                    <Tag v-if="selectedAIProvider" color="processing">
                      {{ selectedAIProvider.label }}
                    </Tag>
                    <Button
                      :loading="loading.aiDraft"
                      class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                      size="small"
                      @click="generateAIDraft"
                    >
                      AI 生成草稿
                    </Button>
                    <Button
                      class="!rounded-xl !border-slate-200 !bg-white !px-3 !text-slate-700 hover:!border-[rgb(var(--primary-5))] hover:!text-[rgb(var(--primary-6))]"
                      size="small"
                      @click="aiSettingModalOpen = true"
                    >
                      AI 设置
                    </Button>
                  </div>
                </div>

                <div class="pt-3">
                  <Input.TextArea
                    v-model:value="draft"
                    :auto-size="{ maxRows: 6, minRows: 4 }"
                    class="composer-textarea"
                    placeholder="输入消息内容，Enter 发送，Shift+Enter 换行"
                    @keydown="onComposerKeydown"
                  />
                </div>

                <div class="mt-3 flex flex-wrap items-center justify-between gap-3">
                  <Typography.Text class="!text-xs !text-slate-400">
                    Enter 直接发送，Shift + Enter 换行
                  </Typography.Text>

                  <div class="flex items-center gap-3">
                    <Typography.Text class="!text-xs !text-slate-400">
                      {{ draft.trim().length }} 字
                    </Typography.Text>
                    <Button
                      :disabled="!draft.trim()"
                      :loading="loading.sendText"
                      class="!h-10 !rounded-xl !px-5 !font-medium"
                      type="primary"
                      @click="sendText"
                    >
                      发送
                    </Button>
                  </div>
                </div>
              </div>
            </footer>
          </template>

          <Empty v-else class="pt-20" description="请选择左侧会话开始处理消息" />
        </section>
      </div>
    </div>

    <Drawer
      :open="detailDrawerOpen"
      :title="conversationDetail?.targetName || selectedConversationId"
      placement="right"
      width="420"
      @close="detailDrawerOpen = false"
    >
      <Descriptions :column="1" bordered size="small">
        <Descriptions.Item label="会话 ID">{{ selectedConversationId }}</Descriptions.Item>
        <Descriptions.Item label="会话类型">
          {{ conversationDetail?.conversationType === 'group' ? '群聊' : '单聊' }}
        </Descriptions.Item>
        <Descriptions.Item label="备注">{{ conversationDetail?.remark || '-' }}</Descriptions.Item>
        <Descriptions.Item label="群主">{{ conversationDetail?.chatRoomOwner || '-' }}</Descriptions.Item>
        <Descriptions.Item label="成员数">{{ conversationDetail?.memberCount || 0 }}</Descriptions.Item>
        <Descriptions.Item label="公告">{{ conversationDetail?.announcement || '-' }}</Descriptions.Item>
      </Descriptions>
      <Divider />
      <Typography.Title :level="5">群成员</Typography.Title>
      <div class="max-h-[420px] overflow-auto rounded-lg border border-slate-200 p-2">
        <Space class="w-full" direction="vertical" size="small">
          <div
            v-for="member in groupMembers"
            :key="member.userName"
            class="rounded-md bg-slate-50 px-3 py-2"
          >
            <div class="truncate text-sm font-medium text-slate-900">
              {{ member.nickName || member.userName }}
            </div>
            <div class="truncate text-xs text-slate-500">{{ member.userName }}</div>
          </div>
          <Empty v-if="!groupMembers.length" description="暂无成员详情" />
        </Space>
      </div>
    </Drawer>

    <Modal
      v-model:open="aiSettingModalOpen"
      :confirm-loading="loading.aiSettingSave"
      title="会话 AI 设置"
      width="70vw"
      @ok="saveAISetting"
    >
      <div class="space-y-3">
        <div class="rounded-xl border border-slate-200 bg-slate-50 px-3 py-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div class="text-sm font-medium text-slate-900">自动回复</div>
              <div class="text-xs text-slate-500">收到新消息后允许 AI 自动生成并发送回复</div>
            </div>
            <Switch v-model:checked="aiSetting.enabled" />
          </div>
          <div class="mt-3 flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 pt-3">
            <div>
              <div class="text-sm font-medium text-slate-900">关键词触发</div>
              <div class="text-xs text-slate-500">仅命中关键词时触发自动回复</div>
            </div>
            <Switch v-model:checked="aiSetting.keywordTriggerEnabled" :disabled="aiAdvancedDisabled" />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-3">
          <div>
            <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">AI 平台</div>
            <Select
              v-model:value="aiSetting.provider"
              :disabled="aiAdvancedDisabled"
              :loading="loading.aiProviders"
              :options="aiProviderOptions"
              placeholder="请选择 AI 平台"
              size="large"
            />
          </div>
          <div>
            <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">接口地址</div>
            <Input
              v-model:value="aiSetting.apiBaseUrl"
              :disabled="aiCustomBaseURLDisabled"
              placeholder="仅自定义中转站可填写，例如 https://your-relay.example.com/v1 或完整 /chat/completions、/messages 地址"
              size="large"
            />
            <div class="mt-1 text-xs text-slate-400">
              仅“自定义中转站”平台可配置。支持填写基础地址，系统会自动补全接口路径；也支持直接填写完整接口地址。
            </div>
          </div>
          <div>
            <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">模型</div>
            <Select
              v-model:value="aiSetting.model"
              :disabled="aiAdvancedDisabled"
              :options="aiModelOptions"
              placeholder="请选择模型"
              size="large"
            />
          </div>
          <div>
            <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">API Key</div>
            <Input.Password
              v-model:value="aiSetting.apiKey"
              :disabled="aiAdvancedDisabled"
              placeholder="会话专用 API Key，优先于系统默认配置"
              size="large"
            />
            <div class="mt-1 text-xs text-slate-400">留空时使用系统默认 API Key。</div>
          </div>
        </div>

        <div>
          <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">关键词</div>
          <Input.TextArea
            v-model:value="aiKeywordsInput"
            :disabled="aiAdvancedDisabled"
            :auto-size="{ minRows: 2, maxRows: 3 }"
            placeholder="一行一个，或使用逗号分隔。"
          />
        </div>

        <div>
          <div class="mb-1.5 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">系统提示词</div>
          <Input.TextArea
            v-model:value="aiSetting.systemPrompt"
            :disabled="aiAdvancedDisabled"
            :auto-size="{ minRows: 3, maxRows: 6 }"
            placeholder="例如：你是售后客服，请礼貌、简洁地回复用户。"
          />
        </div>
      </div>
    </Modal>

    <Modal
      v-model:open="emojiModalOpen"
      :confirm-loading="loading.sendEmoji"
      title="发送表情"
      width="520"
      @ok="sendEmoji"
    >
      <div class="space-y-4">
        <div>
          <div class="mb-2 text-xs font-medium uppercase tracking-[0.12em] text-slate-400">常用 Emoji</div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="symbol in COMMON_EMOJI_CHOICES"
              :key="symbol"
              class="emoji-chip"
              type="button"
              @click="appendUnicodeEmoji(symbol)"
            >
              {{ symbol }}
            </button>
          </div>
          <div class="mt-2 text-xs text-slate-400">点击可直接插入输入框，不会立即发送。</div>
        </div>

        <div>
          <div class="mb-2 flex items-center justify-between gap-3">
            <div class="text-xs font-medium uppercase tracking-[0.12em] text-slate-400">最近微信表情</div>
            <Typography.Text class="!text-xs !text-slate-400">
              选择后点击“确定”发送
            </Typography.Text>
          </div>
          <div v-if="loading.emojiCatalog" class="py-6 text-center">
            <Spin size="small" />
          </div>
          <div v-else-if="emojiCatalog.length" class="grid grid-cols-2 gap-2 sm:grid-cols-3">
            <button
              v-for="item in emojiCatalog"
              :key="`${item.md5}:${item.totalLen}`"
              :class="[
                'emoji-option-card',
                selectedEmojiKey === `${item.md5}:${item.totalLen}` ? 'emoji-option-card-active' : '',
              ]"
              type="button"
              @click="selectedEmojiKey = `${item.md5}:${item.totalLen}`"
            >
              <div class="text-sm font-medium text-slate-800">{{ item.label }}</div>
              <div class="mt-1 text-xs text-slate-500">
                {{ item.width || '-' }} x {{ item.height || '-' }} · {{ item.totalLen || 0 }}B
              </div>
              <div class="mt-2 text-[11px] text-slate-400">{{ item.md5.slice(0, 12) }}</div>
            </button>
          </div>
          <Empty v-else description="当前账号还没有可用的历史微信表情" />
        </div>
      </div>
    </Modal>

    <Modal
      v-model:open="cardModalOpen"
      :confirm-loading="loading.sendCard"
      title="分享名片"
      @ok="sendCard"
    >
      <Space class="w-full" direction="vertical">
        <Input v-model:value="cardForm.cardWxid" placeholder="名片 wxid（必填）" />
        <Input v-model:value="cardForm.cardNickname" placeholder="名片昵称（可选）" />
        <Input v-model:value="cardForm.cardAlias" placeholder="名片别名（可选）" />
      </Space>
    </Modal>

    <Modal
      v-model:open="linkModalOpen"
      :confirm-loading="loading.sendLink"
      title="分享链接"
      @ok="sendLink"
    >
      <Space class="w-full" direction="vertical">
        <Input v-model:value="linkForm.url" placeholder="链接 URL（必填）" />
        <Input v-model:value="linkForm.title" placeholder="标题（可选）" />
        <Input.TextArea
          v-model:value="linkForm.description"
          :auto-size="{ maxRows: 4, minRows: 2 }"
          placeholder="描述（可选）"
        />
        <Input v-model:value="linkForm.thumbUrl" placeholder="封面图 URL（可选）" />
      </Space>
    </Modal>

    <Modal
      v-model:open="imagePreviewOpen"
      :footer="null"
      centered
      width="auto"
      @cancel="imagePreviewLoading = false"
    >
      <Spin :spinning="imagePreviewLoading">
        <img
          v-if="imagePreviewSrc"
          :src="imagePreviewSrc"
          alt="preview"
          class="max-h-[78vh] max-w-[80vw] rounded-lg object-contain"
        />
      </Spin>
    </Modal>
  </Page>
</template>

<style scoped>
.emoji-chip {
  align-items: center;
  background: #fff;
  border: 1px solid rgb(226 232 240);
  border-radius: 12px;
  display: inline-flex;
  font-size: 20px;
  height: 42px;
  justify-content: center;
  transition: all 0.2s ease;
  width: 42px;
}

.emoji-chip:hover {
  border-color: rgb(var(--primary-5));
  transform: translateY(-1px);
}

.emoji-option-card {
  background: #fff;
  border: 1px solid rgb(226 232 240);
  border-radius: 14px;
  padding: 12px;
  text-align: left;
  transition: all 0.2s ease;
}

.emoji-option-card:hover,
.emoji-option-card-active {
  border-color: rgb(var(--primary-5));
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.12);
}

:deep(.composer-textarea textarea.ant-input) {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: inset 0 1px 2px rgba(15, 23, 42, 0.04);
  font-size: 15px;
  line-height: 1.7;
  padding: 14px 16px;
}

:deep(.composer-textarea textarea.ant-input::placeholder) {
  color: rgb(148 163 184);
}

:deep(.composer-textarea.ant-input-textarea-show-count::after) {
  display: none;
}
</style>
