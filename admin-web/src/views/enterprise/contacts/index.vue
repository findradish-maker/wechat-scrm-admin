<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Checkbox,
  Dropdown,
  Empty,
  Input,
  Menu,
  Modal,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import {
  checkFriendRelationBatchApi,
  checkFriendRelationApi,
  deleteFriendApi,
  listContactsApi,
  reloadContactsApi,
  searchFriendCandidatesApi,
  sendFriendRequestBatchApi,
  sendFriendRequestApi,
  setFriendBlacklistApi,
  type ContactSummary,
  type FriendRelationResult,
  type FriendSearchTarget,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

type ContactType = 'contact' | 'group' | 'official_account';
type ContactListSection = {
  category: string;
  categoryLabel: string;
  groups: Array<{
    letter: string;
    items: ContactSummary[];
  }>;
};

const accountStore = useEnterpriseAccountStore();
const router = useRouter();
const CONTACT_PAGE_SIZE = 240;
const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const selectedCategory = ref<
  'contact' | 'group' | 'public_account'
>('contact');
const keyword = ref('');
const selectedId = ref('');
const listPaneRef = ref<HTMLElement | null>(null);
const contactPage = ref(1);
const contactTotalPages = ref(1);
const contactHasMore = ref(false);
const latestContactRequestId = ref(0);

const searchKeyword = ref('');
const verifyContent = ref('你好，我是企业微信运营同学');
const selectedContactWxids = ref<string[]>([]);
const selectedSearchKeys = ref<string[]>([]);
const addContactModalOpen = ref(false);

const accounts = computed(() => accountStore.accounts);
const contacts = ref<ContactSummary[]>([]);
const searchResults = ref<FriendSearchTarget[]>([]);

const relationMap = ref<Record<string, FriendRelationResult>>({});
const relationLoadingMap = ref<Record<string, boolean>>({});
const addLoadingMap = ref<Record<string, boolean>>({});

const loading = reactive({
  accounts: false,
  batchAdd: false,
  batchRelation: false,
  contacts: false,
  contactsMore: false,
  deleteFriend: false,
  search: false,
});

const accountOptions = computed(() => accountStore.accountOptions);

const currentAccount = computed(() => accountStore.currentAccount);

const filteredContacts = computed(() => {
  const key = keyword.value.trim().toLowerCase();
  return contacts.value
    .filter((item) => item.contactCategory === selectedCategory.value)
    .filter((item) => {
      if (!key) return true;
      return [item.displayName, item.nickname, item.remark, item.wxid]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(key));
    });
});

const categorizedSections = computed<ContactListSection[]>(() => {
  const source = filteredContacts.value;
  if (!source.length) return [];

  const categoryLabelMap = new Map(categoryOptions.map((item) => [item.value, item.label]));
  const sectionMap = new Map<string, Map<string, ContactSummary[]>>();

  for (const item of source) {
    const category = item.contactCategory || 'contact';
    const letter = normalizeSortLetter(item.sortLetter);
    const categoryGroup = sectionMap.get(category) ?? new Map<string, ContactSummary[]>();
    const letterItems = categoryGroup.get(letter) ?? [];
    letterItems.push(item);
    categoryGroup.set(letter, letterItems);
    sectionMap.set(category, categoryGroup);
  }

  const orderedCategories = [selectedCategory.value];

  return orderedCategories
    .map((category) => {
      const groups = sectionMap.get(category);
      if (!groups || groups.size === 0) return null;
      return {
        category,
        categoryLabel: categoryLabelMap.get(category) || category,
        groups: [...groups.entries()]
          .sort(([left], [right]) => sortLetterOrder(left, right))
          .map(([letter, items]) => ({
            letter,
            items: [...items].sort(compareContacts),
          })),
      };
    })
    .filter(Boolean) as ContactListSection[];
});

const anchorLetters = computed(() => {
  const section = categorizedSections.value[0];
  if (!section) return [];
  return section.groups.map((group) => group.letter);
});

const currentContact = computed(() =>
  filteredContacts.value.find((item) => item.wxid === selectedId.value),
);

const currentRelation = computed(() => {
  if (!currentContact.value) return null;
  return relationMap.value[currentContact.value.wxid] ?? null;
});

const contactMetaItems = computed(() => {
  if (!currentContact.value) return [];
  const contact = currentContact.value;
  const location = `${contact.province || ''}${contact.city || ''}` || '未设置';
  return [
    { label: '昵称', value: contact.nickname || '未设置' },
    { label: '备注', value: contact.remark || '未设置' },
    { label: '地区', value: location },
    { label: '远端昵称', value: currentRelation.value?.nickName || '未同步' },
    {
      label: '关系状态',
      value: relationMeta(contact.wxid, contact.contactType).label,
    },
  ];
});

const relationSelectableContacts = computed(() =>
  filteredContacts.value.filter((item) => isRelationSupported(item)),
);

const selectedRelationTargets = computed(() =>
  relationSelectableContacts.value
    .filter((item) => selectedContactWxids.value.includes(item.wxid))
    .map((item) => item.wxid),
);

const selectedAddTargets = computed(() =>
  searchResults.value.filter(
    (item) => selectedSearchKeys.value.includes(getSearchResultKey(item)) && canAddFriend(item),
  ),
);

const addableSearchResults = computed(() =>
  searchResults.value.filter((item) => canAddFriend(item)),
);

const searchTerms = computed(() => parseSearchKeywords(searchKeyword.value));

const allAddSelected = computed(
  () => addableSearchResults.value.length > 0 && selectedAddTargets.value.length === addableSearchResults.value.length,
);

const addIndeterminate = computed(
  () => selectedAddTargets.value.length > 0 && selectedAddTargets.value.length < addableSearchResults.value.length,
);

const categoryOptions = [
  { label: '联系人', value: 'contact' },
  { label: '群聊', value: 'group' },
  { label: '公众号', value: 'public_account' },
] as const;

watch(
  filteredContacts,
  (items) => {
    const available = new Set(items.map((item) => item.wxid));
    selectedContactWxids.value = selectedContactWxids.value.filter((wxid) => available.has(wxid));
    if (!items.length) {
      selectedId.value = '';
      return;
    }
    if (!items.some((item) => item.wxid === selectedId.value)) {
      selectedId.value = items[0]?.wxid ?? '';
    }
  },
  { immediate: true },
);

watch(searchResults, (items) => {
  const available = new Set(items.map((item) => getSearchResultKey(item)));
  selectedSearchKeys.value = selectedSearchKeys.value.filter((key) => available.has(key));
});

watch(selectedWxid, async (next) => {
  relationMap.value = {};
  searchResults.value = [];
  selectedId.value = '';
  selectedContactWxids.value = [];
  if (!next) {
    contacts.value = [];
    contactPage.value = 1;
    contactTotalPages.value = 1;
    contactHasMore.value = false;
    return;
  }
  await loadContacts({ reset: true });
});

function mapType(type: ContactType) {
  if (type === 'group') return '群聊';
  if (type === 'official_account') return '公众号';
  return '联系人';
}

function typeTagColor(type: ContactType) {
  if (type === 'group') return 'purple';
  if (type === 'official_account') return 'blue';
  return 'success';
}

function relationMeta(wxid: string, contactType: ContactType) {
  if (contactType !== 'contact') {
    return { color: 'default', label: '非好友对象' };
  }
  const relation = relationMap.value[wxid];
  if (!relation) {
    return { color: 'default', label: '未检测' };
  }
  if (relation.relationKey === 'friend') {
    return { color: 'success', label: relation.relationLabel };
  }
  if (relation.relationKey === 'deleted') {
    return { color: 'default', label: relation.relationLabel };
  }
  if (relation.relationKey === 'blocked_by_self') {
    return { color: 'error', label: relation.relationLabel };
  }
  if (relation.relationKey === 'blocked_by_target') {
    return { color: 'warning', label: relation.relationLabel };
  }
  if (relation.relationKey === 'check_failed') {
    return { color: 'warning', label: relation.relationLabel };
  }
  return { color: 'processing', label: relation.relationLabel };
}

function isRelationSupported(item: ContactSummary) {
  return item.contactType === 'contact';
}

function canOpenConversation(item: ContactSummary) {
  return item.contactType === 'contact' || item.contactType === 'group';
}

function canAddFriend(item: FriendSearchTarget) {
  return !!(item.canAdd && item.v1 && item.v2);
}

function getSearchResultKey(item: FriendSearchTarget) {
  return `${item.source}:${item.wxid}`;
}

function parseSearchKeywords(value: string) {
  return [...new Set(
    value
      .split(/[\n,，;；\s]+/)
      .map((item) => item.trim())
      .filter(Boolean),
  )];
}

function mergeSearchResults(results: FriendSearchTarget[]) {
  const merged = new Map<string, FriendSearchTarget>();

  for (const item of results) {
    const existing = merged.get(item.wxid);
    if (!existing) {
      merged.set(item.wxid, item);
      continue;
    }

    merged.set(item.wxid, {
      ...existing,
      ...item,
      source: existing.source === 'primary' || item.source !== 'primary' ? existing.source : item.source,
      canAdd: existing.canAdd || item.canAdd,
      v1: existing.v1 || item.v1,
      v2: existing.v2 || item.v2,
      displayName: existing.displayName || item.displayName,
      avatar: existing.avatar || item.avatar,
    });
  }

  return [...merged.values()];
}

function toggleContactSelection(wxid: string, checked: boolean) {
  if (checked) {
    if (!selectedContactWxids.value.includes(wxid)) {
      selectedContactWxids.value = [...selectedContactWxids.value, wxid];
    }
    return;
  }
  selectedContactWxids.value = selectedContactWxids.value.filter((item) => item !== wxid);
}

function toggleSearchSelection(key: string, checked: boolean) {
  if (checked) {
    if (!selectedSearchKeys.value.includes(key)) {
      selectedSearchKeys.value = [...selectedSearchKeys.value, key];
    }
    return;
  }
  selectedSearchKeys.value = selectedSearchKeys.value.filter((item) => item !== key);
}

function toggleSelectAllSearchResults(checked: boolean) {
  selectedSearchKeys.value = checked ? addableSearchResults.value.map((item) => getSearchResultKey(item)) : [];
}

function showCategoryTag(item: ContactSummary) {
  return mapType(item.contactType) !== item.contactCategoryLabel;
}

function contactGroupAnchorId(category: string, letter: string) {
  return `contact-anchor-${category}-${letter === '#' ? 'misc' : letter.toLowerCase()}`;
}

function jumpToLetter(letter: string) {
  const section = categorizedSections.value[0];
  const container = listPaneRef.value;
  if (!section || !container) return;
  const target = document.getElementById(contactGroupAnchorId(section.category, letter));
  if (!target) return;
  const offset = target.offsetTop - container.offsetTop - 8;
  container.scrollTo({ top: Math.max(offset, 0), behavior: 'smooth' });
}

function openMessageConversation(item: ContactSummary) {
  if (!selectedWxid.value) {
    message.warning('请先选择账号');
    return;
  }
  void router.push({
    name: 'EnterpriseMessages',
    query: {
      conversation: item.wxid,
      wxid: selectedWxid.value,
    },
  });
}

function confirmReloadContacts() {
  if (!selectedWxid.value) {
    message.warning('请先选择账号');
    return;
  }
  Modal.confirm({
    title: '重新加载当前账号联系人？',
    content: '将从 wechatReal 重新请求当前账号联系人，并覆盖本地缓存数据。',
    okText: '重新加载',
    cancelText: '取消',
    onOk: async () => {
      await reloadContactsApi(selectedWxid.value!);
      message.success('已开始后台重新加载联系人');
      window.setTimeout(() => {
        void loadContacts({ reset: true });
      }, 1500);
      window.setTimeout(() => {
        void loadContacts({ reset: true });
      }, 5000);
    },
  });
}

function openAddContactModal() {
  if (!selectedWxid.value) {
    message.warning('请先选择账号');
    return;
  }
  addContactModalOpen.value = true;
}

async function loadAccounts() {
  loading.accounts = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loading.accounts = false;
  }
}

function resetContactsWorkspace() {
  contacts.value = [];
  contactPage.value = 1;
  contactTotalPages.value = 1;
  contactHasMore.value = false;
  selectedId.value = '';
  selectedContactWxids.value = [];
}

function isCurrentContactRequest(requestId: number, wxid: string) {
  return latestContactRequestId.value === requestId && selectedWxid.value === wxid;
}

async function loadContacts(options: { forceRefresh?: boolean; reset?: boolean } = {}) {
  if (!selectedWxid.value) return;
  const { forceRefresh = false, reset = false } = options;
  if (reset) {
    if (loading.contacts) return;
    loading.contacts = true;
  } else {
    if (loading.contacts || loading.contactsMore || !contactHasMore.value) return;
    loading.contactsMore = true;
  }
  const requestWxid = selectedWxid.value;
  const requestId = latestContactRequestId.value + 1;
  latestContactRequestId.value = requestId;
  const targetPage = reset ? 1 : contactPage.value + 1;
  try {
    const data = await listContactsApi(requestWxid, {
      page: targetPage,
      pageSize: CONTACT_PAGE_SIZE,
      refresh: forceRefresh && reset,
    });
    if (!isCurrentContactRequest(requestId, requestWxid)) return;

    const merged = new Map<string, ContactSummary>();
    if (!reset) {
      for (const item of contacts.value) {
        merged.set(item.wxid, item);
      }
    }
    for (const item of data.contacts) {
      merged.set(item.wxid, item);
    }
    contacts.value = [...merged.values()];
    contactPage.value = data.page || targetPage;
    contactTotalPages.value = Math.max(data.totalPages || 1, 1);
    contactHasMore.value = contactPage.value < contactTotalPages.value;
  } catch (error) {
    if (!isCurrentContactRequest(requestId, requestWxid)) return;
    if (reset) {
      resetContactsWorkspace();
    }
    message.error((error as Error)?.message || '加载联系人失败');
  } finally {
    if (!isCurrentContactRequest(requestId, requestWxid)) return;
    loading.contacts = false;
    loading.contactsMore = false;
  }
}

function handleContactListScroll(event: Event) {
  const element = event.target as HTMLElement | null;
  if (!element || loading.contacts || loading.contactsMore || !contactHasMore.value) return;
  if (element.scrollTop + element.clientHeight >= element.scrollHeight - 160) {
    void loadContacts();
  }
}

async function checkRelation(targetWxid: string) {
  if (!selectedWxid.value || !targetWxid) return;
  relationLoadingMap.value[targetWxid] = true;
  try {
    const result = await checkFriendRelationApi(selectedWxid.value, targetWxid);
    relationMap.value = {
      ...relationMap.value,
      [targetWxid]: result,
    };
    message.success(`好友状态已更新：${result.relationLabel}`);
  } catch (error) {
    console.error(error);
  } finally {
    relationLoadingMap.value[targetWxid] = false;
  }
}

async function searchFriendCandidates() {
  const targets = searchTerms.value;
  if (!selectedWxid.value) {
    message.warning('请先选择账号');
    return;
  }
  if (!targets.length) {
    message.warning('请输入搜索关键词后再搜索');
    return;
  }
  loading.search = true;
  try {
    const settled = await Promise.allSettled(
      targets.map((target) => searchFriendCandidatesApi(selectedWxid.value, target)),
    );
    const mergedResults = mergeSearchResults(
      settled.flatMap((item) => (item.status === 'fulfilled' ? item.value.results : [])),
    );
    searchResults.value = mergedResults;
    if (!mergedResults.length) {
      message.info('未搜索到匹配联系人');
      return;
    }
    if (targets.length > 1) {
      message.success(`批量搜索完成：关键词 ${targets.length} 个，命中 ${mergedResults.length} 条`);
    }
  } finally {
    loading.search = false;
  }
}

async function sendFriendRequest(target: FriendSearchTarget) {
  if (!selectedWxid.value) return;
  if (!target.canAdd || !target.v1 || !target.v2) {
    message.warning('该搜索结果缺少校验参数，暂不可直接添加');
    return;
  }
  addLoadingMap.value[target.wxid] = true;
  try {
    await sendFriendRequestApi(selectedWxid.value, {
      targetWxid: target.wxid,
      v1: target.v1,
      v2: target.v2,
      verifyContent: verifyContent.value.trim() || undefined,
    });
    message.success(`已发送好友申请：${target.displayName || target.wxid}`);
  } catch (error) {
    console.error(error);
  } finally {
    addLoadingMap.value[target.wxid] = false;
  }
}

async function sendFriendRequestBatch() {
  if (!selectedWxid.value) return;
  const targets = selectedAddTargets.value;
  if (!targets.length) {
    message.warning('请先勾选需要批量添加的搜索结果');
    return;
  }

  loading.batchAdd = true;
  try {
    const result = await sendFriendRequestBatchApi(selectedWxid.value, {
      requests: targets.map((item) => ({
        targetWxid: item.wxid,
        v1: item.v1,
        v2: item.v2,
        verifyContent: verifyContent.value.trim() || undefined,
      })),
    });
    if (result.failed.length) {
      message.warning(`批量添加完成：成功 ${result.results.length}，失败 ${result.failed.length}`);
      return;
    }
    message.success(`批量添加完成：共 ${result.results.length} 条`);
  } finally {
    loading.batchAdd = false;
  }
}

async function checkRelationBatch() {
  if (!selectedWxid.value) return;
  const targets = selectedRelationTargets.value;
  if (!targets.length) {
    message.warning('请先勾选需要检测的联系人');
    return;
  }

  loading.batchRelation = true;
  try {
    const result = await checkFriendRelationBatchApi(selectedWxid.value, targets);
    const nextMap = { ...relationMap.value };
    for (const item of result.results) {
      nextMap[item.targetWxid] = item;
    }
    relationMap.value = nextMap;
    if (result.failed.length) {
      message.warning(`批量检测完成：成功 ${result.results.length}，失败 ${result.failed.length}`);
      return;
    }
    message.success(`批量检测完成：共 ${result.results.length} 条`);
  } finally {
    loading.batchRelation = false;
  }
}

function confirmDeleteCurrent() {
  if (!selectedWxid.value || !currentContact.value) return;
  if (!isRelationSupported(currentContact.value)) {
    message.warning('仅联系人支持删除好友');
    return;
  }
  Modal.confirm({
    title: '确认删除好友？',
    content: `将删除好友 ${currentContact.value.displayName || currentContact.value.wxid}`,
    okButtonProps: { danger: true, loading: loading.deleteFriend },
    onOk: async () => {
      loading.deleteFriend = true;
      try {
        await deleteFriendApi(selectedWxid.value, currentContact.value!.wxid);
        relationMap.value = {
          ...relationMap.value,
          [currentContact.value!.wxid]: {
            targetWxid: currentContact.value!.wxid,
            relationCode: 1,
            relationKey: 'deleted',
            relationLabel: '已删除',
            nickName: '',
            headImgUrl: '',
            sign: '',
          },
        };
        message.success('删除好友请求已提交');
      } finally {
        loading.deleteFriend = false;
      }
    },
  });
}

async function setCurrentBlacklist(action: 'add' | 'remove') {
  if (!selectedWxid.value || !currentContact.value) return;
  if (!isRelationSupported(currentContact.value)) {
    message.warning('仅联系人支持黑名单操作');
    return;
  }
  try {
    await setFriendBlacklistApi(selectedWxid.value, currentContact.value.wxid, action);
    if (action === 'add') {
      relationMap.value = {
        ...relationMap.value,
        [currentContact.value.wxid]: {
          targetWxid: currentContact.value.wxid,
          relationCode: 4,
          relationKey: 'blocked_by_self',
          relationLabel: '已拉黑对方',
          nickName: '',
          headImgUrl: '',
          sign: '',
        },
      };
      message.success('已加入黑名单');
      return;
    }
    await checkRelation(currentContact.value.wxid);
  } catch (error) {
    console.error(error);
  }
}

onMounted(async () => {
  await loadAccounts();
  if (selectedWxid.value) {
    await loadContacts({ reset: true });
  }
});

function normalizeSortLetter(value?: string) {
  const text = String(value || '').trim().toUpperCase();
  if (!text) return '#';
  const char = text[0]!;
  return char >= 'A' && char <= 'Z' ? char : '#';
}

function sortLetterOrder(left: string, right: string) {
  if (left === right) return 0;
  if (left === '#') return 1;
  if (right === '#') return -1;
  return left.localeCompare(right);
}

function compareContacts(left: ContactSummary, right: ContactSummary) {
  const leftKey = String(left.sortKey || left.displayName || left.wxid || '').toUpperCase();
  const rightKey = String(right.sortKey || right.displayName || right.wxid || '').toUpperCase();
  const byPinyin = leftKey.localeCompare(rightKey, 'zh-Hans-CN-u-co-pinyin');
  if (byPinyin !== 0) return byPinyin;
  return String(left.wxid || '').localeCompare(String(right.wxid || ''));
}
</script>

<template>
  <Page>
    <div class="mb-4 flex flex-wrap items-center gap-3">
      <Typography.Text type="secondary">当前账号</Typography.Text>
      <Select
        v-model:value="selectedWxid"
        :loading="loading.accounts"
        :options="accountOptions"
        class="!w-[320px]"
        placeholder="请选择账号"
      />
      <Tag v-if="currentAccount" :bordered="false" :color="currentAccount.status === 'online' ? 'success' : 'error'">
        {{ currentAccount.status === 'online' ? '在线' : '离线' }}
      </Tag>
      <Button
        type="primary"
        size="small"
        class="!h-8 !rounded-lg !px-3"
        @click="openAddContactModal"
      >
        添加好友
      </Button>
    </div>

    <div class="h-[calc(100vh-176px)] min-h-[620px] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_18px_48px_rgba(15,23,42,0.06)]">
      <div class="flex h-full">
        <aside class="flex h-full w-[410px] min-w-[390px] flex-col border-r border-slate-200 bg-[#f8fafc] p-4">
          <div class="space-y-2.5">
            <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_8px_20px_rgba(15,23,42,0.04)]">
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <div class="text-[26px] font-semibold tracking-tight text-slate-900">联系人</div>
                  <div class="mt-1 text-xs text-slate-500">当前账号通讯录</div>
                </div>
                <Tag :bordered="false" color="blue" class="!mr-0">
                  {{ filteredContacts.length }}
                </Tag>
              </div>
              <div class="mt-3 border-t border-slate-100 pt-3">
                <Input.Search
                  v-model:value="keyword"
                  allow-clear
                  placeholder="筛选姓名 / 备注 / wxid"
                />
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    v-for="option in categoryOptions"
                    :key="option.value"
                    type="button"
                    :class="[
                      'rounded-full border px-3 py-1.5 text-xs font-medium transition',
                      selectedCategory === option.value
                        ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))] text-[rgb(var(--primary-6))]'
                        : 'border-slate-200 bg-slate-50 text-slate-500 hover:border-slate-300 hover:bg-white',
                    ]"
                    @click="selectedCategory = option.value"
                  >
                    {{ option.label }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="relative mt-3 min-h-0 flex-1 pl-10">
            <div
              v-if="anchorLetters.length"
              class="pointer-events-none absolute left-0 top-3 z-20 flex flex-col items-center gap-0.5 rounded-[20px] border border-slate-200 bg-white/92 px-1 py-1.5 shadow-[0_12px_24px_rgba(15,23,42,0.08)] backdrop-blur"
            >
              <button
                v-for="letter in anchorLetters"
                :key="letter"
                type="button"
                class="pointer-events-auto flex h-4 w-4 items-center justify-center rounded-full text-[9px] font-semibold leading-none text-slate-400 transition hover:bg-[rgb(var(--primary-1))] hover:text-[rgb(var(--primary-6))]"
                @click="jumpToLetter(letter)"
              >
                {{ letter }}
              </button>
            </div>

            <div ref="listPaneRef" class="h-full overflow-auto" @scroll.passive="handleContactListScroll">
              <Spin :spinning="loading.contacts">
              <div class="space-y-4">
                <section
                  v-for="section in categorizedSections"
                  :key="section.category"
                  class="space-y-2.5"
                >
                  <div
                    v-for="group in section.groups"
                    :key="`${section.category}-${group.letter}`"
                    :id="contactGroupAnchorId(section.category, group.letter)"
                    class="space-y-1.5"
                  >
                    <div class="flex items-center gap-2 px-1 pt-1">
                      <div class="flex h-6 w-6 items-center justify-center rounded-full bg-slate-200 text-[11px] font-semibold text-slate-700">
                        {{ group.letter }}
                      </div>
                      <div class="text-[11px] font-medium uppercase tracking-[0.14em] text-slate-400">
                        {{ group.letter === '#' ? '其他' : group.letter }}
                      </div>
                    </div>

                    <button
                      v-for="item in group.items"
                      :key="item.wxid"
                      :class="[
                        'w-full rounded-xl border px-3.5 py-3 text-left transition shadow-[0_6px_20px_rgba(15,23,42,0.035)]',
                        selectedId === item.wxid
                          ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))] shadow-[0_10px_24px_rgba(24,144,255,0.1)]'
                          : 'border-transparent bg-white hover:border-slate-200 hover:bg-slate-50',
                      ]"
                      type="button"
                      @click="selectedId = item.wxid"
                    >
                      <div class="flex items-center gap-3">
                        <div v-if="isRelationSupported(item)" class="shrink-0" @click.stop>
                          <Checkbox
                            :checked="selectedContactWxids.includes(item.wxid)"
                            @change="(event) => toggleContactSelection(item.wxid, event.target.checked)"
                          />
                        </div>
                        <Avatar :size="40" :src="item.avatar || undefined">
                          {{ (item.displayName || item.wxid).slice(0, 1) }}
                        </Avatar>
                        <div class="min-w-0 flex-1">
                          <div class="flex items-start justify-between gap-2">
                            <div class="min-w-0">
                              <div class="truncate text-sm font-semibold text-slate-900">
                                {{ item.displayName || item.wxid }}
                              </div>
                              <div class="truncate text-xs text-slate-500">{{ item.wxid }}</div>
                            </div>
                            <Button
                              v-if="canOpenConversation(item)"
                              size="small"
                              type="link"
                              class="!px-0"
                              @click.stop="openMessageConversation(item)"
                            >
                              发消息
                            </Button>
                          </div>
                          <div class="mt-1.5 flex flex-wrap items-center gap-1.5">
                            <Tag :bordered="false" :color="typeTagColor(item.contactType)">
                              {{ mapType(item.contactType) }}
                            </Tag>
                            <Tag v-if="showCategoryTag(item)" :bordered="false">{{ item.contactCategoryLabel }}</Tag>
                            <Tag v-if="item.memberCount" :bordered="false">
                              {{ item.memberCount }} 人
                            </Tag>
                            <Tag v-if="isRelationSupported(item)" :bordered="false" :color="relationMeta(item.wxid, item.contactType).color">
                              {{ relationMeta(item.wxid, item.contactType).label }}
                            </Tag>
                          </div>
                        </div>
                      </div>
                    </button>
                  </div>
                </section>
              </div>
              </Spin>
              <Empty v-if="!loading.contacts && !filteredContacts.length" class="pt-10" description="暂无匹配联系人" />
              <div v-if="loading.contactsMore" class="py-4 text-center text-xs text-slate-400">
                正在继续加载联系人...
              </div>
            </div>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col bg-[#f8fafc]">
          <template v-if="currentContact">
            <header class="border-b border-slate-200 bg-white px-5 py-4">
              <div class="space-y-4">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <div class="flex min-w-0 items-center gap-4">
                    <Avatar :size="52" :src="currentContact.avatar || undefined">
                      {{ (currentContact.displayName || currentContact.wxid).slice(0, 1) }}
                    </Avatar>
                    <div class="min-w-0">
                      <Typography.Title :level="4" class="!mb-1 truncate">
                        {{ currentContact.displayName || currentContact.wxid }}
                      </Typography.Title>
                      <Typography.Text type="secondary">{{ currentContact.wxid }}</Typography.Text>
                      <div class="mt-2 flex flex-wrap gap-2">
                        <Tag :bordered="false" :color="typeTagColor(currentContact.contactType)">
                          {{ mapType(currentContact.contactType) }}
                        </Tag>
                        <Tag v-if="showCategoryTag(currentContact)" :bordered="false">
                          {{ currentContact.contactCategoryLabel }}
                        </Tag>
                        <Tag
                          v-if="isRelationSupported(currentContact)"
                          :bordered="false"
                          :color="relationMeta(currentContact.wxid, currentContact.contactType).color"
                        >
                          {{ relationMeta(currentContact.wxid, currentContact.contactType).label }}
                        </Tag>
                      </div>
                    </div>
                  </div>

                  <div class="flex flex-wrap items-center justify-end gap-2">
                    <Button class="!border-slate-300 !text-slate-700" @click="confirmReloadContacts">重新加载联系人</Button>
                    <Button
                      :disabled="!selectedRelationTargets.length"
                      :loading="loading.batchRelation"
                      class="!border-sky-200 !text-sky-600"
                      @click="checkRelationBatch"
                    >
                      批量好友检测
                    </Button>
                    <Button
                      v-if="canOpenConversation(currentContact)"
                      type="primary"
                      @click="openMessageConversation(currentContact)"
                    >
                      发消息
                    </Button>
                  </div>
                </div>

                <div class="space-y-3 border-t border-slate-200 pt-4">
                  <div class="flex flex-wrap items-center gap-x-6 gap-y-3 text-sm">
                    <div
                      v-for="item in contactMetaItems"
                      :key="item.label"
                      class="flex min-w-[160px] items-center gap-2"
                    >
                      <span class="text-xs font-medium uppercase tracking-[0.14em] text-slate-400">
                        {{ item.label }}
                      </span>
                      <span class="truncate font-medium text-slate-900">
                        {{ item.value }}
                      </span>
                    </div>
                    <div
                      v-if="currentContact.contactType === 'group'"
                      class="flex min-w-[160px] items-center gap-2"
                    >
                      <span class="text-xs font-medium uppercase tracking-[0.14em] text-slate-400">
                        成员数
                      </span>
                      <span class="truncate font-medium text-slate-900">
                        {{ currentContact.memberCount || 0 }}
                      </span>
                    </div>
                  </div>

                  <div class="flex flex-wrap items-start justify-between gap-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                    <div class="min-w-0 flex-1">
                      <div class="text-sm font-semibold text-slate-900">联系人说明</div>
                      <div class="mt-3 text-sm leading-6 text-slate-600">
                        {{ currentContact.signature || '暂无签名信息，可继续在下方资料区查看完整联系人信息。' }}
                      </div>
                    </div>

                    <div class="flex flex-wrap items-center gap-2">
                      <Button
                        :disabled="!isRelationSupported(currentContact)"
                        :loading="!!relationLoadingMap[currentContact.wxid]"
                        size="small"
                        @click="checkRelation(currentContact.wxid)"
                      >
                        好友检测
                      </Button>
                      <Dropdown v-if="isRelationSupported(currentContact)" :trigger="['click']">
                        <Button size="small">更多操作</Button>
                        <template #overlay>
                          <Menu>
                            <Menu.Item key="delete" @click="confirmDeleteCurrent">删除好友</Menu.Item>
                            <Menu.Item key="block" @click="setCurrentBlacklist('add')">加入黑名单</Menu.Item>
                            <Menu.Item key="unblock" @click="setCurrentBlacklist('remove')">移出黑名单</Menu.Item>
                          </Menu>
                        </template>
                      </Dropdown>
                    </div>
                  </div>
                </div>
              </div>
            </header>

            <div class="min-h-0 flex-1 overflow-auto p-4">
              <div class="space-y-4">
                <section class="rounded-2xl border border-slate-200 bg-white p-5 shadow-[0_12px_32px_rgba(15,23,42,0.05)]">
                  <div class="mb-4 flex items-center justify-between gap-3">
                    <div>
                      <Typography.Title :level="5" class="!mb-0">联系人资料</Typography.Title>
                      <Typography.Text type="secondary">
                        当前联系人资料与会话信息均从本地缓存优先读取，缺失时再回源补全。
                      </Typography.Text>
                    </div>
                    <Tag :bordered="false" color="blue">{{ mapType(currentContact.contactType) }}</Tag>
                  </div>

                  <div class="grid gap-x-8 gap-y-5 md:grid-cols-2 xl:grid-cols-3">
                    <div>
                      <div class="text-xs text-slate-400">显示名</div>
                      <div class="mt-1 text-[15px] font-medium text-slate-800">{{ currentContact.displayName || '-' }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-400">微信ID</div>
                      <div class="mt-1 break-all text-[15px] font-medium text-slate-800">{{ currentContact.wxid }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-400">昵称</div>
                      <div class="mt-1 text-[15px] font-medium text-slate-800">{{ currentContact.nickname || '-' }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-400">备注</div>
                      <div class="mt-1 text-[15px] font-medium text-slate-800">{{ currentContact.remark || '-' }}</div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-400">地区</div>
                      <div class="mt-1 text-[15px] font-medium text-slate-800">
                        {{ `${currentContact.province || ''}${currentContact.city || ''}` || '-' }}
                      </div>
                    </div>
                    <div>
                      <div class="text-xs text-slate-400">远端昵称</div>
                      <div class="mt-1 text-[15px] font-medium text-slate-800">{{ currentRelation?.nickName || '-' }}</div>
                    </div>
                    <template v-if="currentContact.contactType === 'group'">
                      <div>
                        <div class="text-xs text-slate-400">群公告</div>
                        <div class="mt-1 text-[15px] font-medium text-slate-800">{{ currentContact.announcement || '-' }}</div>
                      </div>
                      <div>
                        <div class="text-xs text-slate-400">群主</div>
                        <div class="mt-1 break-all text-[15px] font-medium text-slate-800">{{ currentContact.chatRoomOwner || '-' }}</div>
                      </div>
                    </template>
                    <div class="md:col-span-2 xl:col-span-3">
                      <div class="text-xs text-slate-400">签名</div>
                      <div class="mt-1 text-[15px] leading-7 font-medium text-slate-800">
                        {{ currentContact.signature || '-' }}
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            </div>
          </template>
          <Empty v-else class="pt-20" description="请选择左侧联系人查看详情" />
        </section>
      </div>
    </div>

    <Modal
      v-model:open="addContactModalOpen"
      :footer="null"
      :mask-closable="true"
      :width="620"
      title="搜索并添加联系人"
    >
      <div class="max-h-[70vh] overflow-auto pr-1">
        <Space class="w-full" direction="vertical" size="small">
          <Input.TextArea
            v-model:value="searchKeyword"
            :auto-size="{ minRows: 3, maxRows: 6 }"
            placeholder="输入微信号 / 昵称 / 手机号，支持换行、空格、逗号批量搜索"
          />
          <div class="flex items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-500">
            <span>当前待搜索 {{ searchTerms.length }} 个关键词</span>
            <Button
              :loading="loading.search"
              type="primary"
              @click="searchFriendCandidates"
            >
              {{ searchTerms.length > 1 ? '批量搜索' : '搜索' }}
            </Button>
          </div>
          <Input
            v-model:value="verifyContent"
            placeholder="验证消息（可选）"
          />
          <div class="flex items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
            <Checkbox
              :checked="allAddSelected"
              :indeterminate="addIndeterminate"
              :disabled="!addableSearchResults.length"
              @change="(event) => toggleSelectAllSearchResults(event.target.checked)"
            >
              全选可添加结果
            </Checkbox>
            <Button
              :disabled="!selectedAddTargets.length"
              :loading="loading.batchAdd"
              type="primary"
              @click="sendFriendRequestBatch"
            >
              批量添加已勾选
            </Button>
          </div>
          <Typography.Text class="!text-xs !text-slate-500">
            支持一次输入多个关键词批量搜索；结果会自动聚合去重，再按勾选项批量添加。
          </Typography.Text>
        </Space>

        <div class="my-4 border-t border-slate-200"></div>

        <Spin :spinning="loading.search">
          <Space class="w-full" direction="vertical" size="small">
            <div
              v-for="item in searchResults"
              :key="`${item.source}-${item.wxid}`"
              class="rounded-xl border border-slate-100 bg-slate-50 p-3"
            >
              <div class="flex items-start gap-3">
                <div class="pt-1" @click.stop>
                  <Checkbox
                    :checked="selectedSearchKeys.includes(getSearchResultKey(item))"
                    :disabled="!canAddFriend(item)"
                    @change="(event) => toggleSearchSelection(getSearchResultKey(item), event.target.checked)"
                  />
                </div>
                <Avatar :src="item.avatar || undefined">
                  {{ (item.displayName || item.wxid).slice(0, 1) }}
                </Avatar>
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium text-slate-900">
                    {{ item.displayName || item.wxid }}
                  </div>
                  <div class="truncate text-xs text-slate-500">{{ item.wxid }}</div>
                  <div class="mt-2 flex items-center gap-2">
                    <Tag :bordered="false">{{ item.source === 'primary' ? '主结果' : '候选' }}</Tag>
                    <Tag :bordered="false" :color="canAddFriend(item) ? 'success' : 'default'">
                      {{ canAddFriend(item) ? '可添加' : '缺少参数' }}
                    </Tag>
                  </div>
                </div>
                <Button
                  :disabled="!canAddFriend(item)"
                  :loading="addLoadingMap[item.wxid]"
                  size="small"
                  type="primary"
                  @click="sendFriendRequest(item)"
                >
                  添加好友
                </Button>
              </div>
            </div>
          </Space>
        </Spin>
        <Empty v-if="!loading.search && !searchResults.length" class="pt-8" description="搜索结果将显示在这里" />
      </div>
    </Modal>
  </Page>
</template>
