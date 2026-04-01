<script lang="ts" setup>
import { computed, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Card,
  Empty,
  Form,
  Input,
  Select,
  Space,
  Switch,
  Typography,
} from 'ant-design-vue';

type SectionItem = {
  id: string;
  subtitle: string;
  title: string;
};

const sections: SectionItem[] = [
  { id: 'base', subtitle: '管理员资料维护', title: '基础资料' },
  { id: 'security', subtitle: '登录与设备安全', title: '安全设置' },
  { id: 'notify', subtitle: '消息推送规则', title: '通知偏好' },
];

const selectedSection = ref(sections[0]?.id ?? '');

const formState = ref({
  email: 'admin@enterprise.local',
  language: 'zh-CN',
  mobile: '15170086809',
  nickname: '系统管理员',
  receiveAlarm: true,
  timezone: 'Asia/Shanghai',
});

const current = computed(() => sections.find((item) => item.id === selectedSection.value));
</script>

<template>
  <Page title="个人信息管理">
    <Card class="h-[calc(100vh-220px)] min-h-[640px]">
      <div class="flex h-full gap-4">
        <aside class="flex h-full w-[300px] min-w-[280px] flex-col border-r pr-4">
          <div class="rounded-lg bg-slate-50 p-4">
            <Space align="start">
              <Avatar :size="48">系</Avatar>
              <div>
                <div class="text-sm font-semibold text-slate-900">系统管理员</div>
                <div class="text-xs text-slate-500">admin</div>
              </div>
            </Space>
          </div>
          <div class="mt-3 min-h-0 flex-1 overflow-auto">
            <Space class="w-full" direction="vertical" size="small">
              <button
                v-for="item in sections"
                :key="item.id"
                :class="[
                  'w-full rounded-lg border p-3 text-left transition',
                  selectedSection === item.id
                    ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))]'
                    : 'border-transparent bg-gray-50 hover:border-gray-200',
                ]"
                type="button"
                @click="selectedSection = item.id"
              >
                <div class="text-sm font-medium text-slate-900">{{ item.title }}</div>
                <div class="mt-1 text-xs text-slate-500">{{ item.subtitle }}</div>
              </button>
            </Space>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col">
          <template v-if="current">
            <header class="border-b pb-3">
              <Typography.Title :level="5" class="!mb-0">
                {{ current.title }}
              </Typography.Title>
              <Typography.Text type="secondary">
                {{ current.subtitle }}
              </Typography.Text>
            </header>

            <div class="min-h-0 flex-1 overflow-auto pt-4">
              <Card size="small" title="配置项（UI占位）">
                <Form layout="vertical">
                  <Form.Item label="管理员昵称">
                    <Input v-model:value="formState.nickname" />
                  </Form.Item>
                  <Form.Item label="手机号">
                    <Input v-model:value="formState.mobile" />
                  </Form.Item>
                  <Form.Item label="邮箱">
                    <Input v-model:value="formState.email" />
                  </Form.Item>
                  <Form.Item label="语言">
                    <Select
                      v-model:value="formState.language"
                      :options="[
                        { label: '简体中文', value: 'zh-CN' },
                        { label: 'English', value: 'en-US' },
                      ]"
                    />
                  </Form.Item>
                  <Form.Item label="时区">
                    <Select
                      v-model:value="formState.timezone"
                      :options="[
                        { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
                        { label: 'UTC', value: 'UTC' },
                      ]"
                    />
                  </Form.Item>
                  <Form.Item label="告警通知">
                    <Switch v-model:checked="formState.receiveAlarm" />
                  </Form.Item>
                </Form>
                <div class="flex justify-end gap-2">
                  <Button>重置</Button>
                  <Button type="primary">保存配置</Button>
                </div>
              </Card>
            </div>
          </template>
          <Empty v-else description="请选择左侧设置项" />
        </section>
      </div>
    </Card>
  </Page>
</template>
