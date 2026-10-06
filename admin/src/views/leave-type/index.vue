<script setup lang="ts">
import { computed, h, nextTick, reactive, ref } from 'vue';
import { NButton, NSpace, NTag, useThemeVars } from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { queryLeaveTypes, saveLeaveType, deleteLeaveType } from '@/service/api/management';
import type { LeaveTypeRecord } from '@/service/api/management';

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
const rows = ref<LeaveTypeRecord[]>([]);
const search = reactive({ name: '', isEnabled: null as boolean | null });
const applied = reactive({ ...search });
const statusOptions = [
  { label: '启用', value: 'enabled' },
  { label: '停用', value: 'disabled' }
];
const searchStatus = computed({
  get: () => (search.isEnabled === null ? null : search.isEnabled ? 'enabled' : 'disabled'),
  set: (value: string | null) => {
    search.isEnabled = value === null ? null : value === 'enabled';
  }
});
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
const editingId = ref<number>();
const saving = ref(false);
const formRef = ref<FormInst | null>(null);
const emptyModel = () => ({ name: '', sortOrder: 0 as number | null, isEnabled: true });
const model = reactive(emptyModel());
const rules: FormRules = {
  name: {
    required: true,
    validator: (_rule, value: string) => Boolean(value?.trim()) && value.trim().length <= 32,
    message: '请输入32字符以内的原因名称',
    trigger: ['blur', 'input']
  },
  sortOrder: {
    required: true,
    validator: (_rule, value: number) => Number.isSafeInteger(value) && value >= 0 && value <= 4294967295,
    message: '请输入非负整数',
    trigger: ['blur', 'change']
  }
};
async function reload(): Promise<void> {
  const sequence = ++querySequence;
  loading.value = true;
  try {
    const { data, error } = await queryLeaveTypes({ ...applied, page: page.value, pageSize: pageSize.value });
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
  Object.assign(search, { name: '', isEnabled: null as boolean | null });
  applySearch();
}
async function openForm(row?: LeaveTypeRecord) {
  if (saving.value) return;
  editingId.value = row?.id;
  Object.assign(model, emptyModel());
  if (row) {
    for (const key of Object.keys(emptyModel()) as (keyof typeof model)[]) {
      Object.assign(model, { [key]: row[key] });
    }
  }
  drawerVisible.value = true;
  await nextTick();
  formRef.value?.restoreValidation();
}
async function submit() {
  if (saving.value) return;
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  if (saving.value) return;
  saving.value = true;
  try {
    const { error } = await saveLeaveType({ ...model, sortOrder: model.sortOrder! }, editingId.value);
    if (error) return;
    drawerVisible.value = false;
    await reload();
    window.$message?.success(editingId.value === undefined ? '新增成功' : '修改成功');
  } finally {
    saving.value = false;
  }
}
function confirmDelete(row: LeaveTypeRecord) {
  window.$dialog?.warning({
    title: '删除确认',
    content: `确认删除“${row.name}”吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await deleteLeaveType(row.id);
      if (error) return false;
      await reload();
      window.$message?.success('删除成功');
      return true;
    }
  });
}
function renderTime(value: string) {
  const date = new Date(value);
  return h('div', { class: 'table-time' }, [
    h('span', {}, date.toLocaleDateString('zh-CN')),
    h('span', { class: 'table-time__clock' }, date.toLocaleTimeString('zh-CN', { hour12: false }))
  ]);
}
const columns: DataTableColumns<LeaveTypeRecord> = [
  { key: 'id', title: 'ID', width: 70, align: 'center', className: 'table-id' },
  { key: 'name', title: '原因名称', width: 220, ellipsis: { tooltip: true } },
  { key: 'sortOrder', title: '排序值', width: 100, align: 'center' },
  {
    key: 'isEnabled',
    title: '状态',
    width: 90,
    align: 'center',
    render: row =>
      h(NTag, { type: row.isEnabled ? 'success' : 'default', bordered: false }, () => (row.isEnabled ? '启用' : '停用'))
  },
  { key: 'createdAt', title: '创建时间', width: 130, render: row => renderTime(row.createdAt) },
  { key: 'updatedAt', title: '更新时间', width: 130, render: row => renderTime(row.updatedAt) },
  {
    key: 'actions',
    title: '操作',
    width: 150,
    fixed: 'right',
    render: row =>
      h(NSpace, {}, () => [
        h(
          NButton,
          { size: 'small', type: 'primary', secondary: true, class: 'table-edit-button', onClick: () => openForm(row) },
          () => '编辑'
        ),
        h(NButton, { size: 'small', type: 'error', quaternary: true, onClick: () => confirmDelete(row) }, () => '删除')
      ])
  }
];
reload();
</script>

<template>
  <div class="management-page management-page--leave-type" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon" aria-hidden="true"><SvgIcon icon="mdi:format-list-bulleted" /></span>
        <div>
          <p class="management-header__eyebrow">教务管理 / 基础配置</p>
          <h1 class="management-header__title">请假原因管理</h1>
          <p class="management-header__description">维护请假原因选项、显示顺序与启用状态。</p>
        </div>
      </div>
      <div class="management-header__count" aria-live="polite">
        <span class="management-header__count-label">当前查询结果</span>
        <span>
          <strong>{{ total }}</strong>
          项原因
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
        <NFormItem label="原因名称">
          <NInput v-model:value="search.name" clearable placeholder="请输入原因名称" @keyup.enter="applySearch" />
        </NFormItem>
        <NFormItem label="状态">
          <NSelect
            v-model:value="searchStatus"
            :options="statusOptions"
            clearable
            placeholder="全部状态"
            class="w-full"
          />
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
          <span>原因列表</span>
          <span class="section-heading__badge">{{ total }}</span>
        </div>
      </template>
      <template #header-extra>
        <NSpace align="center">
          <NButton type="primary" @click="openForm()">
            <template #icon><SvgIcon icon="mdi:plus" /></template>
            新增原因
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
        :scroll-x="890"
      >
        <template #empty>
          <div class="table-empty">
            <SvgIcon icon="mdi:format-list-bulleted" aria-hidden="true" />
            <strong>暂无原因数据</strong>
            <span>可以调整筛选条件，或新增原因。</span>
            <NButton size="small" @click="resetSearch">重置筛选</NButton>
          </div>
        </template>
      </NDataTable>
    </NCard>
    <NDrawer
      v-model:show="drawerVisible"
      :width="appStore.isMobile ? '100%' : 560"
      :close-on-esc="!saving"
      :mask-closable="!saving"
    >
      <NDrawerContent
        class="management-drawer"
        :style="themeStyle"
        :title="editingId === undefined ? '新增请假原因' : '编辑请假原因'"
        :closable="!saving"
        :native-scrollbar="false"
      >
        <div class="configuration-summary">
          <span class="configuration-summary__icon" aria-hidden="true">
            <SvgIcon icon="mdi:format-list-bulleted" />
          </span>
          <div>
            <strong>{{ editingId === undefined ? '新增原因' : '编辑原因' }}</strong>
            <p>设置原因名称、排序值与启用状态。</p>
          </div>
        </div>
        <NForm
          ref="formRef"
          :model="model"
          :rules="rules"
          label-placement="top"
          class="edit-form edit-form--configuration"
        >
          <div class="form-section-heading">
            <span>01</span>
            <h2>基础信息</h2>
          </div>
          <NFormItem label="原因名称" path="name">
            <NInput v-model:value="model.name" :maxlength="32" show-count placeholder="请输入原因名称" />
          </NFormItem>
          <NFormItem label="排序值（越小越靠前）" path="sortOrder">
            <NInputNumber v-model:value="model.sortOrder" :min="0" :max="4294967295" :precision="0" class="w-full" />
          </NFormItem>
          <div class="form-section-heading">
            <span>02</span>
            <h2>启用设置</h2>
          </div>
          <NFormItem label="启用状态" path="isEnabled" :show-feedback="false">
            <div class="status-control status-control--configuration">
              <div class="status-control__copy">
                <strong>{{ model.isEnabled ? '已启用' : '已停用' }}</strong>
                <span>控制该原因是否可用于新请假申请。</span>
              </div>
              <NSwitch v-model:value="model.isEnabled" aria-label="启用状态" />
            </div>
          </NFormItem>
        </NForm>
        <template #footer>
          <div class="drawer-footer">
            <span class="drawer-footer__hint">确认信息后保存</span>
            <div class="drawer-actions">
              <NButton :disabled="saving" @click="drawerVisible = false">取消</NButton>
              <NButton type="primary" :loading="saving" @click="submit">保存</NButton>
            </div>
          </div>
        </template>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';
</style>
