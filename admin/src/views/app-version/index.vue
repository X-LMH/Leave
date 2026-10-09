<script setup lang="ts">
import { computed, h, ref } from 'vue';
import { NButton, NTag, useThemeVars } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { useMockManagement } from '@/hooks/business/mock-management';
import { getAppVersion, getAppVersionOverview, queryAppVersions } from '@/service/mock/app-version';
import type { AppVersionQuery, AppVersionRecord } from '@/service/mock/app-version';
import { formatTime } from '@/service/mock/management';
import VersionDrawer from './components/version-drawer.vue';

const theme = useThemeVars();
const themeStyle = computed(() => ({
  '--management-accent': theme.value.primaryColor,
  '--management-text': theme.value.textColor1,
  '--management-muted': theme.value.textColor3,
  '--management-border': theme.value.dividerColor,
  '--management-surface': theme.value.cardColor,
  '--management-inset': theme.value.tableHeaderColor
}));
const { search, rows, loading, total, pagination, reload, applySearch, resetSearch } = useMockManagement<
  AppVersionRecord,
  AppVersionQuery
>(() => ({ keyword: '', platform: null, status: null }), queryAppVersions);
const current = ref<AppVersionRecord | null>(null);
const allCount = ref(0);
const overviewLoading = ref(true);
const overviewError = ref(false);
const drawerVisible = ref(false);
const opening = ref(false);
const selected = ref<AppVersionRecord | null>(null);
const statusOptions = [
  { label: '最新版本', value: 'published' },
  { label: '历史版本', value: 'archived' }
];
const platformOptions = [{ label: 'Android', value: 'android' }];
let overviewSequence = 0;
async function loadOverview() {
  const sequence = ++overviewSequence;
  overviewLoading.value = true;
  overviewError.value = false;
  try {
    const result = await getAppVersionOverview();
    if (sequence !== overviewSequence) return;
    current.value = result.current;
    allCount.value = result.total;
  } catch {
    if (sequence === overviewSequence) overviewError.value = true;
  } finally {
    if (sequence === overviewSequence) overviewLoading.value = false;
  }
}
async function refresh() {
  await Promise.all([loadOverview(), reload()]);
}
function downloadLatestPackage() {
  window.$message?.info('安装包下载功能待接入，暂未开放下载');
}
async function openDrawer(row: AppVersionRecord) {
  if (opening.value || drawerVisible.value) return;
  opening.value = true;
  try {
    selected.value = await getAppVersion(row.id);
    drawerVisible.value = true;
  } catch (error) {
    window.$message?.error(error instanceof Error ? error.message : '加载版本失败，请重试');
  } finally {
    opening.value = false;
  }
}
const columns: DataTableColumns<AppVersionRecord> = [
  {
    key: 'versionName',
    title: '版本',
    width: '12%',
    render: row =>
      h('div', { class: 'release-table-version' }, [
        h('strong', {}, row.versionName),
        h('span', {}, `构建 ${row.versionCode}`)
      ])
  },
  { key: 'platform', title: '平台', width: '9%', render: () => 'Android' },
  {
    key: 'status',
    title: '状态',
    width: '10%',
    render: row =>
      h(NTag, { size: 'small', bordered: false, type: row.status === 'published' ? 'success' : 'default' }, () =>
        row.status === 'published' ? '最新版本' : '历史版本'
      )
  },
  {
    key: 'publishedAt',
    title: '发布时间',
    width: '14%',
    render: row => {
      const date = new Date(row.publishedAt);
      return h('div', { class: 'release-table-time' }, [
        h('span', {}, date.toLocaleDateString('zh-CN')),
        h('span', {}, date.toLocaleTimeString('zh-CN', { hour12: false }))
      ]);
    }
  },
  {
    key: 'releaseNotes',
    title: '更新说明',
    width: '25%',
    ellipsis: { tooltip: false },
    render: row => {
      const notes = row.releaseNotes?.join('；') || '—';
      const characters = Array.from(notes);
      const preview = characters.length > 40 ? `${characters.slice(0, 40).join('')}...` : notes;
      return h('span', { title: notes }, preview);
    }
  },
  {
    key: 'packageFile',
    title: '安装包文件名',
    width: '23%',
    ellipsis: { tooltip: true },
    className: 'release-table-package'
  },
  {
    key: 'actions',
    title: '操作',
    width: '7%',
    fixed: 'right',
    render: row =>
      h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          secondary: true,
          disabled: opening.value,
          onClick: () => openDrawer(row)
        },
        () => '详情'
      )
  }
];
void loadOverview();
</script>

<template>
  <div class="management-page release-page" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <div>
          <p class="management-header__eyebrow">客户端管理 / 版本记录</p>
          <h1 class="management-header__title">版本管理</h1>
          <p class="management-header__description">查看 Android 客户端版本与更新记录。</p>
        </div>
      </div>
    </header>

    <section class="release-current" aria-label="Android 最新版本" aria-live="polite">
      <NSpin :show="overviewLoading">
        <div v-if="overviewError" class="release-current__empty">
          <strong>最新版本加载失败</strong>
          <NButton size="small" @click="loadOverview">重新加载</NButton>
        </div>
        <div v-else-if="current">
          <div class="release-current__label">
            <span>最新版本</span>
            <span class="release-current__platform">Android</span>
          </div>
          <h2 class="release-current__version">{{ current.versionName }}</h2>
          <div class="release-current__meta">
            <span>构建 {{ current.versionCode }}</span>
            <span>发布于 {{ formatTime(current.publishedAt) }}</span>
          </div>
          <section class="release-current__notes" aria-label="完整更新说明">
            <h3>更新说明</h3>
            <ul v-if="current.releaseNotes?.length">
              <li v-for="(note, index) in current.releaseNotes" :key="index">{{ note }}</li>
            </ul>
            <p v-else>该版本未填写更新说明。</p>
          </section>
          <div class="release-current__actions">
            <NButton size="small" type="primary" @click="downloadLatestPackage">
              <template #icon><SvgIcon icon="mdi:download" /></template>
              下载安装包
            </NButton>
            <NButton text type="primary" :disabled="opening" @click="openDrawer(current)">
              查看详情
              <template #icon><SvgIcon icon="mdi:arrow-right" /></template>
            </NButton>
          </div>
        </div>
        <div v-else class="release-current__empty">
          <strong>{{ overviewLoading ? '正在获取最新版本…' : '暂无最新版本' }}</strong>
          <p v-if="!overviewLoading">暂无已发布的 Android 客户端版本记录。</p>
        </div>
      </NSpin>
    </section>

    <NCard :bordered="false" size="small" class="management-card management-card--table release-list">
      <template #header>
        <div class="section-heading">
          <span>版本记录</span>
          <span class="section-heading__badge">{{ total }}</span>
        </div>
      </template>
      <template #header-extra>
        <NButton quaternary size="small" :loading="loading || overviewLoading" @click="refresh">
          <template #icon><SvgIcon icon="mdi:refresh" /></template>
          刷新
        </NButton>
      </template>
      <NForm :show-feedback="false" :show-label="false" class="release-search">
        <NFormItem label="版本关键词">
          <NInput
            v-model:value="search.keyword"
            clearable
            placeholder="搜索版本号或构建号"
            aria-label="版本关键词"
            @keyup.enter="applySearch"
          />
        </NFormItem>
        <NFormItem label="客户端平台">
          <NSelect
            v-model:value="search.platform"
            :options="platformOptions"
            clearable
            placeholder="全部平台"
            aria-label="客户端平台"
          />
        </NFormItem>
        <NFormItem label="发布状态">
          <NSelect
            v-model:value="search.status"
            :options="statusOptions"
            clearable
            placeholder="全部状态"
            aria-label="发布状态"
          />
        </NFormItem>
        <div class="search-actions">
          <NButton type="primary" :loading="loading" @click="applySearch">
            <template #icon><SvgIcon icon="mdi:magnify" /></template>
            查询
          </NButton>
          <NButton @click="resetSearch">重置</NButton>
        </div>
      </NForm>
      <NDataTable
        :columns="columns"
        :data="rows"
        :row-key="row => row.id"
        :loading="loading"
        :pagination="pagination"
        remote
        :bordered="false"
        table-layout="fixed"
        :scroll-x="1200"
      >
        <template #empty>
          <div class="table-empty">
            <SvgIcon icon="mdi:package-variant" aria-hidden="true" />
            <strong>{{ allCount ? '未找到匹配的版本' : '暂无版本记录' }}</strong>
            <span>{{ allCount ? '请调整关键词、平台或发布状态。' : '暂无 Android 客户端版本发布记录。' }}</span>
            <NButton v-if="allCount" size="small" @click="resetSearch">重置筛选</NButton>
          </div>
        </template>
      </NDataTable>
      <div class="release-table-footer">按构建号从高到低排列；历史版本保留完整更新说明。</div>
    </NCard>
    <VersionDrawer v-model:show="drawerVisible" :record="selected" />
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';

.release-page {
  gap: 16px;
}
.release-page .management-header {
  padding: 6px 0 6px 14px;
  border-left: 3px solid var(--management-accent);
}
.release-page .management-header__eyebrow {
  color: var(--management-accent);
  letter-spacing: 0.06em;
}
.release-page .management-header__title {
  font-size: 26px;
  letter-spacing: 0.04em;
}
.release-page .management-header__description {
  margin-top: 2px;
}
.release-current {
  padding: 22px 24px 18px;
  background: linear-gradient(
    110deg,
    color-mix(in srgb, var(--management-accent) 6%, var(--management-surface)),
    var(--management-surface) 65%
  );
  border: 1px solid color-mix(in srgb, var(--management-accent) 16%, var(--management-border));
  border-radius: 12px;
  color: var(--management-text);
  animation: release-enter 200ms ease-out;
}
.release-current__label {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--management-accent);
  font-size: 12px;
  font-weight: 600;
}
.release-current__platform {
  padding-left: 12px;
  border-left: 1px solid var(--management-border);
  color: var(--management-muted);
  font-weight: 400;
}
.release-current__version {
  margin: 10px 0 8px;
  font-size: 36px;
  font-weight: 650;
  line-height: 1.2;
  letter-spacing: -0.5px;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}
.release-current__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  font-size: 12px;
  color: var(--management-muted);
  font-variant-numeric: tabular-nums;
}
.release-current__notes {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--management-border);
}
.release-current__notes h3 {
  margin: 0 0 10px;
  font-size: 14px;
  font-weight: 600;
}
.release-current__notes ul {
  margin: 0;
  padding-left: 18px;
}
.release-current__notes li {
  padding: 4px 0 4px 4px;
  line-height: 1.8;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.release-current__notes li::marker {
  color: var(--management-muted);
}
.release-current__notes p {
  margin: 0;
  color: var(--management-muted);
  font-size: 13px;
}
.release-current__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  margin-top: 18px;
}
.release-current__empty {
  display: flex;
  align-items: flex-start;
  flex-direction: column;
  gap: 8px;
  min-height: 64px;
  justify-content: center;
}
.release-current__empty p {
  color: var(--management-muted);
  margin: 0;
  font-size: 12px;
}
.release-list :deep(.n-card-header) {
  padding: 16px 20px 12px;
}
.release-list :deep(.n-card__content) {
  padding: 0 20px 16px;
}
.release-list :deep(.n-data-table-td) {
  padding-top: 10px;
  padding-bottom: 10px;
}
.release-search {
  display: grid;
  grid-template-columns: minmax(180px, 360px) minmax(120px, 160px) minmax(140px, 180px) auto;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--management-border);
}
.release-search :deep(.n-form-item) {
  min-width: 0;
}
.release-search .search-actions {
  grid-column: auto;
  justify-content: flex-start;
  padding: 0;
  border: none;
}
.release-table-footer {
  color: var(--management-muted);
  font-size: 11px;
  padding-top: 12px;
}
:deep(.release-table-version),
:deep(.release-table-time) {
  display: flex;
  flex-direction: column;
  gap: 3px;
  font-variant-numeric: tabular-nums;
}
:deep(.release-table-version strong) {
  font-weight: 600;
  overflow-wrap: anywhere;
}
:deep(.release-table-version span),
:deep(.release-table-time span:last-child) {
  color: var(--management-muted);
  font-size: 11px;
}
:deep(.release-table-package) {
  font-size: 12px;
}
@keyframes release-enter {
  from {
    opacity: 0;
    transform: translateY(3px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
@media (max-width: 959px) {
  .release-search {
    grid-template-columns: minmax(0, 1fr) 180px;
  }
  .release-search .search-actions {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }
}
@media (max-width: 639px) {
  .release-current {
    padding: 16px;
  }
  .release-list :deep(.n-card-header) {
    padding: 14px 16px 12px;
  }
  .release-list :deep(.n-card__content) {
    padding: 0 16px 14px;
  }
  .release-search {
    grid-template-columns: minmax(0, 1fr);
    gap: 10px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .release-current {
    animation: none;
  }
}
</style>
