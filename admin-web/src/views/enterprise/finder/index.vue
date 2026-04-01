<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Empty,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import { getFinderProfileApi, type FinderProfileResult } from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const accountStore = useEnterpriseAccountStore();

const loadingAccounts = ref(false);
const loadingProfile = ref(false);
const finderProfile = ref<FinderProfileResult | null>(null);

const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const accountOptions = computed(() => accountStore.accountOptions);

const currentAccount = computed(() =>
  accountStore.accounts.find((item) => item.wxid === selectedWxid.value) ?? null,
);

const profileDisplayName = computed(() =>
  finderProfile.value?.nickname || currentAccount.value?.nickname || currentAccount.value?.wxid || '未绑定视频号',
);

const metaItems = computed(() => {
  const profile = finderProfile.value;
  if (!profile) return [];
  return [
    { label: '视频号ID', value: profile.username || '-' },
    { label: '用户标识', value: String(profile.userFlag || 0) },
    { label: '关注状态', value: profile.followFlag ? '已关注' : '未关注' },
    { label: '资料序列', value: String(profile.seq || 0) },
    {
      label: '所在地',
      value: [profile.extInfo.country, profile.extInfo.province, profile.extInfo.city].filter(Boolean).join(' / ') || '-',
    },
    { label: '性别', value: sexLabel(profile.extInfo.sex) },
  ];
});

function sexLabel(sex: number) {
  if (sex === 1) return '男';
  if (sex === 2) return '女';
  return '未知';
}

async function loadAccounts() {
  loadingAccounts.value = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loadingAccounts.value = false;
  }
}

async function loadFinderProfile() {
  if (!selectedWxid.value) {
    finderProfile.value = null;
    return;
  }
  loadingProfile.value = true;
  try {
    finderProfile.value = await getFinderProfileApi(selectedWxid.value);
  } catch (error: any) {
    finderProfile.value = null;
    message.error(error?.message || '加载视频号资料失败');
  } finally {
    loadingProfile.value = false;
  }
}

watch(selectedWxid, async () => {
  await loadFinderProfile();
});

onMounted(async () => {
  await loadAccounts();
  await loadFinderProfile();
});
</script>

<template>
  <Page>
    <div class="mb-4 flex flex-wrap items-center gap-3">
      <Typography.Text type="secondary">当前账号</Typography.Text>
      <Select
        v-model:value="selectedWxid"
        :loading="loadingAccounts"
        :options="accountOptions"
        class="!w-[320px]"
        placeholder="请选择账号"
      />
      <Button @click="loadFinderProfile" :loading="loadingProfile">刷新资料</Button>
    </div>

    <div class="min-h-[calc(100vh-210px)] overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <Spin :spinning="loadingProfile">
        <template v-if="finderProfile">
          <div class="border-b border-slate-200 px-6 py-6">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="flex min-w-0 items-start gap-4">
                <Avatar :size="72" :src="finderProfile.headUrl || currentAccount?.avatar || undefined" class="shrink-0">
                  {{ profileDisplayName.slice(0, 1) }}
                </Avatar>
                <div class="min-w-0">
                  <div class="truncate text-[32px] font-semibold tracking-tight text-slate-900">
                    {{ profileDisplayName }}
                  </div>
                  <div class="mt-1 truncate text-lg text-slate-400">
                    {{ finderProfile.username || '当前账号暂未返回视频号标识' }}
                  </div>
                  <div class="mt-3 flex flex-wrap items-center gap-2">
                    <Tag :bordered="false" color="purple">视频号</Tag>
                    <Tag :bordered="false" :color="finderProfile.canUse ? 'green' : 'volcano'">
                      {{ finderProfile.canUse ? '可用' : '未准备完成' }}
                    </Tag>
                    <Tag
                      v-if="finderProfile.authInfo?.authIconType"
                      :bordered="false"
                      color="blue"
                    >
                      已认证
                    </Tag>
                    <Tag
                      v-if="finderProfile.originalFlag"
                      :bordered="false"
                      color="gold"
                    >
                      原创标记
                    </Tag>
                  </div>
                </div>
              </div>
              <div class="max-w-[420px] rounded-2xl border border-slate-200 bg-slate-50/80 px-4 py-3 text-sm leading-6 text-slate-600">
                {{ finderProfile.message || '已同步当前账号视频号资料。' }}
              </div>
            </div>

            <div class="mt-5 grid gap-3 rounded-2xl border border-slate-200 bg-slate-50/70 p-4 xl:grid-cols-3">
              <div
                v-for="item in metaItems"
                :key="item.label"
                class="rounded-2xl border border-slate-200 bg-white px-4 py-3"
              >
                <div class="text-sm text-slate-400">{{ item.label }}</div>
                <div class="mt-2 break-all text-lg font-semibold text-slate-900">{{ item.value }}</div>
              </div>
            </div>
          </div>

          <div class="grid gap-4 p-6 xl:grid-cols-[minmax(0,1.2fr)_420px]">
            <div class="space-y-4">
              <div class="rounded-2xl border border-slate-200 bg-white p-5">
                <div class="text-xl font-semibold text-slate-900">账号资料</div>
                <div class="mt-4 grid gap-3 md:grid-cols-2">
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div class="text-sm text-slate-400">签名</div>
                    <div class="mt-2 whitespace-pre-wrap break-words text-[15px] leading-7 text-slate-700">
                      {{ finderProfile.signature || '暂无签名' }}
                    </div>
                  </div>
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div class="text-sm text-slate-400">封面地址</div>
                    <div class="mt-2 break-all text-sm leading-6 text-slate-700">
                      {{ finderProfile.coverImgUrl || '暂无封面地址' }}
                    </div>
                  </div>
                </div>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-white p-5">
                <div class="text-xl font-semibold text-slate-900">认证与昵称校验</div>
                <div v-if="finderProfile.verifyInfo || finderProfile.authInfo" class="mt-4 grid gap-3 md:grid-cols-2">
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div class="text-sm text-slate-400">昵称校验</div>
                    <div class="mt-2 space-y-2 text-sm leading-6 text-slate-700">
                      <div><span class="text-slate-400">校验昵称：</span>{{ finderProfile.verifyInfo?.verifyNickname || '-' }}</div>
                      <div><span class="text-slate-400">前缀：</span>{{ finderProfile.verifyInfo?.verifyPrefix || '-' }}</div>
                      <div><span class="text-slate-400">说明：</span>{{ finderProfile.verifyInfo?.bannerWording || finderProfile.nicknameModifyWording || '-' }}</div>
                    </div>
                  </div>
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div class="text-sm text-slate-400">认证信息</div>
                    <div class="mt-2 space-y-2 text-sm leading-6 text-slate-700">
                      <div><span class="text-slate-400">实名：</span>{{ finderProfile.authInfo?.realName || '-' }}</div>
                      <div><span class="text-slate-400">职业：</span>{{ finderProfile.authInfo?.authProfession || '-' }}</div>
                      <div><span class="text-slate-400">应用名：</span>{{ finderProfile.authInfo?.appName || '-' }}</div>
                      <div><span class="text-slate-400">担保方：</span>{{ finderProfile.authInfo?.guarantorName || '-' }}</div>
                    </div>
                  </div>
                </div>
                <Empty v-else class="mt-6" description="暂无认证与校验信息" />
              </div>
            </div>

            <div class="space-y-4">
              <div class="rounded-2xl border border-slate-200 bg-white p-5">
                <div class="text-xl font-semibold text-slate-900">限制与状态</div>
                <div class="mt-4 space-y-3 text-sm leading-6 text-slate-700">
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div class="font-medium text-slate-900">跨地域状态</div>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <Tag :bordered="false" :color="finderProfile.isNonresidentRealtimeLocation ? 'volcano' : 'green'">
                        实时定位{{ finderProfile.isNonresidentRealtimeLocation ? '异常' : '正常' }}
                      </Tag>
                      <Tag :bordered="false" :color="finderProfile.isNonresidentWxacctLocation ? 'volcano' : 'green'">
                        微信属地{{ finderProfile.isNonresidentWxacctLocation ? '异常' : '正常' }}
                      </Tag>
                      <Tag :bordered="false" :color="finderProfile.isNonresidentFinderacctLocation ? 'volcano' : 'green'">
                        视频号属地{{ finderProfile.isNonresidentFinderacctLocation ? '异常' : '正常' }}
                      </Tag>
                    </div>
                  </div>
                  <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3">
                    <div><span class="text-slate-400">ActionType：</span>{{ finderProfile.actionType }}</div>
                    <div><span class="text-slate-400">SpamStatus：</span>{{ finderProfile.spamStatus }}</div>
                    <div><span class="text-slate-400">OriginalFlag：</span>{{ finderProfile.originalFlag }}</div>
                  </div>
                </div>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-white p-5">
                <div class="text-xl font-semibold text-slate-900">使用说明</div>
                <div class="mt-4 space-y-2 text-sm leading-7 text-slate-600">
                  <div>1. 当前模块先对接已验证的 `Finder/UserPrepare`，用于查看视频号准备状态和基础资料。</div>
                  <div>2. 若后续补齐更多 Finder 接口，这里可继续扩展成视频号内容与互动工作台。</div>
                  <div>3. 资料异常时优先点击“刷新资料”，确认当前账号登录态仍然有效。</div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <div v-else class="flex h-[calc(100vh-280px)] items-center justify-center">
          <Empty description="当前账号暂无可用视频号资料" />
        </div>
      </Spin>
    </div>
  </Page>
</template>
