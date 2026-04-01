<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Dropdown,
  Empty,
  Input,
  Menu,
  Modal,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import type { TableRowSelection } from 'ant-design-vue/es/table/interface';

import {
  addGroupFriendApi,
  addGroupMembersApi,
  getConversationDetailApi,
  inviteGroupMembersApi,
  listContactsApi,
  listGroupMembersApi,
  operateGroupAdminApi,
  quitGroupApi,
  refreshGroupApi,
  removeGroupMembersApi,
  setGroupAddressBookApi,
  updateGroupAnnouncementApi,
  updateGroupNameApi,
  updateGroupRemarkApi,
  type ContactSummary,
  type ConversationDetail,
  type GroupMemberItem,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const router = useRouter();
const accountStore = useEnterpriseAccountStore();

const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const accountOptions = computed(() => accountStore.accountOptions);

const groups = ref<ContactSummary[]>([]);
const selectedGroupId = ref('');
const search = ref('');

const loading = reactive({
  accounts: false,
  contacts: false,
  groups: false,
  detail: false,
  members: false,
  action: false,
});

const groupDetail = ref<ConversationDetail | null>(null);
const groupMembers = ref<GroupMemberItem[]>([]);
const selectedMemberKeys = ref<string[]>([]);
const directContactWxids = ref<Set<string>>(new Set());

const editModal = reactive({
  action: '' as '' | 'announcement' | 'name' | 'remark',
  confirmText: '保存',
  open: false,
  placeholder: '',
  title: '',
  value: '',
});

const memberModal = reactive({
  action: '' as '' | 'add' | 'invite',
  confirmText: '提交',
  open: false,
  placeholder: '',
  title: '',
  value: '',
});

const addFriendModal = reactive({
  open: false,
  targets: [] as string[],
  verifyContent: '你好，想加你为好友',
});

const filteredGroups = computed(() => {
  const keyword = search.value.trim().toLowerCase();
  if (!keyword) return groups.value;
  return groups.value.filter((item) =>
    [item.displayName, item.nickname, item.remark, item.wxid]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword)),
  );
});

const currentGroup = computed(() =>
  filteredGroups.value.find((item) => item.wxid === selectedGroupId.value)
  ?? groups.value.find((item) => item.wxid === selectedGroupId.value),
);

const groupMetaItems = computed(() => [
  { label: '群主', value: groupDetail.value?.chatRoomOwner || currentGroup.value?.chatRoomOwner || '-' },
  { label: '成员数', value: String(groupDetail.value?.memberCount || currentGroup.value?.memberCount || 0) },
  { label: '群备注', value: groupDetail.value?.remark || currentGroup.value?.remark || '未设置' },
  { label: '当前身份', value: isCurrentGroupOwner.value ? '群主' : isCurrentGroupAdmin.value ? '管理员' : '普通成员' },
]);

const selectedMembers = computed(() =>
  groupMembers.value.filter((item) => selectedMemberKeys.value.includes(item.userName)),
);

const selectedMemberTargets = computed(() => selectedMembers.value.map((item) => item.userName));

const currentMember = computed(() =>
  groupMembers.value.find((item) => item.userName === selectedWxid.value),
);

const isCurrentGroupOwner = computed(
  () => !!selectedWxid.value && groupDetail.value?.chatRoomOwner === selectedWxid.value,
);

const isCurrentGroupAdmin = computed(
  () => isCurrentGroupOwner.value || isAdminFlag(currentMember.value?.chatroomMemberFlag || 0),
);

const rowSelection = computed<TableRowSelection<GroupMemberItem>>(() => ({
  onChange: (keys) => {
    selectedMemberKeys.value = keys.map((item) => String(item));
  },
  selectedRowKeys: selectedMemberKeys.value,
}));

const memberColumns = [
  {
    dataIndex: 'member',
    key: 'member',
    title: '成员信息',
    width: 360,
  },
  {
    dataIndex: 'inviterUserName',
    key: 'inviterUserName',
    title: '邀请人',
    width: 180,
  },
  {
    dataIndex: 'chatroomMemberFlag',
    key: 'chatroomMemberFlag',
    title: '身份',
    width: 110,
  },
  {
    key: 'action',
    title: '操作',
    width: 220,
  },
];

function isAdminFlag(flag: number) {
  return [9, 25, 8193].includes(flag);
}

function memberRoleLabel(flag: number) {
  return isAdminFlag(flag) ? '管理员' : '普通成员';
}

function memberContactMeta(userName: string) {
  if (!userName) {
    return { color: 'default', isFriend: false, label: '未知状态' };
  }
  if (userName === selectedWxid.value) {
    return { color: 'blue', isFriend: true, label: '当前账号' };
  }
  if (directContactWxids.value.has(userName)) {
    return { color: 'success', isFriend: true, label: '已是联系人' };
  }
  return { color: 'default', isFriend: false, label: '可加好友' };
}

function resetGroupWorkspace() {
  groupDetail.value = null;
  groupMembers.value = [];
  selectedMemberKeys.value = [];
}

function openMessageConversation(targetWxid: string) {
  if (!selectedWxid.value || !targetWxid) return;
  router.push({
    name: 'EnterpriseMessages',
    query: {
      conversation: targetWxid,
      wxid: selectedWxid.value,
    },
  });
}

function jumpToGroupChat() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  openMessageConversation(selectedGroupId.value);
}

function parseTargets(raw: string) {
  return raw
    .split(/[\n,，;\s]+/g)
    .map((item) => item.trim())
    .filter(Boolean);
}

function openEditModal(action: 'announcement' | 'name' | 'remark') {
  editModal.action = action;
  editModal.open = true;
  if (action === 'name') {
    editModal.title = '修改群名称';
    editModal.placeholder = '请输入新的群名称';
    editModal.value = groupDetail.value?.targetName || currentGroup.value?.displayName || '';
  } else if (action === 'announcement') {
    editModal.title = '修改群公告';
    editModal.placeholder = '请输入新的群公告';
    editModal.value = groupDetail.value?.announcement || currentGroup.value?.announcement || '';
  } else {
    editModal.title = '修改群备注';
    editModal.placeholder = '请输入你自己的群备注';
    editModal.value = groupDetail.value?.remark || currentGroup.value?.remark || '';
  }
}

function openMemberModal(action: 'add' | 'invite') {
  memberModal.action = action;
  memberModal.open = true;
  memberModal.title = action === 'add' ? '添加群成员（40人内）' : '邀请群成员（40人以上）';
  memberModal.placeholder = '请输入要操作的 wxid，支持换行、空格、逗号分隔';
  memberModal.confirmText = action === 'add' ? '添加成员' : '发送邀请';
  memberModal.value = '';
}

function openAddFriendModal(targets?: string[]) {
  const resolved = (targets && targets.length ? targets : selectedMemberTargets.value).filter(Boolean);
  if (!resolved.length) {
    message.warning('请先选择群成员');
    return;
  }
  const addableTargets = resolved.filter((target) => !memberContactMeta(target).isFriend);
  if (!addableTargets.length) {
    message.info('所选成员都已经在联系人列表中');
    return;
  }
  if (addableTargets.length < resolved.length) {
    message.info(`已自动跳过 ${resolved.length - addableTargets.length} 个已是联系人的成员`);
  }
  addFriendModal.targets = addableTargets;
  addFriendModal.open = true;
}

async function loadAccounts() {
  loading.accounts = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loading.accounts = false;
  }
}

async function loadGroups() {
  if (!selectedWxid.value) return;
  loading.groups = true;
  try {
    const merged = new Map<string, ContactSummary>();
    let page = 1;
    let totalPages = 1;
    do {
      const result = await listContactsApi(selectedWxid.value, {
        category: 'group',
        contactType: 'group',
        keyword: search.value.trim() || undefined,
        page,
        pageSize: 500,
      });
      result.contacts.forEach((item) => {
        merged.set(item.wxid, item);
      });
      totalPages = result.totalPages || 1;
      page += 1;
    } while (page <= totalPages);

    groups.value = [...merged.values()].sort((left, right) =>
      String(left.displayName || left.wxid).localeCompare(String(right.displayName || right.wxid), 'zh-CN'),
    );

    if (!selectedGroupId.value || !groups.value.some((item) => item.wxid === selectedGroupId.value)) {
      selectedGroupId.value = groups.value[0]?.wxid ?? '';
    }
  } finally {
    loading.groups = false;
  }
}

async function loadDirectContacts() {
  if (!selectedWxid.value) {
    directContactWxids.value = new Set();
    return;
  }
  loading.contacts = true;
  try {
    const nextSet = new Set<string>();
    let page = 1;
    let totalPages = 1;
    do {
      const result = await listContactsApi(selectedWxid.value, {
        category: 'contact',
        contactType: 'contact',
        page,
        pageSize: 1000,
      });
      result.contacts.forEach((item) => {
        if (item.wxid) {
          nextSet.add(item.wxid);
        }
      });
      totalPages = result.totalPages || 1;
      page += 1;
    } while (page <= totalPages);
    directContactWxids.value = nextSet;
  } finally {
    loading.contacts = false;
  }
}

async function loadGroupDetail() {
  if (!selectedWxid.value || !selectedGroupId.value) {
    groupDetail.value = null;
    return;
  }
  loading.detail = true;
  try {
    groupDetail.value = await getConversationDetailApi(selectedWxid.value, selectedGroupId.value);
  } finally {
    loading.detail = false;
  }
}

async function loadGroupMembers() {
  if (!selectedWxid.value || !selectedGroupId.value) {
    groupMembers.value = [];
    return;
  }
  loading.members = true;
  try {
    const result = await listGroupMembersApi(selectedWxid.value, selectedGroupId.value);
    groupMembers.value = result.members;
  } finally {
    loading.members = false;
  }
}

async function loadGroupWorkspace() {
  if (!selectedGroupId.value) {
    resetGroupWorkspace();
    return;
  }
  await Promise.all([loadGroupDetail(), loadGroupMembers()]);
}

async function handleRefreshGroup() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  loading.action = true;
  try {
    groupDetail.value = await refreshGroupApi(selectedWxid.value, selectedGroupId.value);
    await Promise.all([loadGroups(), loadGroupMembers()]);
    message.success('群资料已刷新');
  } finally {
    loading.action = false;
  }
}

async function handleSaveGroupAddressBook(enabled: boolean) {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  loading.action = true;
  try {
    const result = await setGroupAddressBookApi(selectedWxid.value, selectedGroupId.value, { enabled });
    message.success(result.message);
  } finally {
    loading.action = false;
  }
}

async function submitEditModal() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  loading.action = true;
  try {
    const payload = { content: editModal.value.trim() };
    if (editModal.action === 'name') {
      groupDetail.value = await updateGroupNameApi(selectedWxid.value, selectedGroupId.value, payload);
      message.success('群名称已更新');
    } else if (editModal.action === 'announcement') {
      groupDetail.value = await updateGroupAnnouncementApi(selectedWxid.value, selectedGroupId.value, payload);
      message.success('群公告已更新');
    } else {
      groupDetail.value = await updateGroupRemarkApi(selectedWxid.value, selectedGroupId.value, payload);
      message.success('群备注已更新');
    }
    await loadGroups();
    editModal.open = false;
  } finally {
    loading.action = false;
  }
}

async function submitMemberModal() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  const targets = parseTargets(memberModal.value);
  if (!targets.length) {
    message.warning('请先输入成员 wxid');
    return;
  }
  loading.action = true;
  try {
    if (memberModal.action === 'add') {
      const result = await addGroupMembersApi(selectedWxid.value, selectedGroupId.value, { targets });
      message.success(result.message);
    } else {
      const result = await inviteGroupMembersApi(selectedWxid.value, selectedGroupId.value, { targets });
      message.success(result.message);
    }
    memberModal.open = false;
    await handleRefreshGroup();
  } finally {
    loading.action = false;
  }
}

async function handleRemoveMembers(targets?: string[]) {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  if (!isCurrentGroupAdmin.value) {
    message.warning('仅群主或群管理员可移出成员');
    return;
  }
  const resolved = (targets && targets.length ? targets : selectedMemberTargets.value).filter(Boolean);
  if (!resolved.length) {
    message.warning('请先选择要移出的群成员');
    return;
  }
  Modal.confirm({
    centered: true,
    content: `确定移出 ${resolved.length} 个群成员吗？`,
    title: '移出群成员',
    async onOk() {
      loading.action = true;
      try {
        const result = await removeGroupMembersApi(selectedWxid.value!, selectedGroupId.value!, { targets: resolved });
        message.success(result.message);
        selectedMemberKeys.value = [];
        await handleRefreshGroup();
      } finally {
        loading.action = false;
      }
    },
  });
}

async function handleOperateAdmin(targets: string[], val: number) {
  if (!selectedWxid.value || !selectedGroupId.value || !targets.length) {
    message.warning('请先选择群成员');
    return;
  }
  if (!isCurrentGroupOwner.value) {
    message.warning('仅群主可设置或取消管理员');
    return;
  }
  loading.action = true;
  try {
    const result = await operateGroupAdminApi(selectedWxid.value, selectedGroupId.value, { targets, val });
    message.success(result.message);
    await loadGroupMembers();
  } finally {
    loading.action = false;
  }
}

async function handleMemberMenuAction(record: GroupMemberItem, action: string) {
  if (action === 'message') {
    openMessageConversation(record.userName);
    return;
  }
  if (action === 'add-friend') {
    if (memberContactMeta(record.userName).isFriend) {
      message.info('该成员已经在联系人列表中');
      return;
    }
    openAddFriendModal([record.userName]);
    return;
  }
  if (action === 'set-admin') {
    await handleOperateAdmin([record.userName], 1);
    return;
  }
  if (action === 'unset-admin') {
    await handleOperateAdmin([record.userName], 2);
    return;
  }
  if (action === 'remove') {
    await handleRemoveMembers([record.userName]);
  }
}

async function handleHeaderMenuAction(action: string) {
  if (action === 'save-address-book') {
    await handleSaveGroupAddressBook(true);
    return;
  }
  if (action === 'remove-address-book') {
    await handleSaveGroupAddressBook(false);
    return;
  }
  if (action === 'edit-name') {
    openEditModal('name');
    return;
  }
  if (action === 'edit-announcement') {
    openEditModal('announcement');
    return;
  }
  if (action === 'edit-remark') {
    openEditModal('remark');
  }
}

async function submitAddFriendModal() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  if (!addFriendModal.targets.length) {
    message.warning('请先选择群成员');
    return;
  }
  loading.action = true;
  try {
    const result = await addGroupFriendApi(selectedWxid.value, selectedGroupId.value, {
      targets: addFriendModal.targets,
      verifyContent: addFriendModal.verifyContent.trim(),
    });
    const successCount = result.items.filter((item) => item.success).length;
    const failed = result.items.filter((item) => !item.success);
    if (failed.length) {
      message.warning(`已发起 ${successCount} 个好友申请，${failed.length} 个失败`);
    } else {
      message.success(`已发起 ${successCount} 个好友申请`);
    }
    addFriendModal.open = false;
  } finally {
    loading.action = false;
  }
}

function handleQuitGroup() {
  if (!selectedWxid.value || !selectedGroupId.value) return;
  Modal.confirm({
    centered: true,
    content: '退出后该群会从当前账号的后台列表中移除，确定继续吗？',
    okButtonProps: { danger: true },
    okText: '退出群聊',
    title: '退出群聊',
    async onOk() {
      loading.action = true;
      try {
        const result = await quitGroupApi(selectedWxid.value!, selectedGroupId.value!);
        message.success(result.message);
        selectedGroupId.value = '';
        resetGroupWorkspace();
        await loadGroups();
      } finally {
        loading.action = false;
      }
    },
  });
}

watch(selectedWxid, async (next) => {
  resetGroupWorkspace();
  groups.value = [];
  selectedGroupId.value = '';
  directContactWxids.value = new Set();
  if (!next) return;
  await Promise.all([loadGroups(), loadDirectContacts()]);
});

watch(selectedGroupId, async () => {
  selectedMemberKeys.value = [];
  await loadGroupWorkspace();
});

onMounted(async () => {
  await loadAccounts();
  if (selectedWxid.value) {
    await Promise.all([loadGroups(), loadDirectContacts()]);
  }
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
        class="!w-[320px]"
        placeholder="请选择账号"
      />
      <Button @click="loadGroups">刷新群列表</Button>
    </div>

    <div class="h-[calc(100vh-220px)] min-h-[680px] overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <div class="flex h-full">
        <aside class="flex h-full w-[360px] min-w-[340px] flex-col border-r border-slate-200 bg-slate-50/60 p-4">
          <div class="mb-3 rounded-2xl border border-slate-200 bg-white px-4 py-4">
            <div class="flex items-center justify-between">
              <div>
                <Typography.Title :level="4" class="!mb-0">群管理</Typography.Title>
                <Typography.Text type="secondary">当前账号群聊列表</Typography.Text>
              </div>
              <Tag :bordered="false" color="blue">{{ groups.length }}</Tag>
            </div>
            <Input.Search
              v-model:value="search"
              allow-clear
              class="!mt-3"
              placeholder="搜索群名称 / 群 ID"
              @search="loadGroups"
            />
          </div>

          <div class="min-h-0 flex-1 overflow-auto pr-1">
            <Spin :spinning="loading.groups">
              <Space class="w-full" direction="vertical" size="small">
                <button
                  v-for="item in filteredGroups"
                  :key="item.wxid"
                  :class="[
                    'w-full rounded-2xl border px-4 py-3 text-left transition',
                    selectedGroupId === item.wxid
                      ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))]'
                      : 'border-transparent bg-white hover:border-slate-200',
                  ]"
                  type="button"
                  @click="selectedGroupId = item.wxid"
                >
                  <div class="flex items-start gap-3">
                    <Avatar :size="42" :src="item.avatar || undefined">
                      {{ (item.displayName || item.wxid).slice(0, 1) }}
                    </Avatar>
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-sm font-semibold text-slate-900">
                        {{ item.displayName || item.nickname || item.wxid }}
                      </div>
                      <div class="mt-1 truncate text-xs text-slate-500">{{ item.wxid }}</div>
                      <div class="mt-2 flex flex-wrap gap-2">
                        <Tag :bordered="false" color="purple">{{ item.memberCount || 0 }} 人</Tag>
                        <Tag v-if="item.chatRoomOwner" :bordered="false">群主 {{ item.chatRoomOwner }}</Tag>
                      </div>
                    </div>
                  </div>
                </button>
              </Space>
              <Empty v-if="!filteredGroups.length && !loading.groups" class="pt-16" description="暂无群聊" />
            </Spin>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col bg-slate-50">
          <template v-if="currentGroup">
            <header class="border-b border-slate-200 bg-white px-5 py-4">
              <div class="space-y-4">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <div class="flex min-w-0 items-center gap-4">
                    <Avatar :size="52" :src="currentGroup.avatar || undefined">
                      {{ (currentGroup.displayName || currentGroup.wxid).slice(0, 1) }}
                    </Avatar>
                    <div class="min-w-0">
                      <Typography.Title :level="4" class="!mb-1 truncate">
                        {{ groupDetail?.targetName || currentGroup.displayName || currentGroup.wxid }}
                      </Typography.Title>
                      <Typography.Text type="secondary">{{ currentGroup.wxid }}</Typography.Text>
                      <div class="mt-2 flex flex-wrap gap-2">
                        <Tag :bordered="false" color="purple">群聊</Tag>
                        <Tag :bordered="false" color="blue">{{ groupDetail?.memberCount || currentGroup.memberCount || 0 }} 人</Tag>
                      </div>
                    </div>
                  </div>

                  <div class="flex flex-wrap items-center justify-end gap-2">
                    <Button class="!border-emerald-200 !text-emerald-600" @click="handleRefreshGroup">刷新群资料</Button>
                    <Button class="!border-violet-200 !text-violet-600" @click="openMemberModal('add')">添加成员</Button>
                    <Button class="!border-orange-200 !text-orange-600" @click="openMemberModal('invite')">邀请成员</Button>
                    <Dropdown :trigger="['click']">
                      <Button>更多操作</Button>
                      <template #overlay>
                        <Menu @click="({ key }) => handleHeaderMenuAction(String(key))">
                          <Menu.Item key="save-address-book">保存到通讯录</Menu.Item>
                          <Menu.Item key="remove-address-book">移出通讯录</Menu.Item>
                          <Menu.Divider />
                          <Menu.Item key="edit-name">改群名称</Menu.Item>
                          <Menu.Item key="edit-announcement">改群公告</Menu.Item>
                          <Menu.Item key="edit-remark">改群备注</Menu.Item>
                        </Menu>
                      </template>
                    </Dropdown>
                    <Button type="primary" @click="jumpToGroupChat">进入会话</Button>
                  </div>
                </div>

                <div class="space-y-3 border-t border-slate-200 pt-4">
                  <div class="flex flex-wrap items-center gap-x-6 gap-y-3 text-sm">
                    <div
                      v-for="item in groupMetaItems"
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
                  </div>

                  <div class="flex flex-wrap items-start justify-between gap-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                    <div class="min-w-0 flex-1">
                      <div class="text-sm font-semibold text-slate-900">群公告</div>
                      <div class="mt-3 text-sm text-slate-600">
                        <span class="whitespace-pre-wrap">{{ groupDetail?.announcement || currentGroup.announcement || '暂无公告' }}</span>
                      </div>
                    </div>

                    <div class="flex flex-wrap items-center gap-2">
                      <Button size="small" @click="openAddFriendModal()" :disabled="!selectedMemberTargets.length">给选中成员加好友</Button>
                      <Button danger size="small" @click="handleQuitGroup">退出群聊</Button>
                    </div>
                  </div>
                </div>
              </div>
            </header>

            <div class="min-h-0 flex-1 overflow-auto px-5 py-4">
              <Spin :spinning="loading.detail || loading.members || loading.action">
                <div class="space-y-4">
                  <section class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <Typography.Title :level="5" class="!mb-0">群成员</Typography.Title>
                        <Typography.Text type="secondary">
                          已选 {{ selectedMemberTargets.length }} 人，可批量加好友、移出群或设置管理员。
                        </Typography.Text>
                      </div>
                      <div class="flex flex-wrap gap-2">
                        <Button
                          :disabled="!selectedMemberTargets.length"
                          class="!border-violet-200 !text-violet-600"
                          @click="openAddFriendModal()"
                        >
                          批量加好友
                        </Button>
                        <Button
                          class="!border-blue-200 !text-blue-600"
                          :disabled="!selectedMemberTargets.length || !isCurrentGroupOwner"
                          @click="handleOperateAdmin(selectedMemberTargets, 1)"
                        >
                          设为管理员
                        </Button>
                        <Button
                          :disabled="!selectedMemberTargets.length || !isCurrentGroupOwner"
                          class="!border-amber-200 !text-amber-600"
                          @click="handleOperateAdmin(selectedMemberTargets, 2)"
                        >
                          取消管理员
                        </Button>
                        <Button danger :disabled="!selectedMemberTargets.length || !isCurrentGroupAdmin" @click="handleRemoveMembers()">
                          批量移出群
                        </Button>
                      </div>
                    </div>

                    <Table
                      :columns="memberColumns"
                      :data-source="groupMembers"
                      :pagination="{ pageSize: 12 }"
                      :row-selection="rowSelection"
                      :scroll="{ x: 900 }"
                      class="group-members-table"
                      row-key="userName"
                      size="middle"
                      table-layout="fixed"
                    >
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'member'">
                          <div class="flex items-start gap-3">
                            <Avatar :size="42" :src="record.avatar || undefined">
                              {{ (record.displayName || record.userName).slice(0, 1) }}
                            </Avatar>
                            <div class="min-w-0 flex-1">
                              <div class="truncate text-sm font-semibold text-slate-900">
                                {{ record.displayName || record.nickName || record.userName }}
                              </div>
                              <div class="mt-1 truncate text-xs text-slate-500">
                                {{ record.userName }}
                              </div>
                              <div class="mt-2 flex flex-wrap gap-2">
                                <Tag
                                  v-if="record.nickName && record.nickName !== record.displayName"
                                  :bordered="false"
                                  color="blue"
                                >
                                  昵称 {{ record.nickName }}
                                </Tag>
                                <Tag v-if="record.displayName" :bordered="false" color="purple">
                                  群内名片 {{ record.displayName }}
                                </Tag>
                              </div>
                            </div>
                          </div>
                        </template>
                        <template v-else-if="column.key === 'chatroomMemberFlag'">
                          <div class="flex flex-wrap items-center gap-2">
                            <Tag :bordered="false" :color="isAdminFlag(record.chatroomMemberFlag) ? 'blue' : 'default'">
                              {{ memberRoleLabel(record.chatroomMemberFlag) }}
                            </Tag>
                            <Tag
                              v-if="record.userName !== selectedWxid"
                              :bordered="false"
                              :color="memberContactMeta(record.userName).color"
                            >
                              {{ memberContactMeta(record.userName).label }}
                            </Tag>
                          </div>
                        </template>
                        <template v-else-if="column.key === 'inviterUserName'">
                          <div class="min-w-0 text-sm">
                            <div class="truncate font-medium text-slate-700">
                              {{ record.inviterUserName || '系统加入/未知' }}
                            </div>
                            <div class="mt-1 text-xs text-slate-400">
                              {{ record.inviterUserName ? '邀请来源成员' : '无邀请人记录' }}
                            </div>
                          </div>
                        </template>
                        <template v-else-if="column.key === 'action'">
                          <div class="flex items-center justify-end gap-3 whitespace-nowrap">
                            <Button
                              class="!h-auto !px-0 !text-sm !font-medium"
                              size="small"
                              type="link"
                              @click="openMessageConversation(record.userName)"
                            >
                              发消息
                            </Button>
                            <Button
                              class="!h-auto !px-0 !text-sm !font-medium"
                              :disabled="memberContactMeta(record.userName).isFriend"
                              size="small"
                              type="link"
                              @click="openAddFriendModal([record.userName])"
                            >
                              {{ memberContactMeta(record.userName).isFriend ? '已是联系人' : '加好友' }}
                            </Button>
                            <Dropdown :trigger="['click']">
                              <Button class="!h-8 !rounded-lg !px-3 !text-sm" size="small">更多</Button>
                              <template #overlay>
                                <Menu @click="({ key }) => handleMemberMenuAction(record, String(key))">
                                  <Menu.Item key="set-admin" :disabled="!isCurrentGroupOwner">设为管理员</Menu.Item>
                                  <Menu.Item key="unset-admin" :disabled="!isCurrentGroupOwner">取消管理员</Menu.Item>
                                  <Menu.Item key="remove" :disabled="!isCurrentGroupAdmin">移出群聊</Menu.Item>
                                </Menu>
                              </template>
                            </Dropdown>
                          </div>
                        </template>
                      </template>
                    </Table>
                  </section>
                </div>
              </Spin>
            </div>
          </template>

          <div v-else class="flex h-full items-center justify-center">
            <Empty description="请选择左侧群聊" />
          </div>
        </section>
      </div>
    </div>

    <Modal
      v-model:open="editModal.open"
      :confirm-loading="loading.action"
      :ok-text="editModal.confirmText"
      :title="editModal.title"
      centered
      @ok="submitEditModal"
    >
      <Input.TextArea
        v-model:value="editModal.value"
        :auto-size="{ minRows: 4, maxRows: 8 }"
        :placeholder="editModal.placeholder"
      />
    </Modal>

    <Modal
      v-model:open="memberModal.open"
      :confirm-loading="loading.action"
      :ok-text="memberModal.confirmText"
      :title="memberModal.title"
      centered
      @ok="submitMemberModal"
    >
      <Input.TextArea
        v-model:value="memberModal.value"
        :auto-size="{ minRows: 5, maxRows: 10 }"
        :placeholder="memberModal.placeholder"
      />
    </Modal>

    <Modal
      v-model:open="addFriendModal.open"
      :confirm-loading="loading.action"
      ok-text="发起好友申请"
      title="添加群好友"
      width="560px"
      centered
      @ok="submitAddFriendModal"
    >
      <div class="space-y-4">
        <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
          <div class="text-xs font-medium uppercase tracking-[0.16em] text-slate-400">目标成员</div>
          <div class="mt-3 flex flex-wrap gap-2">
            <Tag v-for="target in addFriendModal.targets" :key="target" :bordered="false" color="blue">
              {{ target }}
            </Tag>
          </div>
        </div>
        <div>
          <div class="mb-2 text-sm font-medium text-slate-700">验证语</div>
          <Input.TextArea
            v-model:value="addFriendModal.verifyContent"
            :auto-size="{ minRows: 3, maxRows: 6 }"
            placeholder="请输入添加群好友时的验证语"
          />
        </div>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
:deep(.group-members-table .ant-table-thead > tr > th) {
  padding-bottom: 12px;
  padding-top: 12px;
}

:deep(.group-members-table .ant-table-tbody > tr > td) {
  padding-bottom: 14px;
  padding-top: 14px;
  vertical-align: top;
}

:deep(.group-members-table .ant-table-cell) {
  white-space: normal;
}

:deep(.group-members-table .ant-table-thead > tr > th:last-child) {
  text-align: right;
}
</style>
