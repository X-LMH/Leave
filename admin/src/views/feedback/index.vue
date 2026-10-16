<script setup lang="ts">
import { computed, h, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { NButton, NSpace, NTag, useThemeVars } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import {
  queryFeedbacks,
  getFeedback,
  deleteFeedback,
  updateFeedbackHandling,
  getFeedbackAttachment
} from '@/service/api/feedback';
import type { FeedbackRecord, FeedbackDetail } from '@/service/api/feedback';

const appStore = useAppStore();
const themeVars = useThemeVars();
const themeStyle = computed(() => ({
  '--management-accent': themeVars.value.primaryColor,
  '--management-text': themeVars.value.textColor1,
  '--management-muted': themeVars.value.textColor3,
  '--management-border': themeVars.value.dividerColor,
  '--management-surface': themeVars.value.cardColor,
  '--management-inset': themeVars.value.tableHeaderColor
}));
const emptySearch = () => ({
  keyword: '',
  studentId: '',
  isAnonymous: null as boolean | null,
  isHandled: null as boolean | null
});
const search = reactive(emptySearch());
const applied = reactive(emptySearch());
const anonymousOptions = [
  { label: '匿名反馈', value: 'anonymous' },
  { label: '实名反馈', value: 'identified' }
];
const handlingOptions = [
  { label: '未处理', value: 'unhandled' },
  { label: '已处理', value: 'handled' }
];
const searchHandling = computed({
  get: () => (search.isHandled === null ? null : search.isHandled ? 'handled' : 'unhandled'),
  set: (value: string | null) => {
    search.isHandled = value === null ? null : value === 'handled';
  }
});
const searchAnonymous = computed({
  get: () => (search.isAnonymous === null ? null : search.isAnonymous ? 'anonymous' : 'identified'),
  set: (value: string | null) => {
    search.isAnonymous = value === null ? null : value === 'anonymous';
  }
});
const rows = ref<FeedbackRecord[]>([]);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const loading = ref(false);
let querySequence = 0;
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
const drawerVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<FeedbackDetail | null>(null);
let detailSequence = 0;
const handling = ref(false);
const attachmentPreviews = ref<{ index: number; url: string; failed: boolean }[]>([]);
const attachmentsLoading = ref(false);

function clearAttachmentPreviews() {
  attachmentPreviews.value.forEach(item => {
    if (item.url) URL.revokeObjectURL(item.url);
  });
  attachmentPreviews.value = [];
}
function invalidateDetail() {
  detailSequence++;
  clearAttachmentPreviews();
  attachmentsLoading.value = false;
  detail.value = null;
  detailLoading.value = false;
}
watch(drawerVisible, visible => {
  if (!visible) invalidateDetail();
});
onBeforeUnmount(invalidateDetail);

async function loadAttachments(value: FeedbackDetail, sequence: number) {
  attachmentsLoading.value = true;
  try {
    const images = await Promise.all(
      value.attachments.map(async item => {
        try {
          const blob = await getFeedbackAttachment(value.id, item.index);
          return { index: item.index, blob };
        } catch {
          return { index: item.index, blob: null };
        }
      })
    );
    // 全部请求结束且详情仍有效时才创建 URL，避免关闭期间留下无人持有的资源。
    if (sequence !== detailSequence || !drawerVisible.value) return;
    attachmentPreviews.value = images.map(item => ({
      index: item.index,
      url: item.blob ? URL.createObjectURL(item.blob) : '',
      failed: item.blob === null
    }));
  } finally {
    if (sequence === detailSequence) attachmentsLoading.value = false;
  }
}

async function toggleHandling() {
  if (!detail.value || handling.value) return;
  const id = detail.value.id;
  const sequence = detailSequence;
  handling.value = true;
  try {
    const { data, error } = await updateFeedbackHandling(id, !detail.value.isHandled);
    if (error) return;
    if (sequence === detailSequence && drawerVisible.value) detail.value = data;
    window.$message?.success('处理状态已更新');
    await reload();
  } catch (error) {
    window.$message?.error(error instanceof Error ? error.message : '更新失败，请重试');
  } finally {
    handling.value = false;
  }
}

async function reload(): Promise<void> {
  const sequence = ++querySequence;
  loading.value = true;
  try {
    const { data, error } = await queryFeedbacks({
      ...applied,
      page: page.value,
      pageSize: pageSize.value
    });
    if (sequence !== querySequence) return;
    if (error) {
      rows.value = [];
      total.value = 0;
      return;
    }
    total.value = data.total;
    const lastPage = Math.max(1, Math.ceil(data.total / pageSize.value));
    if (page.value > lastPage) {
      page.value = lastPage;
      await reload();
      return;
    }
    rows.value = data.items;
  } catch (error) {
    if (sequence === querySequence) window.$message?.error(error instanceof Error ? error.message : '加载失败，请重试');
  } finally {
    if (sequence === querySequence) loading.value = false;
  }
}
function applySearch() {
  Object.assign(applied, search);
  page.value = 1;
  void reload();
}
function resetSearch() {
  Object.assign(search, emptySearch());
  applySearch();
}
async function openDetail(row: FeedbackRecord) {
  clearAttachmentPreviews();
  const sequence = ++detailSequence;
  detail.value = null;
  detailLoading.value = true;
  drawerVisible.value = true;
  try {
    const { data, error } = await getFeedback(row.id);
    if (sequence !== detailSequence) return;
    if (error) {
      drawerVisible.value = false;
      return;
    }
    detail.value = data;
    void loadAttachments(data, sequence);
  } catch (error) {
    if (sequence !== detailSequence) return;
    drawerVisible.value = false;
    window.$message?.error(error instanceof Error ? error.message : '加载详情失败，请重试');
  } finally {
    if (sequence === detailSequence) detailLoading.value = false;
  }
}
function confirmDelete(row: FeedbackRecord) {
  window.$dialog?.warning({
    title: '删除确认',
    content: `确认删除反馈 #${row.id} 吗？删除后该反馈将不再显示在列表中。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const { error } = await deleteFeedback(row.id);
        if (error) return false;
        await reload();
        window.$message?.success('删除成功');
        return true;
      } catch (error) {
        window.$message?.error(error instanceof Error ? error.message : '删除失败，请重试');
        return false;
      }
    }
  });
}
function submitter(row: FeedbackRecord) {
  return row.isAnonymous ? '匿名' : row.studentId || '—';
}
function formatTime(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}
function renderTime(value: string) {
  const date = new Date(value);
  return h('div', { class: 'table-time' }, [
    h('span', {}, date.toLocaleDateString('zh-CN')),
    h('span', { class: 'table-time__clock' }, date.toLocaleTimeString('zh-CN', { hour12: false }))
  ]);
}
const columns: DataTableColumns<FeedbackRecord> = [
  { key: 'id', title: 'ID', width: 70, align: 'center', className: 'table-id' },
  { key: 'studentId', title: '提交者学号', width: 155, render: submitter },
  {
    key: 'isAnonymous',
    title: '匿名状态',
    width: 100,
    align: 'center',
    render: row =>
      h(NTag, { type: row.isAnonymous ? 'default' : 'info', bordered: false }, () =>
        row.isAnonymous ? '匿名' : '实名'
      )
  },
  {
    key: 'content',
    title: '反馈内容',
    width: 360,
    ellipsis: { tooltip: true }
  },
  {
    key: 'isHandled',
    title: '处理状态',
    width: 100,
    align: 'center',
    render: row =>
      h(NTag, { type: row.isHandled ? 'success' : 'warning', bordered: false }, () =>
        row.isHandled ? '已处理' : '未处理'
      )
  },
  {
    key: 'createdAt',
    title: '提交时间',
    width: 145,
    render: row => renderTime(row.createdAt)
  },
  {
    key: 'actions',
    title: '操作',
    width: 150,
    fixed: 'right',
    render: row =>
      h(NSpace, {}, () => [
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            secondary: true,
            onClick: () => openDetail(row)
          },
          () => '详情'
        ),
        h(
          NButton,
          {
            size: 'small',
            type: 'error',
            quaternary: true,
            onClick: () => confirmDelete(row)
          },
          () => '删除'
        )
      ])
  }
];
void reload();
</script>

<template>
  <div class="management-page management-page--feedback" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon" aria-hidden="true"><SvgIcon icon="mdi:message-text-outline" /></span>
        <div>
          <p class="management-header__eyebrow">用户服务 / 意见反馈</p>
          <h1 class="management-header__title">意见反馈管理</h1>
          <p class="management-header__description">查看用户提交的意见与建议，支持匿名反馈。</p>
        </div>
      </div>
      <div class="management-header__count" aria-live="polite">
        <span class="management-header__count-label">当前查询结果</span>
        <span>
          <strong>{{ total }}</strong>
          条反馈
        </span>
      </div>
    </header>
    <NCard
      :bordered="false"
      size="small"
      class="management-card management-card--search management-card--compact-search"
    >
      <template #header>
        <div class="section-heading">
          <SvgIcon icon="mdi:filter-variant" aria-hidden="true" />
          <span>筛选条件</span>
        </div>
      </template>
      <NForm :show-feedback="false" label-placement="top" class="search-form search-form--compact">
        <NFormItem label="反馈内容">
          <NInput v-model:value="search.keyword" clearable placeholder="请输入内容关键词" @keyup.enter="applySearch" />
        </NFormItem>
        <NFormItem label="提交者学号">
          <NInput v-model:value="search.studentId" clearable placeholder="请输入学号" @keyup.enter="applySearch" />
        </NFormItem>
        <NFormItem label="匿名状态">
          <NSelect v-model:value="searchAnonymous" :options="anonymousOptions" clearable placeholder="全部反馈" />
        </NFormItem>
        <NFormItem label="处理状态">
          <NSelect v-model:value="searchHandling" :options="handlingOptions" clearable placeholder="全部状态" />
        </NFormItem>
        <div class="search-actions">
          <NButton type="primary" :loading="loading" @click="applySearch">
            <template #icon><SvgIcon icon="mdi:magnify" /></template>
            查询
          </NButton>
          <NButton @click="resetSearch">
            <template #icon><SvgIcon icon="mdi:refresh" /></template>
            重置
          </NButton>
        </div>
      </NForm>
    </NCard>
    <NCard :bordered="false" size="small" class="management-card management-card--table">
      <template #header>
        <div class="section-heading">
          <span>反馈列表</span>
          <span class="section-heading__badge">{{ total }}</span>
        </div>
      </template>
      <template #header-extra>
        <NSpace align="center">
          <NButton :loading="loading" @click="reload">
            <template #icon><SvgIcon icon="mdi:refresh" /></template>
            刷新
          </NButton>
        </NSpace>
      </template>
      <NDataTable
        :bordered="false"
        table-layout="fixed"
        :columns="columns"
        :data="rows"
        remote
        :loading="loading"
        :row-key="row => row.id"
        :pagination="pagination"
        :scroll-x="1100"
      >
        <template #empty>
          <div class="table-empty">
            <SvgIcon icon="mdi:message-text-outline" aria-hidden="true" />
            <strong>暂无反馈数据</strong>
            <span>可以调整筛选条件，查看其他反馈。</span>
            <NButton size="small" @click="resetSearch">重置筛选</NButton>
          </div>
        </template>
      </NDataTable>
    </NCard>
    <NDrawer v-model:show="drawerVisible" :width="appStore.isMobile ? '100%' : 600">
      <NDrawerContent
        class="management-drawer feedback-drawer"
        :style="themeStyle"
        title="意见反馈详情"
        closable
        :native-scrollbar="false"
      >
        <NSpin :show="detailLoading">
          <template v-if="detail">
            <div class="configuration-summary">
              <span class="configuration-summary__icon" aria-hidden="true">
                <SvgIcon icon="mdi:message-text-outline" />
              </span>
              <div>
                <strong>反馈 #{{ detail.id }}</strong>
                <p>{{ detail.isAnonymous ? '匿名反馈' : '实名反馈' }}</p>
              </div>
            </div>
            <section class="detail-section">
              <NDescriptions class="feedback-metadata" :column="1" label-placement="left">
                <NDescriptionsItem label="提交者学号">
                  {{ submitter(detail) }}
                </NDescriptionsItem>
                <NDescriptionsItem label="是否匿名">
                  {{ detail.isAnonymous ? '是' : '否' }}
                </NDescriptionsItem>
                <NDescriptionsItem label="提交时间">
                  {{ formatTime(detail.createdAt) }}
                </NDescriptionsItem>
                <NDescriptionsItem label="处理状态">
                  <NTag :type="detail.isHandled ? 'success' : 'warning'" :bordered="false">
                    {{ detail.isHandled ? '已处理' : '未处理' }}
                  </NTag>
                </NDescriptionsItem>
                <NDescriptionsItem label="更新时间">
                  {{ formatTime(detail.updatedAt) }}
                </NDescriptionsItem>
              </NDescriptions>
            </section>
            <section class="detail-section">
              <div class="detail-section__heading"><h2>反馈内容</h2></div>
              <p class="feedback-content">{{ detail.content }}</p>
            </section>
            <section class="detail-section">
              <div class="detail-section__heading"><h2>附件图片</h2></div>
              <NSpin :show="attachmentsLoading">
                <p v-if="detail.attachments.length === 0" class="feedback-attachment-empty">无附件</p>
                <NImageGroup v-else>
                  <div class="feedback-attachments">
                    <div v-for="item in attachmentPreviews" :key="item.index" class="feedback-attachment">
                      <NImage
                        v-if="!item.failed"
                        :src="item.url"
                        :width="140"
                        :height="110"
                        object-fit="cover"
                        :alt="`附件 ${item.index + 1}`"
                      />
                      <span v-else>附件 {{ item.index + 1 }} 加载失败</span>
                    </div>
                  </div>
                </NImageGroup>
                <NButton
                  v-if="attachmentPreviews.some(item => item.failed)"
                  size="small"
                  :disabled="attachmentsLoading"
                  @click="
                    clearAttachmentPreviews();
                    loadAttachments(detail, detailSequence);
                  "
                >
                  重新加载附件
                </NButton>
              </NSpin>
            </section>
          </template>
        </NSpin>
        <template #footer>
          <div class="drawer-footer">
            <span class="drawer-footer__hint">
              {{ detail?.isAnonymous ? '匿名反馈不保留提交者学号' : '用户提交的原始反馈内容' }}
            </span>
            <NSpace>
              <NButton
                v-if="detail"
                type="primary"
                :loading="handling"
                :disabled="detailLoading || handling"
                @click="toggleHandling"
              >
                {{ detail.isHandled ? '恢复未处理' : '标记已处理' }}
              </NButton>
              <NButton @click="drawerVisible = false">关闭</NButton>
            </NSpace>
          </div>
        </template>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';

@container management-search (min-width: 1000px) {
  .management-page--feedback .search-form--compact {
    grid-template-columns: minmax(160px, 1fr) minmax(150px, 1fr) repeat(2, minmax(130px, 0.8fr));
  }
}

.feedback-metadata {
  :deep(.n-descriptions-table-content__label),
  :deep(.n-descriptions-table-content__content) {
    display: inline-block;
    vertical-align: middle;
  }
}

.feedback-attachments {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  min-height: 40px;
}
.feedback-attachment {
  overflow: hidden;
  border-radius: 8px;
  color: var(--management-muted);
}
.feedback-attachment-empty {
  color: var(--management-muted);
}
.feedback-content {
  margin: 0;
  color: var(--management-text);
  font-size: 14px;
  line-height: 1.9;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
