<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Input,
  List,
  Modal,
  Row,
  Select,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import {
  bootstrapAccountApi,
  createAccountAwakenLoginSessionApi,
  createAccountLoginSessionApi,
  getDashboardOverviewApi,
  logoutAccountApi,
  pollAccountLoginSessionApi,
  startAccountHeartbeatApi,
  stopAccountHeartbeatApi,
  type DashboardOverview,
  type LoginSessionState,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const accountStore = useEnterpriseAccountStore();

const accounts = computed(() => accountStore.accounts);
const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const selectedAccount = computed(() => accountStore.currentAccount);
const accountsLoading = computed(() => accountStore.loading);

const loading = ref(false);
const actionLoading = ref(false);
const overview = ref<DashboardOverview | null>(null);

const loginModalOpen = ref(false);
const loginSession = ref<LoginSessionState | null>(null);
const loginPolling = ref(false);
const loginCreating = ref(false);
const loginForm = reactive({
  deviceName: 'Enterprise Console',
  platform: 'ipad',
});

const loginChannelOptions = [
  { label: 'iPad 经典', value: 'ipad' },
  { label: '安卓 Pad（最新）', value: 'pad_latest' },
  { label: 'WinUnified 统一 PC 版（最新）', value: 'win_unified' },
  { label: 'Car', value: 'car' },
] as const;

let loginPollTimer: null | ReturnType<typeof window.setTimeout> = null;

const selectedAccountActionLabel = computed(() => {
  if (!selectedAccount.value) return '请选择账号';
  return selectedAccount.value.status === 'online' ? '在线工作中' : '离线待重新登录';
});

const qrImageSrc = computed(() => {
  const raw = String(loginSession.value?.qrBase64 || '').trim();
  if (raw.startsWith('data:image')) return raw;
  if (raw) return `data:image/png;base64,${raw}`;
  return loginSession.value?.qrUrl || '';
});

const loginStatusMeta = computed(() => {
  const status = loginSession.value?.status || 'pending';
  if (status === 'ready') return { color: 'success', label: '登录完成' };
  if (status === 'awaken_pending') return { color: 'processing', label: '已发起唤醒，请手机确认' };
  if (status === 'authenticated') return { color: 'processing', label: '初始化账号中' };
  if (status === 'scanned') return { color: 'processing', label: '已扫码，请在手机确认' };
  if (status === 'confirmed') return { color: 'processing', label: '已确认，等待账号接入' };
  if (status === 'failed' || status === 'expired') return { color: 'error', label: '登录失败/已过期' };
  return { color: 'default', label: '等待扫码' };
});

const summaryCards = computed(() => {
  const summary = overview.value?.summary;
  return [
    { label: '接入账号', value: summary?.totalAccounts ?? 0 },
    { label: '在线账号', value: summary?.onlineAccounts ?? 0 },
    { label: '24h 消息', value: summary?.messages24Hours ?? 0 },
    { label: '联系人总量', value: summary?.totalContacts ?? 0 },
    { label: '公众号总量', value: summary?.officialAccounts ?? 0 },
    { label: '群聊总量', value: summary?.groups ?? 0 },
  ];
});

function clearLoginPolling() {
  if (loginPollTimer) {
    window.clearTimeout(loginPollTimer);
    loginPollTimer = null;
  }
}

function resetLoginModalState() {
  clearLoginPolling();
  loginCreating.value = false;
  loginPolling.value = false;
  loginSession.value = null;
}

function closeLoginModal() {
  loginModalOpen.value = false;
  resetLoginModalState();
}

function scheduleLoginPolling(sessionId: string) {
  clearLoginPolling();
  loginPolling.value = true;
  loginPollTimer = window.setTimeout(() => {
    void pollLoginSession(sessionId);
  }, 1800);
}

async function loadOverview() {
  loading.value = true;
  try {
    const [overviewResult] = await Promise.all([
      getDashboardOverviewApi(),
      accountStore.refreshAccounts({ preferredWxid: selectedWxid.value }),
    ]);
    overview.value = overviewResult;
  } finally {
    loading.value = false;
  }
}

async function doAction(action: 'boot' | 'heartbeat' | 'offline' | 'stopHeartbeat') {
  const wxid = selectedAccount.value?.wxid;
  if (!wxid) return;

  actionLoading.value = true;
  try {
    if (action === 'boot') {
      await bootstrapAccountApi(wxid);
      message.success('初始化已触发');
    } else if (action === 'heartbeat') {
      await startAccountHeartbeatApi(wxid);
      message.success('心跳已启动');
    } else if (action === 'stopHeartbeat') {
      await stopAccountHeartbeatApi(wxid);
      message.success('心跳已停止');
    } else {
      await logoutAccountApi(wxid);
      message.success('账号已下线，需要重新扫码登录');
    }
    await loadOverview();
  } catch (error) {
    console.error(error);
  } finally {
    actionLoading.value = false;
  }
}

async function createLoginSession() {
  loginCreating.value = true;
  loginPolling.value = false;
  clearLoginPolling();
  try {
    loginSession.value = await createAccountLoginSessionApi({
      deviceName: loginForm.deviceName.trim() || 'Enterprise Console',
      platform: loginForm.platform.trim() || 'ipad',
    });
    if (loginSession.value?.sessionId) {
      scheduleLoginPolling(loginSession.value.sessionId);
    }
  } catch (error) {
    console.error(error);
    message.error('创建扫码登录会话失败');
  } finally {
    loginCreating.value = false;
  }
}

async function createAwakenLoginSession() {
  const wxid = selectedAccount.value?.wxid;
  if (!wxid) return;

  loginCreating.value = true;
  loginPolling.value = false;
  clearLoginPolling();
  try {
    loginSession.value = await createAccountAwakenLoginSessionApi(wxid);
    if (loginSession.value?.sessionId) {
      scheduleLoginPolling(loginSession.value.sessionId);
    }
  } catch (error) {
    console.error(error);
    message.error('发起唤醒登录失败');
  } finally {
    loginCreating.value = false;
  }
}

async function pollLoginSession(sessionId: string) {
  try {
    const result = await pollAccountLoginSessionApi(sessionId);
    loginSession.value = result;

    if (result.status === 'ready' && result.account?.wxid) {
      loginPolling.value = false;
      clearLoginPolling();
      await Promise.all([
        accountStore.refreshAccounts({ preferredWxid: result.account.wxid }),
        loadOverview(),
      ]);
      selectedWxid.value = result.account.wxid;
      message.success(`账号 ${result.account.nickname || result.account.wxid} 已接入`);
      window.setTimeout(() => {
        closeLoginModal();
      }, 800);
      return;
    }

    if (result.status === 'failed' || result.status === 'expired') {
      loginPolling.value = false;
      clearLoginPolling();
      return;
    }

    scheduleLoginPolling(sessionId);
  } catch (error) {
    console.error(error);
    scheduleLoginPolling(sessionId);
  }
}

async function openLoginModal() {
  loginModalOpen.value = true;
  resetLoginModalState();
  await createLoginSession();
}

async function openAwakenLoginModal() {
  if (!selectedAccount.value?.wxid) return;
  loginModalOpen.value = true;
  resetLoginModalState();
  await createAwakenLoginSession();
}

onMounted(() => {
  void loadOverview();
});

onBeforeUnmount(() => {
  clearLoginPolling();
});
</script>

<template>
  <Page>
    <div class="mb-4 flex flex-wrap items-center justify-end gap-3">
      <Button @click="loadOverview">刷新状态</Button>
      <Button type="primary" @click="openLoginModal">扫码登录新账号</Button>
    </div>
    <Space class="w-full" direction="vertical" size="large">
      <Card>
        <Spin :spinning="loading">
          <Row :gutter="[16, 16]">
            <Col
              v-for="item in summaryCards"
              :key="item.label"
              :lg="8"
              :md="12"
              :sm="12"
              :xl="4"
              :xs="24"
            >
              <Card size="small">
                <Statistic :title="item.label" :value="item.value" />
              </Card>
            </Col>
          </Row>
        </Spin>
      </Card>

      <Row :gutter="[16, 16]">
        <Col :lg="9" :xs="24">
          <Card title="微信账号池">
            <Spin :spinning="loading || accountsLoading">
              <List
                v-if="accounts.length"
                :data-source="accounts"
                bordered
                item-layout="horizontal"
                size="small"
              >
                <template #renderItem="{ item }">
                  <List.Item
                    :class="[
                      'cursor-pointer rounded-md px-2 transition',
                      selectedWxid === item.wxid ? 'bg-[rgb(var(--primary-1))]' : '',
                    ]"
                    @click="selectedWxid = item.wxid"
                  >
                    <List.Item.Meta>
                      <template #avatar>
                        <Avatar :src="item.avatar || undefined">
                          {{ (item.nickname || item.alias || item.wxid).slice(0, 1) }}
                        </Avatar>
                      </template>
                      <template #title>
                        <Space>
                          <Typography.Text strong>
                            {{ item.nickname || item.alias || item.wxid }}
                          </Typography.Text>
                          <Tag :color="item.status === 'online' ? 'success' : 'error'">
                            {{ item.status === 'online' ? '在线' : '离线' }}
                          </Tag>
                          <Tag v-if="item.status !== 'online'">待重新登录</Tag>
                        </Space>
                      </template>
                      <template #description>{{ item.wxid }}</template>
                    </List.Item.Meta>
                  </List.Item>
                </template>
              </List>
              <Empty v-else description="暂无接入账号，请先扫码登录">
                <Button type="primary" @click="openLoginModal">扫码登录</Button>
              </Empty>
            </Spin>
          </Card>
        </Col>

        <Col :lg="15" :xs="24">
          <Card title="当前账号工作台">
            <template v-if="selectedAccount">
              <Space class="w-full" direction="vertical" size="middle">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <Space align="start">
                    <Avatar :size="68" :src="selectedAccount.avatar || undefined">
                      {{ (selectedAccount.nickname || selectedAccount.alias || selectedAccount.wxid).slice(0, 1) }}
                    </Avatar>
                    <Space direction="vertical" size="small">
                      <Typography.Title :level="4" class="!mb-0">
                        {{ selectedAccount.nickname || selectedAccount.alias || selectedAccount.wxid }}
                      </Typography.Title>
                      <Typography.Text type="secondary">
                        {{ selectedAccount.wxid }}
                      </Typography.Text>
                      <Typography.Text>{{ selectedAccount.signature || '暂无签名' }}</Typography.Text>
                    </Space>
                  </Space>

                  <Space wrap>
                    <Tag :color="selectedAccount.status === 'online' ? 'success' : 'error'">
                      {{ selectedAccountActionLabel }}
                    </Tag>
                    <Button :loading="actionLoading" @click="doAction('boot')">初始化</Button>
                    <Button :loading="actionLoading" type="primary" @click="doAction('heartbeat')">
                      开启心跳
                    </Button>
                    <Button :loading="actionLoading" @click="doAction('stopHeartbeat')">停止心跳</Button>
                    <Button
                      v-if="selectedAccount.status !== 'online'"
                      :loading="loginCreating"
                      type="primary"
                      @click="openAwakenLoginModal"
                    >
                      唤醒登录
                    </Button>
                    <Button
                      v-else
                      :loading="actionLoading"
                      @click="openLoginModal"
                    >
                      重新扫码登录
                    </Button>
                    <Button :loading="actionLoading" danger @click="doAction('offline')">下线</Button>
                  </Space>
                </div>

                <Descriptions :column="2" bordered size="small">
                  <Descriptions.Item label="微信号">{{ selectedAccount.alias || '-' }}</Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag :color="selectedAccount.status === 'online' ? 'success' : 'error'">
                      {{ selectedAccount.status === 'online' ? '在线' : '离线' }}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="手机号">{{ selectedAccount.mobile || '-' }}</Descriptions.Item>
                  <Descriptions.Item label="设备">{{ selectedAccount.deviceName || '-' }}</Descriptions.Item>
                  <Descriptions.Item label="最近心跳">{{ selectedAccount.lastHeartbeatAt || '-' }}</Descriptions.Item>
                  <Descriptions.Item label="最近同步">{{ selectedAccount.lastSyncAt || '-' }}</Descriptions.Item>
                </Descriptions>
              </Space>
            </template>
            <Empty v-else description="请先选择账号或扫码接入" />
          </Card>
        </Col>
      </Row>

      <Card title="账号资产概览">
        <Spin :spinning="loading">
          <Table
            :columns="[
              { dataIndex: 'nickname', title: '账号' },
              { dataIndex: 'status', title: '状态' },
              { dataIndex: 'directContactCount', title: '联系人' },
              { dataIndex: 'officialAccountCount', title: '公众号' },
              { dataIndex: 'groupCount', title: '群聊' },
              { dataIndex: 'messageCount', title: '消息沉淀' },
              { dataIndex: 'lastSyncAt', title: '最近同步' },
            ]"
            :data-source="overview?.accounts ?? []"
            :pagination="{ pageSize: 8 }"
            row-key="wxid"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'status'">
                <Tag :color="record.status === 'online' ? 'success' : 'error'">
                  {{ record.status === 'online' ? '在线' : '离线' }}
                </Tag>
              </template>
              <template v-if="column.dataIndex === 'nickname'">
                {{ record.nickname || record.wxid }}
              </template>
            </template>
          </Table>
        </Spin>
      </Card>
    </Space>

    <Modal
      v-model:open="loginModalOpen"
      :footer="null"
      destroy-on-close
      title="扫码登录微信账号"
      width="720px"
      @cancel="closeLoginModal"
    >
      <div class="grid gap-6 lg:grid-cols-[280px_minmax(0,1fr)]">
        <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <div class="aspect-square overflow-hidden rounded-xl border border-slate-200 bg-white">
            <img
              v-if="qrImageSrc"
              :src="qrImageSrc"
              alt="微信登录二维码"
              class="h-full w-full object-contain"
            />
            <div v-else class="flex h-full items-center justify-center text-sm text-slate-400">
              正在生成二维码
            </div>
          </div>
          <Space class="mt-4" direction="vertical" size="small">
            <Tag :color="loginStatusMeta.color">{{ loginStatusMeta.label }}</Tag>
            <Typography.Text type="secondary">
              {{
                loginSession?.status === 'awaken_pending'
                  ? '已向该账号发送唤醒登录请求，请在手机端确认。'
                  : loginSession?.message || '请使用微信扫描二维码，并在手机上确认登录。'
              }}
            </Typography.Text>
            <Typography.Text v-if="loginSession?.expiresAt" type="secondary">
              二维码有效期至：{{ loginSession.expiresAt }}
            </Typography.Text>
          </Space>
        </div>

        <Space class="w-full" direction="vertical" size="middle">
          <div class="rounded-2xl border border-slate-200 bg-white p-4">
            <Typography.Title :level="5">登录参数</Typography.Title>
            <Space class="w-full" direction="vertical" size="middle">
              <div>
                <Typography.Text type="secondary">平台</Typography.Text>
                <Select
                  v-model:value="loginForm.platform"
                  :options="loginChannelOptions"
                  class="w-full"
                  placeholder="请选择登录渠道"
                />
              </div>
              <div>
                <Typography.Text type="secondary">设备名称</Typography.Text>
                <Input v-model:value="loginForm.deviceName" placeholder="后台显示的设备名" />
              </div>
            </Space>
          </div>

          <div class="rounded-2xl border border-slate-200 bg-white p-4">
            <Typography.Title :level="5">接入流程</Typography.Title>
            <Space direction="vertical" size="small">
              <Typography.Text v-if="loginSession?.status === 'awaken_pending'">
                1. 已对离线账号发起唤醒登录
              </Typography.Text>
              <Typography.Text v-if="loginSession?.status === 'awaken_pending'">
                2. 轮询 UUID 登录状态，确认后自动恢复账号在线
              </Typography.Text>
              <template v-else>
                <Typography.Text>1. 创建登录会话并生成二维码</Typography.Text>
                <Typography.Text>2. 手机微信扫码确认</Typography.Text>
                <Typography.Text>3. 后端自动初始化、开启心跳并写入账号池</Typography.Text>
              </template>
            </Space>
          </div>

          <Space wrap>
            <Button :loading="loginCreating" type="primary" @click="createLoginSession">
              重新生成二维码
            </Button>
            <Button :loading="loading || accountsLoading" @click="loadOverview">刷新账号池</Button>
          </Space>
        </Space>
      </div>
    </Modal>
  </Page>
</template>
