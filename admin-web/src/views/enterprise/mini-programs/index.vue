<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Button, Select, Tag, Typography } from 'ant-design-vue';

import { useEnterpriseAccountStore } from '#/store';

type MiniProgramInterface = {
  method: 'POST';
  name: string;
  note: string;
  path: string;
  params: string[];
  status: '待接入';
};

type MiniProgramCategory = {
  description: string;
  key: string;
  title: string;
  items: MiniProgramInterface[];
};

const categories: MiniProgramCategory[] = [
  {
    key: 'auth',
    title: '授权与登录',
    description: '处理小程序授权、网页登录和支付场景的 session / 二维码能力。',
    items: [
      {
        name: '授权小程序',
        path: '/api/Wxapp/JSLogin',
        method: 'POST',
        note: '返回授权后的 code，适合登录或用户授权起点。',
        params: ['Wxid', 'Appid'],
        status: '待接入',
      },
      {
        name: '获取支付 SessionId',
        path: '/api/Wxapp/JSGetSessionid',
        method: 'POST',
        note: '获取小程序支付所需 sessionid，适合支付前链路准备。',
        params: ['Wxid', 'Appid'],
        status: '待接入',
      },
      {
        name: '获取付款二维码',
        path: '/api/Wxapp/JSGetSessionidQRcode',
        method: 'POST',
        note: '结合支付 session 获取付款二维码。',
        params: ['Wxid', 'Appid', 'Sessionid'],
        status: '待接入',
      },
      {
        name: '扫码授权登录',
        path: '/api/Wxapp/QrcodeAuthLogin',
        method: 'POST',
        note: '为网页或 App 场景提供二维码授权登录能力。',
        params: ['Wxid', 'Appid', 'Uuid/Code'],
        status: '待接入',
      },
    ],
  },
  {
    key: 'data',
    title: '数据与云函数',
    description: '执行小程序端操作、云函数和记录类接口。',
    items: [
      {
        name: '小程序操作',
        path: '/api/Wxapp/JSOperateWxData',
        method: 'POST',
        note: '统一封装小程序数据操作能力，适合开放给场景动作调用。',
        params: ['Wxid', 'Appid', 'OperateData'],
        status: '待接入',
      },
      {
        name: '云函数调用',
        path: '/api/Wxapp/CloudCallFunction',
        method: 'POST',
        note: '执行小程序云函数，适合业务编排或后台联动。',
        params: ['Wxid', 'Appid', 'FuncName', 'Data'],
        status: '待接入',
      },
      {
        name: '新增小程序记录',
        path: '/api/Wxapp/AddWxAppRecord',
        method: 'POST',
        note: '用于记录小程序行为或业务轨迹。',
        params: ['Wxid', 'Appid', 'Path/Scene'],
        status: '待接入',
      },
    ],
  },
  {
    key: 'mobile',
    title: '手机号与身份',
    description: '围绕手机号、OpenID 和插件校验的账号身份能力。',
    items: [
      {
        name: '绑定增加手机号',
        path: '/api/Wxapp/AddMobile',
        method: 'POST',
        note: '支持发送验证码和提交验证码两段式绑定。',
        params: ['Wxid', 'Appid', 'Mobile', 'VerifyCode'],
        status: '待接入',
      },
      {
        name: '删除手机号',
        path: '/api/Wxapp/DelMobile',
        method: 'POST',
        note: '删除已绑定手机号。',
        params: ['Wxid', 'Appid', 'Mobile'],
        status: '待接入',
      },
      {
        name: '获取用户 OpenID',
        path: '/api/Wxapp/GetUserOpenId',
        method: 'POST',
        note: '根据小程序场景获取用户 OpenID/OpenData。',
        params: ['Wxid', 'Appid', 'ToUserName/Code'],
        status: '待接入',
      },
      {
        name: '获取全部手机号',
        path: '/api/Wxapp/GetAllMobile',
        method: 'POST',
        note: '读取当前小程序可用手机号集合。',
        params: ['Wxid', 'Appid'],
        status: '待接入',
      },
      {
        name: '获取 HostSign',
        path: '/api/Wxapp/Verifyplugin',
        method: 'POST',
        note: '小程序插件 / HostSign 校验能力。',
        params: ['Wxid', 'Appid'],
        status: '待接入',
      },
    ],
  },
  {
    key: 'asset',
    title: '头像与素材',
    description: '处理小程序头像生成、上传与写入。',
    items: [
      {
        name: '获取随机头像',
        path: '/api/Wxapp/GetRandomAvatar',
        method: 'POST',
        note: '生成或分配随机头像素材。',
        params: ['Wxid', 'Appid'],
        status: '待接入',
      },
      {
        name: '上传头像图片',
        path: '/api/Wxapp/UploadAvatarImg',
        method: 'POST',
        note: '上传头像素材图片，为后续设置头像做准备。',
        params: ['Wxid', 'Appid', 'ImgUrl/FilePath'],
        status: '待接入',
      },
      {
        name: '写入头像',
        path: '/api/Wxapp/AddAvatar',
        method: 'POST',
        note: '将上传后的头像正式写入小程序资料。',
        params: ['Wxid', 'Appid', 'AvatarId'],
        status: '待接入',
      },
    ],
  },
  {
    key: 'payment',
    title: '支付能力',
    description: '当前已暴露的支付场景接口，后续可与小程序支付流程联动。',
    items: [
      {
        name: '云闪付支付',
        path: '/api/Wxapp/GetUnionPay',
        method: 'POST',
        note: '处理云闪付支付请求参数与支付提交。',
        params: ['Wxid', 'Appid', 'PayInfo'],
        status: '待接入',
      },
    ],
  },
];

const accountStore = useEnterpriseAccountStore();

const loadingAccounts = ref(false);
const activeCategoryKey = ref(categories[0]?.key ?? '');

const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const accountOptions = computed(() => accountStore.accountOptions);
const currentCategory = computed(
  () => categories.find((item) => item.key === activeCategoryKey.value) ?? categories[0],
);

async function loadAccounts() {
  loadingAccounts.value = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loadingAccounts.value = false;
  }
}

onMounted(async () => {
  await loadAccounts();
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
      <Button size="small" type="primary">
        接口展示模式
      </Button>
    </div>

    <div class="min-h-[calc(100vh-210px)] overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <div class="grid h-full min-h-0 xl:grid-cols-[340px_minmax(0,1fr)]">
        <aside class="border-r border-slate-200 bg-slate-50/70 p-4">
          <div class="rounded-2xl border border-slate-200 bg-white p-5">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-[34px] font-semibold tracking-tight text-slate-900">小程序操作</div>
                <div class="mt-2 text-sm leading-6 text-slate-500">
                  先做接口展示，不触发真实调用。当前整理自 `Wxapp` 控制器。
                </div>
              </div>
              <Tag :bordered="false" color="blue">{{ categories.length }} 组</Tag>
            </div>
          </div>

          <div class="mt-4 space-y-3">
            <button
              v-for="category in categories"
              :key="category.key"
              class="w-full rounded-2xl border px-4 py-4 text-left transition"
              :class="
                category.key === activeCategoryKey
                  ? 'border-slate-900 bg-white shadow-sm'
                  : 'border-slate-200 bg-white/70 hover:border-slate-300 hover:bg-white'
              "
              type="button"
              @click="activeCategoryKey = category.key"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="text-lg font-semibold text-slate-900">{{ category.title }}</div>
                  <div class="mt-1 text-sm leading-6 text-slate-500">
                    {{ category.description }}
                  </div>
                </div>
                <Tag :bordered="false" color="processing">{{ category.items.length }}</Tag>
              </div>
            </button>
          </div>
        </aside>

        <section class="min-h-0 overflow-y-auto p-5">
          <div class="rounded-2xl border border-slate-200 bg-slate-50/60 p-5">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="text-[30px] font-semibold tracking-tight text-slate-900">
                  {{ currentCategory.title }}
                </div>
                <div class="mt-2 max-w-3xl text-sm leading-7 text-slate-500">
                  {{ currentCategory.description }}
                </div>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <Tag :bordered="false" color="purple">Wxapp</Tag>
                <Tag :bordered="false" color="gold">展示中</Tag>
                <Tag :bordered="false" color="cyan">{{ currentCategory.items.length }} 个接口</Tag>
              </div>
            </div>
          </div>

          <div class="mt-5 space-y-4">
            <article
              v-for="item in currentCategory.items"
              :key="item.path"
              class="rounded-2xl border border-slate-200 bg-white p-5 shadow-[0_10px_30px_rgba(15,23,42,0.04)]"
            >
              <div class="flex flex-wrap items-start justify-between gap-4">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <div class="text-xl font-semibold text-slate-900">{{ item.name }}</div>
                    <Tag :bordered="false" color="green">{{ item.method }}</Tag>
                    <Tag :bordered="false" color="default">{{ item.status }}</Tag>
                  </div>
                  <div class="mt-2 break-all font-mono text-sm text-slate-500">{{ item.path }}</div>
                  <div class="mt-3 text-[15px] leading-7 text-slate-600">{{ item.note }}</div>
                </div>
              </div>

              <div class="mt-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4">
                <div class="text-sm font-medium text-slate-900">推荐参数</div>
                <div class="mt-3 flex flex-wrap gap-2">
                  <Tag v-for="param in item.params" :key="param" :bordered="false" color="blue">
                    {{ param }}
                  </Tag>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </Page>
</template>
