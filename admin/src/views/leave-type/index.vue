<script setup lang="ts">
import { computed, h, nextTick, reactive, ref } from 'vue';
import { NButton, NSpace, NTag } from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { queryLeaveTypes, saveLeaveType, deleteLeaveType } from '@/service/api/management';
import type { LeaveTypeRecord } from '@/service/api/management';

const appStore = useAppStore();
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
const formatTime = (value: string) => new Date(value).toLocaleString('zh-CN', { hour12: false });
const columns: DataTableColumns<LeaveTypeRecord> = [
  { key: 'id', title: 'ID', width: 70, align: 'center' },
  { key: 'name', title: '原因名称', width: 260, ellipsis: { tooltip: true } },
  { key: 'sortOrder', title: '排序值', width: 100, align: 'center' },
  {
    key: 'isEnabled',
    title: '状态',
    width: 90,
    align: 'center',
    render: row =>
      h(NTag, { type: row.isEnabled ? 'success' : 'default', bordered: false }, () => (row.isEnabled ? '启用' : '停用'))
  },
  { key: 'createdAt', title: '创建时间', width: 180, render: row => formatTime(row.createdAt) },
  { key: 'updatedAt', title: '更新时间', width: 180, render: row => formatTime(row.updatedAt) },
  {
    key: 'actions',
    title: '操作',
    width: 150,
    fixed: 'right',
    render: row =>
      h(NSpace, {}, () => [
        h(NButton, { size: 'small', onClick: () => openForm(row) }, () => '编辑'),
        h(NButton, { size: 'small', type: 'error', ghost: true, onClick: () => confirmDelete(row) }, () => '删除')
      ])
  }
];
reload();
</script>

<template>
  <div class="flex-col-stretch gap-16px">
    <NCard title="查询条件" :bordered="false" size="small" class="card-wrapper">
      <NForm :show-feedback="false" label-placement="top" class="search-form">
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
        <NSpace class="search-actions">
          <NButton type="primary" @click="applySearch">查询</NButton>
          <NButton @click="resetSearch">重置</NButton>
        </NSpace>
      </NForm>
    </NCard>
    <NCard title="请假原因管理" :bordered="false" size="small" class="card-wrapper">
      <template #header-extra>
        <NSpace align="center">
          <span class="record-count">共 {{ total }} 条</span>
          <NButton type="primary" @click="openForm()">新增原因</NButton>
        </NSpace>
      </template>
      <NDataTable
        :bordered="false"
        striped
        table-layout="fixed"
        :columns="columns"
        :data="rows"
        remote
        :loading="loading"
        :row-key="row => row.id"
        :pagination="pagination"
        :scroll-x="1030"
      />
    </NCard>
    <NDrawer
      v-model:show="drawerVisible"
      :width="appStore.isMobile ? '100%' : 480"
      :close-on-esc="!saving"
      :mask-closable="!saving"
    >
      <NDrawerContent
        :title="editingId === undefined ? '新增请假原因' : '编辑请假原因'"
        :closable="!saving"
        :native-scrollbar="false"
        :body-content-style="{ padding: '24px' }"
      >
        <NForm ref="formRef" :model="model" :rules="rules" label-placement="top" class="edit-form">
          <NFormItem label="原因名称" path="name">
            <NInput v-model:value="model.name" :maxlength="32" show-count placeholder="请输入原因名称" />
          </NFormItem>
          <NFormItem label="排序值" path="sortOrder">
            <NInputNumber v-model:value="model.sortOrder" :min="0" :max="4294967295" :precision="0" class="w-full" />
          </NFormItem>
          <NFormItem label="启用状态" path="isEnabled" :show-feedback="false">
            <div class="status-control">
              <NSwitch v-model:value="model.isEnabled" />
              <span>{{ model.isEnabled ? '启用' : '停用' }}</span>
            </div>
          </NFormItem>
        </NForm>
        <div class="drawer-actions">
          <NButton class="drawer-action-button" :disabled="saving" @click="drawerVisible = false">取消</NButton>
          <NButton class="drawer-action-button" type="primary" :loading="saving" @click="submit">保存</NButton>
        </div>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>

<style scoped>
.edit-form :deep(.n-form-item-label) {
  font-weight: 500;
}

.status-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.drawer-actions {
  display: flex;
  width: 100%;
  margin-top: 24px;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
}

.drawer-action-button {
  min-width: 88px;
}

.search-form {
  display: grid;
  grid-template-columns: minmax(180px, 280px) minmax(140px, 180px) auto;
  align-items: end;
  gap: 16px;
}

.search-actions {
  padding-bottom: 2px;
}

.record-count {
  color: var(--n-text-color-3);
  font-size: 13px;
}

@media (max-width: 1100px) {
  .search-form {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 600px) {
  .search-form {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
