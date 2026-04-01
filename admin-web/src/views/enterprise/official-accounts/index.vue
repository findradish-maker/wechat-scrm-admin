<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Descriptions,
  Empty,
  Input,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from 'ant-design-vue';

import {
  getConversationDetailApi,
  listContactsApi,
  listConversationMessagesApi,
  type ContactSummary,
  type ConversationDetail,
  type ConversationMessage,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const router = useRouter();

const accountsLoading = ref(false);
const officialLoading = ref(false);
const detailLoading = ref(false);
const articleLoading = ref(false);

const accountStore = useEnterpriseAccountStore();
const accounts = computed(() => accountStore.accounts);
const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const search = ref('');

const officialAccounts = ref<ContactSummary[]>([]);
const selectedOfficialId = ref('');
const officialDetail = ref<ConversationDetail | null>(null);
const articleMessages = ref<ConversationMessage[]>([]);

const accountOptions = computed(() => accountStore.accountOptions);

const currentOfficial = computed(() =>
  officialAccounts.value.find((item) => item.wxid === selectedOfficialId.value),
);

const articleRows = computed(() =>
  articleMessages.value.map((item) => ({
    key: `${item.messageId}-${item.msgId}`,
    status: item.status || 'received',
    time: item.createdAt,
    title: item.content || '[图文消息]',
    type: item.messageType,
  })),
);

const articleColumns = [
  { dataIndex: 'title', title: '消息摘要' },
  { dataIndex: 'type', title: '类型', width: 110 },
  { dataIndex: 'status', title: '状态', width: 120 },
  { dataIndex: 'time', title: '时间', width: 160 },
];

function formatTime(timestamp: number) {
  if (!timestamp) return '-';
  const millis = timestamp > 10_000_000_000 ? timestamp : timestamp * 1000;
  return new Date(millis).toLocaleString('zh-CN', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: '2-digit',
  });
}

async function loadAccounts() {
  accountsLoading.value = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    accountsLoading.value = false;
  }
}

async function loadOfficialAccounts() {
  if (!selectedWxid.value) return;
  officialLoading.value = true;
  try {
    const result = await listContactsApi(selectedWxid.value, {
      contactType: 'official_account',
      keyword: search.value.trim() || undefined,
      page: 1,
      pageSize: 200,
    });
    officialAccounts.value = result.contacts;
    if (!selectedOfficialId.value || !officialAccounts.value.some((item) => item.wxid === selectedOfficialId.value)) {
      selectedOfficialId.value = officialAccounts.value[0]?.wxid ?? '';
    }
  } finally {
    officialLoading.value = false;
  }
}

async function loadOfficialDetail() {
  if (!selectedWxid.value || !selectedOfficialId.value) {
    officialDetail.value = null;
    return;
  }
  detailLoading.value = true;
  try {
    officialDetail.value = await getConversationDetailApi(selectedWxid.value, selectedOfficialId.value);
  } finally {
    detailLoading.value = false;
  }
}

async function loadRecentMessages() {
  if (!selectedWxid.value || !selectedOfficialId.value) {
    articleMessages.value = [];
    return;
  }
  articleLoading.value = true;
  try {
    const result = await listConversationMessagesApi(selectedWxid.value, selectedOfficialId.value, {
      page: 1,
      pageSize: 20,
    });
    articleMessages.value = result.messages;
  } finally {
    articleLoading.value = false;
  }
}

function jumpToChat() {
  if (!selectedOfficialId.value) return;
  router.push({
    name: 'EnterpriseMessages',
    query: {
      conversation: selectedOfficialId.value,
      wxid: selectedWxid.value,
    },
  });
}

watch(selectedWxid, async () => {
  selectedOfficialId.value = '';
  officialDetail.value = null;
  articleMessages.value = [];
  await loadOfficialAccounts();
});

watch(selectedOfficialId, async () => {
  await Promise.all([loadOfficialDetail(), loadRecentMessages()]);
});

onMounted(async () => {
  await loadAccounts();
  if (selectedWxid.value) {
    await loadOfficialAccounts();
  }
});
</script>

<template>
  <Page>
    <div class="mb-4 flex flex-wrap items-center gap-3">
      <Typography.Text type="secondary">当前账号</Typography.Text>
      <Select
        v-model:value="selectedWxid"
        :loading="accountsLoading"
        :options="accountOptions"
        class="!w-[320px]"
        placeholder="请选择账号"
      />
      <Button @click="loadOfficialAccounts">刷新列表</Button>
    </div>

    <div class="h-[calc(100vh-220px)] min-h-[640px] overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div class="flex h-full">
        <aside class="flex h-full w-[340px] min-w-[320px] flex-col border-r border-slate-200 bg-slate-50/50 p-4">
          <div class="mb-3 rounded-2xl border border-slate-200 bg-white p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-[18px] font-semibold text-slate-900">公众号</div>
                <div class="mt-1 text-sm text-slate-500">当前账号公众号列表</div>
              </div>
              <Tag :bordered="false" color="blue">{{ officialAccounts.length }}</Tag>
            </div>
            <Input.Search
              v-model:value="search"
              allow-clear
              class="!mt-3"
              placeholder="搜索公众号名称 / gh_id"
              @search="loadOfficialAccounts"
            />
          </div>

          <div class="min-h-0 flex-1 overflow-auto pr-1">
            <Spin :spinning="officialLoading">
              <Space class="w-full" direction="vertical" size="small">
                <button
                  v-for="item in officialAccounts"
                  :key="item.wxid"
                  :class="[
                    'w-full rounded-xl border p-3 text-left transition',
                    selectedOfficialId === item.wxid
                      ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))]'
                      : 'border-transparent bg-white hover:border-slate-200',
                  ]"
                  type="button"
                  @click="selectedOfficialId = item.wxid"
                >
                  <div class="flex items-start gap-3">
                    <Avatar :src="item.avatar || undefined">
                      {{ (item.displayName || item.wxid).slice(0, 1) }}
                    </Avatar>
                    <div class="min-w-0 flex-1">
                      <div class="truncate text-sm font-medium text-slate-900">
                        {{ item.displayName || item.nickname || item.wxid }}
                      </div>
                      <div class="mt-1 truncate text-xs text-slate-500">{{ item.wxid }}</div>
                      <Space class="mt-2" size="small" wrap>
                        <Tag :bordered="false" color="blue">公众号</Tag>
                        <Tag v-if="item.remark" :bordered="false">{{ item.remark }}</Tag>
                      </Space>
                    </div>
                  </div>
                </button>
              </Space>
              <Empty v-if="!officialAccounts.length && !officialLoading" class="pt-12" description="暂无公众号" />
            </Spin>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col bg-slate-50">
          <template v-if="currentOfficial">
            <header class="flex items-center justify-between border-b border-slate-200 bg-white px-5 py-3">
              <Space>
                <Avatar :size="44" :src="currentOfficial.avatar || undefined">
                  {{ (currentOfficial.displayName || currentOfficial.wxid).slice(0, 1) }}
                </Avatar>
                <div class="min-w-0">
                  <Typography.Title :level="5" class="!mb-0 truncate">
                    {{ officialDetail?.targetName || currentOfficial.displayName || currentOfficial.wxid }}
                  </Typography.Title>
                  <Typography.Text type="secondary">{{ currentOfficial.wxid }}</Typography.Text>
                </div>
              </Space>
              <Button type="primary" @click="jumpToChat">进入消息中心处理</Button>
            </header>

            <div class="min-h-0 flex-1 overflow-auto px-5 py-4">
              <Space class="w-full" direction="vertical" size="middle">
                <div class="rounded-xl border border-slate-200 bg-white p-4">
                  <Spin :spinning="detailLoading">
                    <Descriptions :column="2" bordered size="small">
                      <Descriptions.Item label="公众号名称">{{ officialDetail?.targetName || '-' }}</Descriptions.Item>
                      <Descriptions.Item label="公众号ID">{{ currentOfficial.wxid }}</Descriptions.Item>
                      <Descriptions.Item label="备注">{{ currentOfficial.remark || '-' }}</Descriptions.Item>
                      <Descriptions.Item label="归属账号">{{ selectedWxid }}</Descriptions.Item>
                      <Descriptions.Item :span="2" label="简介">
                        {{ currentOfficial.signature || officialDetail?.announcement || '暂无简介' }}
                      </Descriptions.Item>
                    </Descriptions>
                  </Spin>
                </div>

                <div class="rounded-xl border border-slate-200 bg-white p-4">
                  <Typography.Title :level="5">最近消息</Typography.Title>
                  <Spin :spinning="articleLoading">
                    <Table
                      :columns="articleColumns"
                      :data-source="articleRows"
                      :pagination="{ pageSize: 8 }"
                      row-key="key"
                      size="small"
                    >
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.dataIndex === 'status'">
                          <Tag :color="record.status === 'sent' ? 'success' : 'processing'">
                            {{ record.status }}
                          </Tag>
                        </template>
                        <template v-if="column.dataIndex === 'time'">
                          {{ formatTime(record.time) }}
                        </template>
                      </template>
                    </Table>
                  </Spin>
                </div>
              </Space>
            </div>
          </template>
          <Empty v-else class="pt-20" description="请选择左侧公众号" />
        </section>
      </div>
    </div>
  </Page>
</template>
