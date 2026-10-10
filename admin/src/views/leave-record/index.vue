<script setup lang="ts">
import { computed, h, nextTick, reactive, ref } from 'vue';
import { NButton, useThemeVars } from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import { useThemeStore } from '@/store/modules/theme';
import { useAppStore } from '@/store/modules/app';
import { getLeaveRecord, queryLeaveRecords, saveLeaveRecord } from '@/service/api/leave-record';
import type { LeaveContent, LeaveQuery, LeaveRecord, LeaveRecordListItem } from '@/service/api/leave-record';
import { getClassOptions, getLeaveTypeOptions } from '@/service/api/options';
import type { ClassOption, LeaveTypeOption } from '@/service/api/options';

function formatTime(value: number | null) {
  return value === null ? '—' : new Date(value).toLocaleString('zh-CN', { hour12: false });
}
function genderLabel(value: LeaveRecord['gender']) {
  if (value === 'male') return '男';
  if (value === 'female') return '女';
  return '—';
}
function durationLabel(start: number | null, end: number | null) {
  if (start === null || end === null || end <= start) return '—';
  const hours = Math.ceil((end - start) / 3600000);
  return `${Math.floor(hours / 24)}天${hours % 24}小时`;
}

const appStore = useAppStore();
const themeVars = useThemeVars();
const themeStore = useThemeStore();
const themeStyle = computed(() => ({
  '--management-accent': themeVars.value.primaryColor,
  '--management-reason-sick': themeStore.darkMode ? '#a9c2db' : '#526f8c',
  '--management-reason-personal': themeStore.darkMode ? '#d7b98e' : '#96734c',
  '--management-reason-other': themeStore.darkMode ? '#c4b6cf' : '#796987',
  '--management-away-color': themeStore.darkMode ? '#a9c9e8' : '#426b91',
  '--management-text': themeVars.value.textColor1,
  '--management-muted': themeVars.value.textColor3,
  '--management-border': themeVars.value.dividerColor,
  '--management-surface': themeVars.value.cardColor,
  '--management-inset': themeVars.value.tableHeaderColor
}));
function renderTime(value: number | null, empty = '—') {
  if (value === null) return h('span', { class: 'table-muted' }, empty);
  const date = new Date(value);
  return h('div', { class: 'table-time' }, [
    h('span', {}, date.toLocaleDateString('zh-CN')),
    h('span', { class: 'table-time__clock' }, date.toLocaleTimeString('zh-CN', { hour12: false }))
  ]);
}
const emptySearch = (): LeaveQuery => ({
  studentId: '',
  name: '',
  classSnapshot: null,
  leaveTypeId: null,
  isLeaveSchool: null,
  dateRange: null
});
const search = reactive<LeaveQuery>(emptySearch());
const applied = reactive<LeaveQuery>(emptySearch());
const rows = ref<LeaveRecordListItem[]>([]);
const loading = ref(false);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
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
async function reload(): Promise<void> {
  const sequence = ++querySequence;
  loading.value = true;
  try {
    const { data, error } = await queryLeaveRecords({ ...applied, page: page.value, pageSize: pageSize.value });
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
  Object.assign(applied, search, { dateRange: search.dateRange ? [...search.dateRange] : null });
  page.value = 1;
  void reload();
}
function resetSearch() {
  Object.assign(search, emptySearch());
  applySearch();
}
const classes = ref<ClassOption[]>([]);
const leaveTypes = ref<LeaveTypeOption[]>([]);
async function loadOptions() {
  const [classResult, typeResult] = await Promise.all([getClassOptions(), getLeaveTypeOptions()]);
  if (!classResult.error) classes.value = classResult.data;
  if (!typeResult.error) leaveTypes.value = typeResult.data;
  return !typeResult.error;
}
const classOptions = computed(() =>
  classes.value.map(row => ({
    label: `${row.college} / ${row.major} / ${row.className}`,
    value: JSON.stringify({ college: row.college, major: row.major, className: row.className })
  }))
);
const typeOptions = computed(() =>
  leaveTypes.value.map(row => ({
    label: `${row.name}${row.isEnabled ? '' : '（已停用）'}`,
    value: row.id
  }))
);
const selected = ref<LeaveRecord | null>(null);
const editTypeOptions = computed(() => {
  const current = selected.value;
  const options = leaveTypes.value.map(row => ({
    label: row.id === current?.leaveTypeId ? current.leaveTypeName : `${row.name}${row.isEnabled ? '' : '（已停用）'}`,
    value: row.id,
    disabled: !row.isEnabled && row.id !== current?.leaveTypeId
  }));
  if (current?.leaveTypeId && !options.some(row => row.value === current.leaveTypeId)) {
    options.push({ label: `${current.leaveTypeName}（已删除）`, value: current.leaveTypeId, disabled: false });
  }
  return options;
});
void loadOptions();
void reload();
const schoolOptions = [
  { label: '离校', value: 'yes' },
  { label: '不离校', value: 'no' }
];
const searchSchool = computed({
  get: () => (search.isLeaveSchool === null ? null : search.isLeaveSchool ? 'yes' : 'no'),
  set: (value: string | null) => {
    search.isLeaveSchool = value === null ? null : value === 'yes';
  }
});
function reasonClass(name: string) {
  // 真实名称可能包含适用对象后缀，例如“病假-本科生”。
  const reason = name.trim();
  if (reason.startsWith('病假')) return 'detail-reason--sick';
  if (reason.startsWith('事假')) return 'detail-reason--personal';
  if (reason.startsWith('其他')) return 'detail-reason--other';
  return '';
}
const drawerVisible = ref(false);
const editing = ref(false);
const saving = ref(false);
const opening = ref(false);
const formRef = ref<FormInst | null>(null);
const model = reactive<LeaveContent>({
  leaveTypeId: null,
  leaveReason: '',
  startTime: null,
  endTime: null,
  affectedCourse: '',
  isLeaveSchool: false,
  travelWay: '',
  destination: '',
  appliedAt: null,
  approvedAt: null
});
const requiredTime = { required: true, type: 'number' as const, message: '请选择时间', trigger: 'change' };
const rules: FormRules = {
  leaveTypeId: { required: true, type: 'number', message: '请选择请假原因', trigger: 'change' },
  leaveReason: {
    required: true,
    validator: (_rule, value: string) => Boolean(value?.trim()),
    message: '请输入具体事由',
    trigger: ['blur', 'input']
  },
  startTime: requiredTime,
  endTime: [
    requiredTime,
    {
      validator: (_rule, value: number | null) => value !== null && model.startTime !== null && value > model.startTime,
      message: '结束时间必须晚于开始时间',
      trigger: 'change'
    }
  ],
  appliedAt: requiredTime,
  approvedAt: [
    requiredTime,
    {
      validator: (_rule, value: number | null) =>
        value !== null && model.appliedAt !== null && value >= model.appliedAt,
      message: '审批时间不得早于申请时间',
      trigger: 'change'
    }
  ],
  destination: {
    validator: (_rule, value: string) => !model.isLeaveSchool || Boolean(value?.trim()),
    message: '离校时请输入目的地',
    trigger: ['blur', 'input']
  }
};
async function openDrawer(row: { id: number }, edit: boolean) {
  if (opening.value || saving.value) return;
  opening.value = true;
  try {
    const { data, error } = await getLeaveRecord(row.id);
    if (error) return;
    if (edit && !(await loadOptions())) return;
    selected.value = data;
    Object.assign(model, {
      leaveTypeId: data.leaveTypeId,
      leaveReason: data.leaveReason,
      startTime: data.startTime,
      endTime: data.endTime,
      affectedCourse: data.affectedCourse,
      isLeaveSchool: data.isLeaveSchool,
      travelWay: data.travelWay,
      destination: data.destination,
      appliedAt: data.appliedAt,
      approvedAt: data.approvedAt
    });
    editing.value = edit;
    drawerVisible.value = true;
    await nextTick();
    formRef.value?.restoreValidation();
  } catch {
    window.$message?.error('加载请假记录失败');
  } finally {
    opening.value = false;
  }
}
async function submit() {
  if (saving.value || !selected.value || !formRef.value) return;
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  if (saving.value) return;
  saving.value = true;
  try {
    const { data, error } = await saveLeaveRecord(selected.value.id, {
      ...model,
      leaveReason: model.leaveReason.trim(),
      affectedCourse: model.affectedCourse.trim(),
      travelWay: model.travelWay.trim(),
      destination: model.destination.trim()
    });
    if (error) return;
    selected.value = data;
    editing.value = false;
    await reload();
    window.$message?.success('保存成功');
  } catch {
    window.$message?.error('保存失败，请重试');
  } finally {
    saving.value = false;
  }
}
const columns: DataTableColumns<LeaveRecordListItem> = [
  { key: 'studentId', title: '学号', width: 130, className: 'table-id' },
  {
    key: 'name',
    title: '姓名',
    width: 120,
    render: row => h('span', { class: 'table-person__name' }, row.name || '未完善资料')
  },
  { key: 'className', title: '班级', width: 160, ellipsis: { tooltip: true } },
  {
    key: 'leaveTypeId',
    title: '请假原因',
    width: 110,
    render: row => {
      const name = row.leaveTypeName || '—';
      return name === '—' ? name : h('span', { class: ['detail-reason', reasonClass(name)] }, name);
    }
  },
  { key: 'startTime', title: '开始时间', width: 130, render: row => renderTime(row.startTime) },
  { key: 'endTime', title: '结束时间', width: 130, render: row => renderTime(row.endTime) },
  { key: 'duration', title: '请假时长', width: 110, render: row => durationLabel(row.startTime, row.endTime) },
  {
    key: 'isLeaveSchool',
    title: '是否离校',
    width: 90,
    render: row =>
      h(
        'span',
        { class: ['leave-school-tag', row.isLeaveSchool && 'leave-school-tag--away'] },
        row.isLeaveSchool ? '离校' : '不离校'
      )
  },
  { key: 'appliedAt', title: '申请时间', width: 130, render: row => renderTime(row.appliedAt) },
  {
    key: 'actions',
    title: '操作',
    width: 100,
    align: 'left',
    fixed: 'right',
    render: row =>
      h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          secondary: true,
          class: 'table-detail-button',
          disabled: opening.value || saving.value,
          onClick: () => openDrawer(row, false)
        },
        () => '详情'
      )
  }
];
</script>

<template>
  <div class="management-page management-page--leave-record" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon" aria-hidden="true">
          <SvgIcon icon="mdi:clipboard-text-clock-outline" />
        </span>
        <div>
          <p class="management-header__eyebrow">教务管理 / 请假事务</p>
          <h1 class="management-header__title">请假记录管理</h1>
          <p class="management-header__description">查看学生请假安排，核对申请详情与行程信息。</p>
        </div>
      </div>
      <div class="management-header__count" aria-live="polite">
        <span class="management-header__count-label">当前查询结果</span>
        <span>
          <strong>{{ total }}</strong>
          条记录
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
        <NFormItem label="学号">
          <NInput v-model:value="search.studentId" clearable placeholder="请输入学号" @keyup.enter="applySearch" />
        </NFormItem>
        <NFormItem label="姓名">
          <NInput v-model:value="search.name" clearable placeholder="请输入姓名" @keyup.enter="applySearch" />
        </NFormItem>
        <NFormItem label="班级">
          <NSelect
            v-model:value="search.classSnapshot"
            :options="classOptions"
            filterable
            clearable
            placeholder="全部班级"
          />
        </NFormItem>
        <NFormItem label="请假原因">
          <NSelect v-model:value="search.leaveTypeId" :options="typeOptions" clearable placeholder="全部原因" />
        </NFormItem>
        <NFormItem label="是否离校">
          <NSelect v-model:value="searchSchool" :options="schoolOptions" clearable placeholder="全部" />
        </NFormItem>
        <NFormItem label="请假开始日期" class="search-field--date">
          <NDatePicker v-model:value="search.dateRange" type="daterange" clearable class="w-full" />
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
          <span>请假记录</span>
          <span class="section-heading__badge">{{ total }}</span>
        </div>
      </template>
      <template #header-extra>
        <span class="record-count">按申请时间倒序</span>
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
        :scroll-x="1200"
      >
        <template #empty>
          <div class="table-empty">
            <SvgIcon icon="mdi:clipboard-search-outline" aria-hidden="true" />
            <strong>没有找到匹配的请假记录</strong>
            <span>试试调整筛选条件，或重置后查看全部数据。</span>
            <NButton size="small" @click="resetSearch">重置筛选</NButton>
          </div>
        </template>
      </NDataTable>
    </NCard>
    <NDrawer
      v-model:show="drawerVisible"
      :width="appStore.isMobile ? '100%' : 760"
      :close-on-esc="!saving"
      :mask-closable="!saving"
    >
      <NDrawerContent
        class="management-drawer"
        :style="themeStyle"
        :title="editing ? '编辑请假记录' : '请假记录详情'"
        :closable="!saving"
        :native-scrollbar="false"
      >
        <template v-if="selected">
          <div class="drawer-summary">
            <NAvatar class="drawer-summary__avatar" :size="48" :img-props="{ alt: `${selected.name || '学生'}的头像` }">
              {{ selected.name?.slice(0, 1) || '学' }}
            </NAvatar>
            <div class="drawer-summary__identity">
              <strong>{{ selected.name || '未完善资料' }}</strong>
              <span>{{ selected.studentId }} · {{ selected.className || '未设置班级' }}</span>
            </div>
          </div>
          <div v-if="!editing" class="detail-highlights detail-highlights--record">
            <div>
              <span>记录创建时间</span>
              <strong>{{ formatTime(selected.createdAt) }}</strong>
            </div>
            <div>
              <span>请假原因</span>
              <span
                v-if="(selected.leaveTypeName || '—') !== '—'"
                class="detail-reason"
                :class="reasonClass(selected.leaveTypeName || '—')"
              >
                {{ selected.leaveTypeName || '—' }}
              </span>
              <strong v-else class="detail-value">—</strong>
            </div>
            <div>
              <span>请假时长</span>
              <strong>{{ durationLabel(selected.startTime, selected.endTime) }}</strong>
            </div>
          </div>
          <NDescriptions
            v-if="editing"
            :column="appStore.isMobile ? 1 : 2"
            class="detail-section"
            label-placement="top"
          >
            <template #header>
              <div class="detail-section__heading">
                <SvgIcon icon="mdi:account-school-outline" aria-hidden="true" />
                <span>学生与班级信息</span>
              </div>
            </template>
            <NDescriptionsItem label="学号">{{ selected.studentId }}</NDescriptionsItem>
            <NDescriptionsItem label="姓名">{{ selected.name || '—' }}</NDescriptionsItem>
            <NDescriptionsItem label="性别">{{ genderLabel(selected.gender) }}</NDescriptionsItem>
            <NDescriptionsItem label="班级">{{ selected.className || '—' }}</NDescriptionsItem>
            <NDescriptionsItem label="学院">{{ selected.college || '—' }}</NDescriptionsItem>
            <NDescriptionsItem label="专业">{{ selected.major || '—' }}</NDescriptionsItem>
          </NDescriptions>
          <NDescriptions
            v-if="editing"
            :column="appStore.isMobile ? 1 : 2"
            class="detail-section"
            label-placement="top"
          >
            <template #header>
              <div class="detail-section__heading">
                <SvgIcon icon="mdi:account-group-outline" aria-hidden="true" />
                <span>家长与辅导员信息</span>
              </div>
            </template>
            <NDescriptionsItem label="家长姓名">{{ selected.parentName || '—' }}</NDescriptionsItem>
            <NDescriptionsItem label="家长电话">
              <strong class="detail-value">{{ selected.parentPhone || '—' }}</strong>
            </NDescriptionsItem>
            <NDescriptionsItem label="辅导员">{{ selected.teacherName || '—' }}</NDescriptionsItem>
          </NDescriptions>
          <NForm v-if="editing" ref="formRef" :model="model" :rules="rules" label-placement="top" class="edit-form">
            <div class="form-section-heading">
              <span>01</span>
              <h2>请假内容</h2>
            </div>
            <NFormItem label="请假原因" path="leaveTypeId">
              <NSelect v-model:value="model.leaveTypeId" :options="editTypeOptions" />
            </NFormItem>
            <NFormItem label="请假时长">
              <NInput :value="durationLabel(model.startTime, model.endTime)" readonly />
            </NFormItem>
            <NFormItem label="开始时间" path="startTime">
              <NDatePicker v-model:value="model.startTime" type="datetime" clearable class="w-full" />
            </NFormItem>
            <NFormItem label="结束时间" path="endTime">
              <NDatePicker v-model:value="model.endTime" type="datetime" clearable class="w-full" />
            </NFormItem>
            <NFormItem label="具体事由" path="leaveReason" class="full-width">
              <NInput v-model:value="model.leaveReason" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
            </NFormItem>
            <NFormItem label="影响课程" class="full-width">
              <NInput v-model:value="model.affectedCourse" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" />
            </NFormItem>
            <div class="form-section-heading">
              <span>02</span>
              <h2>离校与行程</h2>
            </div>
            <NFormItem label="是否离校" class="full-width">
              <NSpace align="center">
                <NSwitch v-model:value="model.isLeaveSchool" />
                <span>{{ model.isLeaveSchool ? '离校' : '不离校' }}</span>
              </NSpace>
            </NFormItem>
            <template v-if="model.isLeaveSchool">
              <NFormItem label="交通方式"><NInput v-model:value="model.travelWay" /></NFormItem>
              <NFormItem label="目的地" path="destination"><NInput v-model:value="model.destination" /></NFormItem>
            </template>
            <div class="form-section-heading">
              <span>03</span>
              <h2>申请与审批时间</h2>
            </div>
            <NFormItem label="申请时间" path="appliedAt">
              <NDatePicker v-model:value="model.appliedAt" type="datetime" clearable class="w-full" />
            </NFormItem>
            <NFormItem label="审批时间" path="approvedAt">
              <NDatePicker v-model:value="model.approvedAt" type="datetime" clearable class="w-full" />
            </NFormItem>
          </NForm>
          <template v-else>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:clipboard-text-outline" aria-hidden="true" />
                  <span>请假内容</span>
                </div>
              </template>
              <NDescriptionsItem label="请假原因">
                <strong class="detail-value">{{ selected.leaveTypeName || '—' }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="请假时长">
                <strong class="detail-value">{{ durationLabel(selected.startTime, selected.endTime) }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="开始时间">
                <strong class="detail-value">{{ formatTime(selected.startTime) }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="结束时间">
                <strong class="detail-value">{{ formatTime(selected.endTime) }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="具体事由" :span="appStore.isMobile ? 1 : 2">
                <span class="detail-text">{{ selected.leaveReason || '—' }}</span>
              </NDescriptionsItem>
              <NDescriptionsItem label="影响课程" :span="appStore.isMobile ? 1 : 2">
                <span class="detail-text">{{ selected.affectedCourse || '—' }}</span>
              </NDescriptionsItem>
            </NDescriptions>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:account-school-outline" aria-hidden="true" />
                  <span>学生与班级信息</span>
                </div>
              </template>
              <NDescriptionsItem label="学号">{{ selected.studentId }}</NDescriptionsItem>
              <NDescriptionsItem label="姓名">{{ selected.name || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="性别">{{ genderLabel(selected.gender) }}</NDescriptionsItem>
              <NDescriptionsItem label="班级">{{ selected.className || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="学院">{{ selected.college || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="专业">{{ selected.major || '—' }}</NDescriptionsItem>
            </NDescriptions>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:account-group-outline" aria-hidden="true" />
                  <span>家长与辅导员信息</span>
                </div>
              </template>
              <NDescriptionsItem label="家长姓名">{{ selected.parentName || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="家长电话">
                <strong class="detail-value">{{ selected.parentPhone || '—' }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="辅导员">{{ selected.teacherName || '—' }}</NDescriptionsItem>
            </NDescriptions>

            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:map-marker-path" aria-hidden="true" />
                  <span>行程信息</span>
                </div>
              </template>
              <NDescriptionsItem label="是否离校">
                {{ selected.isLeaveSchool ? '离校' : '不离校' }}
              </NDescriptionsItem>
              <NDescriptionsItem label="交通方式">
                {{ selected.isLeaveSchool ? selected.travelWay || '—' : '不涉及' }}
              </NDescriptionsItem>
              <NDescriptionsItem label="目的地">
                <span class="detail-text">{{ selected.isLeaveSchool ? selected.destination || '—' : '不涉及' }}</span>
              </NDescriptionsItem>
            </NDescriptions>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:clock-outline" aria-hidden="true" />
                  <span>申请与审批时间</span>
                </div>
              </template>
              <NDescriptionsItem label="申请时间">{{ formatTime(selected.appliedAt) }}</NDescriptionsItem>
              <NDescriptionsItem label="审批时间">{{ formatTime(selected.approvedAt) }}</NDescriptionsItem>
            </NDescriptions>
          </template>
        </template>
        <template #footer>
          <div v-if="selected" class="drawer-footer">
            <span class="drawer-footer__hint">{{ editing ? '保存后更新当前记录' : '请假记录详情' }}</span>
            <div class="drawer-actions">
              <NButton :disabled="saving" @click="editing ? (editing = false) : (drawerVisible = false)">
                {{ editing ? '取消编辑' : '关闭' }}
              </NButton>
              <NButton v-if="editing" type="primary" :loading="saving" @click="submit">保存修改</NButton>
              <NButton v-else type="primary" :disabled="opening || saving" @click="openDrawer(selected, true)">
                编辑请假记录
              </NButton>
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
