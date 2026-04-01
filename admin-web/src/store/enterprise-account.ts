import { computed, ref } from 'vue';

import { defineStore } from 'pinia';

import { listAccountsApi, type WechatAccount } from '#/api';

const STORAGE_KEY = 'enterprise:selected-wxid';

function readStoredWxid() {
  if (typeof window === 'undefined') return '';
  try {
    return window.localStorage.getItem(STORAGE_KEY) || '';
  } catch {
    return '';
  }
}

function writeStoredWxid(wxid: string) {
  if (typeof window === 'undefined') return;
  try {
    if (wxid) {
      window.localStorage.setItem(STORAGE_KEY, wxid);
      return;
    }
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
    // ignore storage failures
  }
}

export const useEnterpriseAccountStore = defineStore('enterprise-account', () => {
  const accounts = ref<WechatAccount[]>([]);
  const initialized = ref(false);
  const loading = ref(false);
  const lastLoadedAt = ref(0);
  const selectedWxidState = ref(readStoredWxid());

  const selectedWxid = computed({
    get: () => selectedWxidState.value,
    set: (value: string) => {
      selectedWxidState.value = value;
      writeStoredWxid(value);
    },
  });

  const currentAccount = computed(
    () => accounts.value.find((item) => item.wxid === selectedWxid.value) || null,
  );

  const accountOptions = computed(() =>
    accounts.value.map((item) => ({
      label: `${item.nickname || item.alias || item.wxid} · ${item.status === 'online' ? '在线' : '离线'}`,
      value: item.wxid,
    })),
  );

  function setSelectedWxid(wxid: string) {
    selectedWxid.value = wxid;
  }

  function resolvePreferredWxid(preferredWxid?: string) {
    const preferred = preferredWxid || selectedWxidState.value;
    if (preferred && accounts.value.some((item) => item.wxid === preferred)) {
      return preferred;
    }
    return accounts.value[0]?.wxid ?? '';
  }

  async function refreshAccounts(options?: { preferredWxid?: string }) {
    if (loading.value) {
      return accounts.value;
    }

    loading.value = true;
    try {
      const result = await listAccountsApi();
      accounts.value = result;
      initialized.value = true;
      lastLoadedAt.value = Date.now();
      selectedWxid.value = resolvePreferredWxid(options?.preferredWxid);
      return accounts.value;
    } finally {
      loading.value = false;
    }
  }

  async function ensureAccounts(options?: { preferredWxid?: string }) {
    if (initialized.value) {
      selectedWxid.value = resolvePreferredWxid(options?.preferredWxid);
      return accounts.value;
    }
    return refreshAccounts(options);
  }

  return {
    accountOptions,
    accounts,
    currentAccount,
    ensureAccounts,
    initialized,
    lastLoadedAt,
    loading,
    refreshAccounts,
    selectedWxid,
    setSelectedWxid,
  };
});
