<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Avatar,
  Button,
  Empty,
  Image,
  Popconfirm,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'ant-design-vue';

import {
  deleteFavoriteApi,
  getFavoriteDetailApi,
  listFavoritesApi,
  type FavoriteSummary,
} from '#/api';
import { useEnterpriseAccountStore } from '#/store';

const accountStore = useEnterpriseAccountStore();

const loading = reactive({
  accounts: false,
  deleting: false,
  detail: false,
  list: false,
});

const selectedWxid = computed({
  get: () => accountStore.selectedWxid,
  set: (value: string) => accountStore.setSelectedWxid(value),
});
const accountOptions = computed(() => accountStore.accountOptions);

const favorites = ref<FavoriteSummary[]>([]);
const selectedFavId = ref<number>(0);
const detailCache = ref<Record<number, FavoriteSummary>>({});
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const totalPages = ref(1);
const latestListRequestId = ref(0);
const listLoaded = ref(false);

const currentFavorite = computed(() => {
  const detail = selectedFavId.value ? detailCache.value[selectedFavId.value] : undefined;
  const summary = detail ?? favorites.value.find((item) => item.favId === selectedFavId.value);
  if (!summary) return null;
  return {
    ...summary,
    media: Array.isArray(summary.media) ? summary.media : [],
  };
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

function kindColor(kind: string) {
  if (kind === 'image') return 'green';
  if (kind === 'video') return 'purple';
  if (kind === 'link') return 'blue';
  if (kind === 'file') return 'orange';
  if (kind === 'card') return 'cyan';
  return 'default';
}

function resetWorkspace() {
  favorites.value = [];
  detailCache.value = {};
  selectedFavId.value = 0;
  total.value = 0;
  totalPages.value = 1;
  listLoaded.value = false;
}

function isCurrentRequest(requestId: number, wxid: string) {
  return latestListRequestId.value === requestId && selectedWxid.value === wxid;
}

async function loadAccounts() {
  loading.accounts = true;
  try {
    await accountStore.ensureAccounts();
  } finally {
    loading.accounts = false;
  }
}

async function loadFavorites(targetPage = 1) {
  if (!selectedWxid.value) return;
  const requestWxid = selectedWxid.value;
  const requestId = latestListRequestId.value + 1;
  latestListRequestId.value = requestId;
  loading.list = true;
  try {
    const result = await listFavoritesApi(requestWxid, {
      page: targetPage,
      pageSize: pageSize.value,
    });
    if (!isCurrentRequest(requestId, requestWxid)) return;
    favorites.value = Array.isArray(result.items) ? result.items : [];
    page.value = result.page || targetPage;
    pageSize.value = result.pageSize || pageSize.value;
    total.value = result.total || 0;
    totalPages.value = result.totalPages || 1;
    listLoaded.value = true;
    const nextSelected = favorites.value.find((item) => item.favId === selectedFavId.value)?.favId
      ?? favorites.value[0]?.favId
      ?? 0;
    selectedFavId.value = nextSelected;
    if (nextSelected) {
      void loadFavoriteDetail(nextSelected, false);
    }
  } catch (error: any) {
    if (!isCurrentRequest(requestId, requestWxid)) return;
    resetWorkspace();
    message.error(error?.message || '加载收藏列表失败');
  } finally {
    if (!isCurrentRequest(requestId, requestWxid)) return;
    loading.list = false;
  }
}

async function loadFavoriteDetail(favId: number, force = false) {
  if (!selectedWxid.value || !favId) return;
  if (!force && detailCache.value[favId]) return;
  loading.detail = true;
  try {
    const result = await getFavoriteDetailApi(selectedWxid.value, favId);
    detailCache.value = {
      ...detailCache.value,
      [favId]: result.item,
    };
  } catch (error: any) {
    message.error(error?.message || '加载收藏详情失败');
  } finally {
    loading.detail = false;
  }
}

async function handleDeleteFavorite() {
  if (!selectedWxid.value || !selectedFavId.value) return;
  loading.deleting = true;
  try {
    const result = await deleteFavoriteApi(selectedWxid.value, selectedFavId.value);
    message.success(result.message || '删除成功');
    detailCache.value = {};
    await loadFavorites(Math.min(page.value, totalPages.value));
  } catch (error: any) {
    message.error(error?.message || '删除收藏失败');
  } finally {
    loading.deleting = false;
  }
}

function canPreviewImage(src: string) {
  return /^https?:\/\//i.test(src) || /^data:/i.test(src);
}

function prevPage() {
  if (page.value <= 1) return;
  void loadFavorites(page.value - 1);
}

function nextPage() {
  if (page.value >= totalPages.value) return;
  void loadFavorites(page.value + 1);
}

watch(selectedWxid, async () => {
  resetWorkspace();
  page.value = 1;
  if (selectedWxid.value) {
    await loadFavorites(1);
  }
});

watch(selectedFavId, (favId) => {
  if (favId) {
    void loadFavoriteDetail(favId, false);
  }
});

onMounted(async () => {
  await loadAccounts();
  if (selectedWxid.value) {
    await loadFavorites(1);
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
      <Button :loading="loading.list" @click="loadFavorites(page)">刷新收藏</Button>
    </div>

    <div class="h-[calc(100vh-220px)] min-h-[680px] overflow-hidden rounded-2xl border border-slate-200 bg-white">
      <div class="flex h-full">
        <aside class="flex h-full w-[360px] min-w-[340px] flex-col border-r border-slate-200 bg-slate-50/60 p-4">
          <div class="rounded-2xl border border-slate-200 bg-white p-5">
            <div class="flex items-start justify-between gap-4">
              <div>
                <div class="text-[18px] font-semibold text-slate-900">收藏管理</div>
                <div class="mt-1 text-sm text-slate-500">当前账号微信收藏列表</div>
              </div>
              <div class="rounded-xl bg-[rgb(var(--primary-1))] px-3 py-1 text-lg font-semibold text-[rgb(var(--primary-6))]">
                {{ total }}
              </div>
            </div>
          </div>

          <div class="mt-4 min-h-0 flex-1 overflow-auto pr-1">
            <Spin :spinning="loading.list">
              <div class="space-y-3">
                <button
                  v-for="item in favorites"
                  :key="item.favId"
                  :class="[
                    'w-full rounded-2xl border p-4 text-left transition',
                    selectedFavId === item.favId
                      ? 'border-slate-400 bg-white shadow-[0_12px_30px_rgba(15,23,42,0.08)]'
                      : 'border-slate-200 bg-white hover:border-slate-300',
                  ]"
                  type="button"
                  @click="selectedFavId = item.favId"
                >
                  <div class="flex items-start gap-3">
                    <Avatar
                      :size="48"
                      :src="item.media[0]?.thumbUrl || undefined"
                      class="shrink-0"
                    >
                      {{ item.kindLabel.slice(0, 1) }}
                    </Avatar>
                    <div class="min-w-0 flex-1">
                      <div class="flex items-start justify-between gap-3">
                        <div class="truncate text-[18px] font-semibold text-slate-900">
                          {{ item.title || `收藏 #${item.favId}` }}
                        </div>
                        <div class="shrink-0 text-sm text-slate-500">
                          {{ formatTime(item.updatedAt) }}
                        </div>
                      </div>
                      <div class="mt-1 line-clamp-2 text-sm leading-6 text-slate-600">
                        {{ item.preview || '暂无摘要' }}
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2 text-xs">
                        <Tag :color="kindColor(item.kind)">{{ item.kindLabel }}</Tag>
                        <Tag color="blue">{{ item.mediaCount }} 个附件</Tag>
                        <Tag v-if="item.sourceDisplay">{{ item.sourceDisplay }}</Tag>
                      </div>
                    </div>
                  </div>
                </button>

                <Empty v-if="!favorites.length && listLoaded" description="暂无收藏内容" />
              </div>
            </Spin>
          </div>

          <div class="mt-4 flex items-center justify-between rounded-2xl border border-slate-200 bg-white px-4 py-3">
            <div class="text-sm text-slate-500">
              第 {{ page }} / {{ totalPages }} 页
            </div>
            <Space>
              <Button :disabled="page <= 1 || loading.list" size="small" @click="prevPage">
                上一页
              </Button>
              <Button :disabled="page >= totalPages || loading.list" size="small" @click="nextPage">
                下一页
              </Button>
            </Space>
          </div>
        </aside>

        <section class="flex min-w-0 flex-1 flex-col bg-slate-50/40">
          <template v-if="currentFavorite">
            <header class="border-b border-slate-200 bg-white px-6 py-5">
              <div class="flex items-start justify-between gap-6">
                <div class="min-w-0">
                  <div class="text-[30px] font-semibold leading-none text-slate-900">
                    {{ currentFavorite.title || `收藏 #${currentFavorite.favId}` }}
                  </div>
                  <div class="mt-3 text-lg text-slate-500">
                    收藏 ID：{{ currentFavorite.favId }}
                  </div>
                  <div class="mt-4 flex flex-wrap items-center gap-2">
                    <Tag :color="kindColor(currentFavorite.kind)">{{ currentFavorite.kindLabel }}</Tag>
                    <Tag color="blue">{{ formatTime(currentFavorite.updatedAt) }}</Tag>
                    <Tag v-if="currentFavorite.sourceDisplay">{{ currentFavorite.sourceDisplay }}</Tag>
                  </div>
                </div>
                <Space>
                  <Button :loading="loading.detail" @click="loadFavoriteDetail(currentFavorite.favId, true)">
                    刷新详情
                  </Button>
                  <Popconfirm
                    ok-text="删除"
                    cancel-text="取消"
                    title="删除后无法恢复，确认删除这条收藏吗？"
                    @confirm="handleDeleteFavorite"
                  >
                    <Button danger :loading="loading.deleting">删除收藏</Button>
                  </Popconfirm>
                </Space>
              </div>
            </header>

            <div class="min-h-0 flex-1 overflow-auto px-6 py-5">
              <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
                <div class="space-y-5">
                  <section class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="text-xl font-semibold text-slate-900">正文内容</div>
                    <div class="mt-4 whitespace-pre-wrap text-[15px] leading-7 text-slate-700">
                      {{ currentFavorite.preview || currentFavorite.title || '暂无正文内容' }}
                    </div>
                  </section>

                  <section class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="mb-4 text-xl font-semibold text-slate-900">原始收藏 XML</div>
                    <pre class="max-h-[320px] overflow-auto rounded-2xl bg-slate-950 p-4 text-xs leading-6 text-slate-100">{{ currentFavorite.rawXml || '暂无原始 XML' }}</pre>
                  </section>
                </div>

                <div class="space-y-5">
                  <section class="rounded-2xl border border-slate-200 bg-white p-5">
                    <div class="mb-4 text-xl font-semibold text-slate-900">媒体预览</div>
                    <div v-if="currentFavorite.media.length" class="space-y-4">
                      <div
                        v-for="(media, index) in currentFavorite.media"
                        :key="`${currentFavorite.favId}-${index}`"
                        class="overflow-hidden rounded-2xl border border-slate-200"
                      >
                        <div class="border-b border-slate-200 bg-slate-50 px-4 py-2 text-sm font-medium text-slate-600">
                          {{ media.title || media.description || `${currentFavorite.kindLabel}附件 ${index + 1}` }}
                        </div>
                        <div class="p-3">
                          <Image
                            v-if="media.type === 'image' && canPreviewImage(media.thumbUrl || media.url)"
                            :src="media.thumbUrl || media.url"
                            class="max-h-[260px] w-full rounded-xl object-cover"
                          />
                          <video
                            v-else-if="media.type === 'video' && canPreviewImage(media.url)"
                            :src="media.url"
                            class="max-h-[260px] w-full rounded-xl bg-black"
                            controls
                          />
                          <div
                            v-else
                            class="rounded-xl bg-slate-50 px-4 py-6 text-sm leading-6 text-slate-500"
                          >
                            暂无直接可预览媒体地址
                            <div v-if="media.url" class="mt-2 break-all text-xs text-slate-400">
                              {{ media.url }}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <Empty v-else description="当前收藏没有媒体附件" />
                  </section>
                </div>
              </div>
            </div>
          </template>

          <div v-else class="flex h-full items-center justify-center">
            <Empty description="请选择一条收藏查看详情" />
          </div>
        </section>
      </div>
    </div>
  </Page>
</template>
