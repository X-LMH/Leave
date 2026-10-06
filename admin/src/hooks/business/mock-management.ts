import { computed, reactive, ref, shallowRef } from 'vue';
import type { PageQuery } from '@/service/mock/management';

/** Shared pagination for the two local management lists, including stale-query protection. */
export function useMockManagement<T, Q extends object>(
  emptyQuery: () => Q,
  query: (input: Q & PageQuery) => Promise<{ items: T[]; total: number }>
) {
  const search = reactive(emptyQuery());
  const applied = shallowRef(emptyQuery());
  const rows = shallowRef<T[]>([]);
  const loading = ref(false);
  const total = ref(0);
  const page = ref(1);
  const pageSize = ref(10);
  let sequence = 0;
  async function reload(): Promise<void> {
    const current = ++sequence;
    loading.value = true;
    try {
      const result = await query({ ...applied.value, page: page.value, pageSize: pageSize.value });
      if (current !== sequence) return;
      total.value = result.total;
      const lastPage = Math.max(1, Math.ceil(result.total / pageSize.value));
      if (page.value > lastPage) {
        page.value = lastPage;
        await reload();
        return;
      }
      rows.value = result.items;
    } catch {
      if (current === sequence) window.$message?.error('加载失败，请重试');
    } finally {
      if (current === sequence) loading.value = false;
    }
  }
  function applySearch() {
    applied.value = { ...search } as Q;
    page.value = 1;
    void reload();
  }
  function resetSearch() {
    Object.assign(search, emptyQuery());
    applySearch();
  }
  const pagination = computed(() => ({
    page: page.value,
    pageSize: pageSize.value,
    itemCount: total.value,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
    onUpdatePage: (value: number) => {
      page.value = value;
      void reload();
    },
    onUpdatePageSize: (value: number) => {
      pageSize.value = value;
      page.value = 1;
      void reload();
    }
  }));
  void reload();
  return { search, rows, loading, total, pagination, reload, applySearch, resetSearch };
}
