<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Empty,
  Image,
  Input,
  Modal,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import {
  getMomentDetailApi,
  listMomentsApi,
  operateMomentApi,
  publishMomentApi,
  type MomentSummary,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const accountStore = useEnterpriseAccountStore();

const loading = reactive({
  accounts: false,
  detail: false,
  list: false,
  loadMore: false,
  publish: false,
  operate: false,
});

const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const accountOptions = computed(() => accountStore.accountOptions);

const listKeyword = ref('');
const momentListPaneRef = ref<HTMLElement | null>(null);
const moments = ref<MomentSummary[]>([]);
const selectedMomentId = ref('');
const firstPageMd5 = ref('');
const nextMaxId = ref('');
const hasMore = ref(false);
const listLoaded = ref(false);
const momentsBootstrapping = ref(true);
const latestListRequestId = ref(0);
const mediaFallbackIndex = ref<Record<string, number>>({});

const detailCache = ref<Record<string, MomentSummary>>({});

const publishModal = reactive({
  blackList: '',
  content: '',
  open: false,
  withUserList: '',
});

const filteredMoments = computed(() => {
  const keyword = listKeyword.value.trim().toLowerCase();
  if (!keyword) return moments.value;
  return moments.value.filter((item) =>
    [
      item.authorName,
      item.authorWxid,
      item.preview,
      item.content,
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword)),
  );
});

const currentMoment = computed(() => {
  const cached = selectedMomentId.value ? detailCache.value[selectedMomentId.value] : null;
  const moment = cached ?? moments.value.find((item) => item.id === selectedMomentId.value) ?? null;
  if (!moment) return null;
  return {
    ...moment,
    comments: Array.isArray(moment.comments) ? moment.comments : [],
    likes: Array.isArray(moment.likes) ? moment.likes : [],
    media: Array.isArray(moment.media) ? moment.media : [],
  };
});

const currentMetaItems = computed(() => {
  if (!currentMoment.value) return [];
  return [
    { label: '发布时间', value: formatTime(currentMoment.value.createdAt) },
    { label: '媒体数', value: String(currentMoment.value.mediaCount) },
    { label: '点赞', value: String(currentMoment.value.likeCount) },
    { label: '评论', value: String(currentMoment.value.commentCount) },
    { label: '可见范围', value: currentMoment.value.visibility === 'private' ? '仅自己可见' : '公开' },
  ];
});

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

function mediaTypeLabel(type: string) {
  if (type === 'video') return '视频';
  if (type === 'image') return '图片';
  return '文本';
}

function mediaTagColor(type: string) {
  if (type === 'video') return 'purple';
  if (type === 'image') return 'green';
  return 'default';
}

function upsertMoments(items: MomentSummary[], replace = false) {
  const merged = new Map<string, MomentSummary>();
  if (!replace) {
    for (const item of moments.value) merged.set(item.id, item);
  }
  for (const item of items) {
    merged.set(item.id, {
      ...(merged.get(item.id) ?? {}),
      ...item,
    });
  }
  moments.value = Array.from(merged.values()).sort((a, b) => b.createdAt - a.createdAt);
  if (!selectedMomentId.value || !moments.value.some((item) => item.id === selectedMomentId.value)) {
    selectedMomentId.value = moments.value[0]?.id ?? '';
  }
}

function resetMomentWorkspace() {
  moments.value = [];
  detailCache.value = {};
  selectedMomentId.value = '';
  firstPageMd5.value = '';
  nextMaxId.value = '';
  hasMore.value = false;
  mediaFallbackIndex.value = {};
}

async function loadAccounts() {
  loading.accounts = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loading.accounts = false;
  }
}

function normalizeMomentListPayload(result: any) {
  const payload = result && typeof result === 'object' && Array.isArray(result.items)
    ? result
    : (result?.data ?? result ?? {});
  return {
    firstPageMd5: payload?.firstPageMd5 || '',
    hasMore: !!payload?.hasMore,
    items: Array.isArray(payload?.items) ? payload.items : [],
    nextMaxId: payload?.nextMaxId || '',
  };
}

function isCurrentListRequest(requestId: number, wxid: string) {
  return latestListRequestId.value === requestId && selectedWxid.value === wxid;
}

async function loadMoments(reset = false) {
  if (!selectedWxid.value) return;
  const requestWxid = selectedWxid.value;
  const requestId = latestListRequestId.value + 1;
  latestListRequestId.value = requestId;
  if (reset) {
    loading.list = true;
    listLoaded.value = false;
  } else {
    loading.loadMore = true;
  }
  try {
    const rawResult = await listMomentsApi(requestWxid, {
      firstPageMd5: reset ? undefined : firstPageMd5.value || undefined,
      maxId: reset ? undefined : nextMaxId.value || undefined,
    });
    if (!isCurrentListRequest(requestId, requestWxid)) return;
    const result = normalizeMomentListPayload(rawResult);
    const items = result.items;
    if (reset) {
      resetMomentWorkspace();
    }
    firstPageMd5.value = result.firstPageMd5 || firstPageMd5.value;
    nextMaxId.value = result.nextMaxId || '';
    hasMore.value = !!result.hasMore && !!result.nextMaxId;
    upsertMoments(items, reset);
  } catch (error: any) {
    if (!isCurrentListRequest(requestId, requestWxid)) return;
    if (reset) {
      resetMomentWorkspace();
    }
    message.error(error?.message || '加载朋友圈列表失败');
  } finally {
    if (!isCurrentListRequest(requestId, requestWxid)) return;
    loading.list = false;
    loading.loadMore = false;
    if (reset) {
      listLoaded.value = true;
    }
  }
}

async function loadMomentDetail(moment = currentMoment.value) {
  if (!selectedWxid.value || !moment?.id) return;
  loading.detail = true;
  try {
    const result = await getMomentDetailApi(selectedWxid.value, moment.id, moment.authorWxid);
    detailCache.value = {
      ...detailCache.value,
      [moment.id]: result.item,
    };
    upsertMoments([result.item], false);
  } catch (error: any) {
    message.warning(error?.message || '朋友圈详情暂不可用，已显示列表摘要');
  } finally {
    loading.detail = false;
  }
}

async function activateMomentsWorkspace() {
  await loadAccounts();
  if (!selectedWxid.value) {
    resetMomentWorkspace();
    return;
  }
  await loadMoments(true);
}

function openPublishModal() {
  publishModal.content = '';
  publishModal.blackList = '';
  publishModal.withUserList = '';
  publishModal.open = true;
}

function parseCSV(raw: string) {
  return raw
    .split(/[\n,，;\s]+/g)
    .map((item) => item.trim())
    .filter(Boolean);
}

async function submitPublishMoment() {
  if (!selectedWxid.value) return;
  loading.publish = true;
  try {
    const result = await publishMomentApi(selectedWxid.value, {
      blackList: parseCSV(publishModal.blackList),
      content: publishModal.content.trim(),
      withUserList: parseCSV(publishModal.withUserList),
    });
    message.success(result.message || '朋友圈发布请求已提交');
    publishModal.open = false;
    await loadMoments(true);
  } catch (error: any) {
    message.error(error?.message || '发布朋友圈失败');
  } finally {
    loading.publish = false;
  }
}

async function operateMoment(action: 'delete' | 'private' | 'public') {
  if (!selectedWxid.value || !currentMoment.value?.id) return;
  const labels: Record<string, string> = {
    delete: '删除这条朋友圈',
    private: '将这条朋友圈设为仅自己可见',
    public: '将这条朋友圈设为公开',
  };
  Modal.confirm({
    centered: true,
    content: labels[action],
    okText: '确认',
    title: '确认操作',
    async onOk() {
      loading.operate = true;
      try {
        const result = await operateMomentApi(selectedWxid.value!, currentMoment.value!.id, {
          action,
        });
        message.success(result.message || '操作成功');
        if (action === 'delete') {
          delete detailCache.value[currentMoment.value!.id];
          moments.value = moments.value.filter((item) => item.id !== currentMoment.value!.id);
          selectedMomentId.value = moments.value[0]?.id ?? '';
          return;
        }
        await loadMomentDetail(currentMoment.value);
      } finally {
        loading.operate = false;
      }
    },
  });
}

watch(selectedWxid, async () => {
  resetMomentWorkspace();
  listLoaded.value = false;
  if (!selectedWxid.value) return;
  if (momentsBootstrapping.value) return;
  await loadMoments(true);
});

watch(selectedMomentId, async () => {
  mediaFallbackIndex.value = {};
  const moment = moments.value.find((item) => item.id === selectedMomentId.value);
  if (!moment || detailCache.value[moment.id]) return;
  await loadMomentDetail(moment);
});

function getMomentMediaKey(index: number) {
  return `${selectedMomentId.value || 'moment'}-${index}`;
}

function normalizeMomentMediaSource(raw: string) {
  const value = String(raw || '').trim();
  if (!value) return '';
  if (value.startsWith('http://') || value.startsWith('https://') || value.startsWith('data:')) {
    return value;
  }
  return '';
}

function getMomentMediaCandidates(media: MomentSummary['media'][number], mediaType: string) {
  const imageFirst = [
    normalizeMomentMediaSource(media.url),
    normalizeMomentMediaSource(media.thumbUrl),
  ];
  const videoFirst = [
    normalizeMomentMediaSource(media.thumbUrl),
    normalizeMomentMediaSource(media.url),
  ];
  const sourceList = mediaType === 'video' ? videoFirst : imageFirst;
  return Array.from(new Set(sourceList.filter(Boolean)));
}

function getMomentMediaSrc(media: MomentSummary['media'][number], index: number, mediaType: string) {
  const candidates = getMomentMediaCandidates(media, mediaType);
  if (!candidates.length) return '';
  const key = getMomentMediaKey(index);
  const currentIndex = mediaFallbackIndex.value[key] || 0;
  return candidates[currentIndex] || candidates[0] || '';
}

function handleMomentMediaError(media: MomentSummary['media'][number], index: number, mediaType: string) {
  const candidates = getMomentMediaCandidates(media, mediaType);
  if (candidates.length <= 1) return;
  const key = getMomentMediaKey(index);
  const currentIndex = mediaFallbackIndex.value[key] || 0;
  if (currentIndex >= candidates.length - 1) return;
  mediaFallbackIndex.value = {
    ...mediaFallbackIndex.value,
    [key]: currentIndex + 1,
  };
}

function handleMomentListScroll(event: Event) {
  const element = event.target as HTMLElement | null;
  if (!element || loading.list || loading.loadMore || !hasMore.value) return;
  if (element.scrollTop + element.clientHeight >= element.scrollHeight - 160) {
    void loadMoments(false);
  }
}

onMounted(async () => {
  try {
    await activateMomentsWorkspace();
  } finally {
    momentsBootstrapping.value = false;
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
      <Button :loading="loading.list" @click="loadMoments(true)">刷新列表</Button>
      <Button type="primary" size="small" class="!h-8 !rounded-lg !px-3" @click="openPublishModal">
        发布朋友圈
      </Button>
    </div>

    <div class="h-[calc(100vh-210px)] min-h-[680px] overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <div class="flex h-full">
        <aside class="flex h-full w-[390px] min-w-[360px] flex-col border-r border-slate-200 bg-slate-50/70 p-4">
          <div class="flex items-center justify-between rounded-2xl border border-slate-200 bg-white px-4 py-4">
            <div>
              <div class="text-[26px] font-semibold tracking-tight text-slate-900">朋友圈</div>
              <div class="mt-1 text-sm text-slate-500">当前账号动态列表</div>
            </div>
            <div class="rounded-2xl bg-[rgb(var(--primary-1))] px-4 py-2 text-xl font-semibold text-[rgb(var(--primary-6))]">
              {{ filteredMoments.length }}
            </div>
          </div>

          <Input.Search
            v-model:value="listKeyword"
            allow-clear
            class="!mt-3"
            placeholder="搜索朋友圈作者 / 内容"
          />

          <div ref="momentListPaneRef" class="mt-4 min-h-0 flex-1 overflow-auto pr-1" @scroll.passive="handleMomentListScroll">
            <Spin :spinning="loading.list">
              <Space class="w-full" direction="vertical" size="small">
                <button
                  v-for="item in filteredMoments"
                  :key="item.id"
                  :class="[
                    'w-full rounded-2xl border px-4 py-4 text-left transition',
                    selectedMomentId === item.id
                      ? 'border-[rgb(var(--primary-5))] bg-[rgb(var(--primary-1))]'
                      : 'border-slate-200 bg-white hover:border-slate-300',
                  ]"
                  type="button"
                  @click="selectedMomentId = item.id"
                >
                  <div class="flex items-start gap-3">
                    <Avatar :size="48" :src="item.authorAvatar || undefined">
                      {{ item.authorName?.slice(0, 1) || '友' }}
                    </Avatar>
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center justify-between gap-3">
                        <Typography.Text class="truncate text-[15px] font-semibold text-slate-900">
                          {{ item.authorName }}
                        </Typography.Text>
                        <Typography.Text class="shrink-0 text-xs text-slate-400">
                          {{ formatTime(item.createdAt) }}
                        </Typography.Text>
                      </div>
                      <div class="mt-1 line-clamp-2 text-sm leading-6 text-slate-600">
                        {{ item.preview || '暂无文案内容' }}
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Tag :bordered="false" :color="mediaTagColor(item.mediaType)">
                          {{ mediaTypeLabel(item.mediaType) }}
                        </Tag>
                        <Tag :bordered="false" :color="item.visibility === 'private' ? 'volcano' : 'blue'">
                          {{ item.visibility === 'private' ? '仅自己可见' : '公开' }}
                        </Tag>
                        <Typography.Text class="text-xs text-slate-400">
                          {{ item.likeCount }} 赞 · {{ item.commentCount }} 评论
                        </Typography.Text>
                      </div>
                    </div>
                  </div>
                </button>
              </Space>
              <Empty
                v-if="listLoaded && !filteredMoments.length && !loading.list"
                class="pt-16"
                description="暂无朋友圈动态"
              />
            </Spin>
            <div v-if="loading.loadMore" class="py-4 text-center text-xs text-slate-400">
              正在继续加载朋友圈...
            </div>
            <div v-else-if="hasMore && filteredMoments.length" class="py-4 text-center text-xs text-slate-400">
              下滑继续加载更多动态
            </div>
          </div>
        </aside>

        <section class="min-w-0 flex-1 bg-slate-50">
          <template v-if="currentMoment">
            <div class="h-full overflow-auto p-4">
              <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white">
                <div class="border-b border-slate-200 px-5 py-5">
                  <div class="flex flex-wrap items-start justify-between gap-4">
                    <div class="flex min-w-0 items-start gap-4">
                      <Avatar :size="64" :src="currentMoment.authorAvatar || undefined" class="shrink-0">
                        {{ currentMoment.authorName?.slice(0, 1) || '友' }}
                      </Avatar>
                      <div class="min-w-0">
                        <div class="truncate text-[32px] font-semibold tracking-tight text-slate-900">
                          {{ currentMoment.authorName }}
                        </div>
                        <div class="mt-1 truncate text-lg text-slate-400">
                          {{ currentMoment.authorWxid }}
                        </div>
                        <div class="mt-3 flex flex-wrap items-center gap-2">
                          <Tag :bordered="false" color="purple">朋友圈</Tag>
                          <Tag :bordered="false" :color="mediaTagColor(currentMoment.mediaType)">
                            {{ mediaTypeLabel(currentMoment.mediaType) }}
                          </Tag>
                          <Tag :bordered="false" :color="currentMoment.visibility === 'private' ? 'volcano' : 'blue'">
                            {{ currentMoment.visibility === 'private' ? '仅自己可见' : '公开' }}
                          </Tag>
                          <Tag v-if="currentMoment.liked" :bordered="false" color="green">已点赞</Tag>
                        </div>
                      </div>
                    </div>
                    <div class="flex flex-wrap items-center justify-end gap-2">
                      <Button @click="loadMomentDetail(currentMoment)" :loading="loading.detail">刷新详情</Button>
                      <Button
                        v-if="currentMoment.canEdit && currentMoment.visibility !== 'private'"
                        @click="operateMoment('private')"
                        :loading="loading.operate"
                      >
                        设为私密
                      </Button>
                      <Button
                        v-if="currentMoment.canEdit && currentMoment.visibility === 'private'"
                        @click="operateMoment('public')"
                        :loading="loading.operate"
                      >
                        设为公开
                      </Button>
                      <Button
                        v-if="currentMoment.canEdit"
                        danger
                        @click="operateMoment('delete')"
                        :loading="loading.operate"
                      >
                        删除动态
                      </Button>
                      <Button type="primary" size="small" class="!h-8 !rounded-lg !px-3" @click="openPublishModal">
                        发布朋友圈
                      </Button>
                    </div>
                  </div>

                  <div class="mt-5 grid gap-3 rounded-2xl border border-slate-200 bg-slate-50/70 p-4 md:grid-cols-5">
                    <div
                      v-for="item in currentMetaItems"
                      :key="item.label"
                      class="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                    >
                      <div class="text-sm text-slate-400">{{ item.label }}</div>
                      <div class="mt-2 text-lg font-semibold text-slate-900">{{ item.value }}</div>
                    </div>
                  </div>
                </div>

                <div class="grid gap-4 border-b border-slate-200 px-5 py-5 xl:grid-cols-[minmax(0,1.2fr)_minmax(360px,0.9fr)]">
                  <div class="space-y-4">
                    <div class="rounded-2xl border border-slate-200 bg-white p-5">
                      <div class="text-xl font-semibold text-slate-900">正文内容</div>
                      <div class="mt-4 whitespace-pre-wrap text-[15px] leading-7 text-slate-700">
                        {{ currentMoment.content || '暂无正文内容' }}
                      </div>
                    </div>

                    <div class="grid gap-4 xl:grid-cols-2">
                      <div class="rounded-2xl border border-slate-200 bg-white p-5">
                        <div class="flex items-center justify-between">
                          <div class="text-xl font-semibold text-slate-900">点赞列表</div>
                          <Tag :bordered="false" color="green">{{ currentMoment.likes.length }}</Tag>
                        </div>
                        <div v-if="currentMoment.likes.length" class="mt-4 space-y-3">
                          <div
                            v-for="item in currentMoment.likes"
                            :key="`${item.username}-${item.commentId}`"
                            class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3"
                          >
                            <div class="font-medium text-slate-900">{{ item.nickname }}</div>
                            <div class="mt-1 text-xs text-slate-400">{{ item.username }}</div>
                          </div>
                        </div>
                        <Empty v-else class="mt-8" description="暂无点赞" />
                      </div>

                      <div class="rounded-2xl border border-slate-200 bg-white p-5">
                        <div class="flex items-center justify-between">
                          <div class="text-xl font-semibold text-slate-900">评论列表</div>
                          <Tag :bordered="false" color="blue">{{ currentMoment.comments.length }}</Tag>
                        </div>
                        <div v-if="currentMoment.comments.length" class="mt-4 space-y-3">
                          <div
                            v-for="item in currentMoment.comments"
                            :key="`${item.username}-${item.commentId}`"
                            class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3"
                          >
                            <div class="flex items-center justify-between gap-3">
                              <div class="font-medium text-slate-900">{{ item.nickname }}</div>
                              <div class="text-xs text-slate-400">{{ formatTime(item.createTime) }}</div>
                            </div>
                            <div class="mt-1 text-sm leading-6 text-slate-600">{{ item.content || '无文字评论' }}</div>
                          </div>
                        </div>
                        <Empty v-else class="mt-8" description="暂无评论" />
                      </div>
                    </div>
                  </div>

                  <div class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="flex items-center justify-between">
                      <div class="text-xl font-semibold text-slate-900">媒体预览</div>
                      <Tag :bordered="false" :color="mediaTagColor(currentMoment.mediaType)">
                        {{ currentMoment.mediaCount }} 个媒体
                      </Tag>
                    </div>
                    <div
                      v-if="currentMoment.media.length"
                      class="mt-4 grid gap-3"
                      :class="currentMoment.mediaType === 'video' ? 'grid-cols-1' : 'sm:grid-cols-2'"
                    >
                      <div
                        v-for="(media, index) in currentMoment.media"
                        :key="`${currentMoment.id}-${index}`"
                        class="overflow-hidden rounded-2xl border border-slate-200 bg-slate-50"
                      >
                        <template v-if="getMomentMediaSrc(media, index, currentMoment.mediaType)">
                          <video
                            v-if="currentMoment.mediaType === 'video' && normalizeMomentMediaSource(media.url)"
                            :controls="true"
                            :poster="getMomentMediaSrc(media, index, 'video')"
                            :src="normalizeMomentMediaSource(media.url)"
                            class="max-h-[260px] w-full bg-black object-contain"
                            preload="metadata"
                          />
                          <Image
                            v-else
                            :fallback="normalizeMomentMediaSource(media.thumbUrl) || undefined"
                            :preview="true"
                            :src="getMomentMediaSrc(media, index, currentMoment.mediaType)"
                            class="block w-full"
                            @error="handleMomentMediaError(media, index, currentMoment.mediaType)"
                          />
                        </template>
                        <div v-else class="flex h-[220px] items-center justify-center text-sm text-slate-400">
                          暂无可预览媒体
                        </div>
                        <div class="px-4 py-3 text-sm text-slate-500">
                          {{ mediaTypeLabel(currentMoment.mediaType) }} · {{ media.width || '-' }} x {{ media.height || '-' }}
                        </div>
                      </div>
                    </div>
                    <Empty v-else class="mt-6" description="当前朋友圈没有媒体内容" />
                  </div>
                </div>

                <div class="px-5 py-5">
                  <div class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="flex items-center justify-between">
                      <div class="text-xl font-semibold text-slate-900">动态原始信息</div>
                      <Tag :bordered="false" color="default">调试信息</Tag>
                    </div>
                    <div class="mt-4 text-sm leading-7 text-slate-500">
                      作者：{{ currentMoment.authorWxid }}，ID：{{ currentMoment.id }}，原始媒体数：{{ currentMoment.media.length }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>
          <Empty v-else class="pt-24" description="请选择左侧朋友圈动态" />
        </section>
      </div>
    </div>

    <Modal
      v-model:open="publishModal.open"
      :confirm-loading="loading.publish"
      centered
      ok-text="提交发布"
      title="发布朋友圈"
      width="680px"
      @ok="submitPublishMoment"
    >
      <div class="space-y-4">
        <div>
          <div class="mb-2 text-sm font-medium text-slate-700">动态内容</div>
          <Input.TextArea
            v-model:value="publishModal.content"
            :auto-size="{ minRows: 5, maxRows: 10 }"
            placeholder="请输入朋友圈文案内容"
          />
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <div class="mb-2 text-sm font-medium text-slate-700">提醒谁看</div>
            <Input
              v-model:value="publishModal.withUserList"
              placeholder="多个 wxid 用逗号、空格或换行分隔"
            />
          </div>
          <div>
            <div class="mb-2 text-sm font-medium text-slate-700">不给谁看</div>
            <Input
              v-model:value="publishModal.blackList"
              placeholder="多个 wxid 用逗号、空格或换行分隔"
            />
          </div>
        </div>
        <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm leading-6 text-slate-500">
          当前先接通基础文本发布链路，底层仍依赖 `wechatReal` 的朋友圈发布能力；如上游要求更完整 XML/媒体参数，会按其返回结果提示。
        </div>
      </div>
    </Modal>
  </Page>
</template>
