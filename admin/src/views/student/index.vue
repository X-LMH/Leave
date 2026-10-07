<script setup lang="ts">
import { computed, h, nextTick, reactive, ref, watch } from 'vue';
import { NButton, NTag, useThemeVars } from 'naive-ui';
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui';
import StudentVersion from './components/student-version.vue';
import { useAppStore } from '@/store/modules/app';
import { getStudent, queryStudents, saveStudent, setStudentStatus } from '@/service/api/student';
import type { Student, StudentListItem, StudentProfile, StudentQuery } from '@/service/api/student';
import { getClassOptions, getApartmentOptions } from '@/service/api/options';
import type { ClassOption, ApartmentOption } from '@/service/api/options';

function formatTime(value: number) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}
function genderLabel(value: StudentProfile['gender']) {
  if (value === 'male') return '男';
  if (value === 'female') return '女';
  return '—';
}

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
function renderTime(value: number | null, empty = '—') {
  if (value === null) return h('span', { class: 'table-muted' }, empty);
  const date = new Date(value);
  return h('div', { class: 'table-time' }, [
    h('span', {}, date.toLocaleDateString('zh-CN')),
    h('span', { class: 'table-time__clock' }, date.toLocaleTimeString('zh-CN', { hour12: false }))
  ]);
}
const search = reactive<StudentQuery>({ studentId: '', name: '', classId: null, status: null });
const applied = reactive({ ...search });
const rows = ref<StudentListItem[]>([]);
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
    const { data, error } = await queryStudents({ ...applied, page: page.value, pageSize: pageSize.value });
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
  Object.assign(search, { studentId: '', name: '', classId: null, status: null });
  applySearch();
}
const selected = ref<Student | null>(null);
const classes = ref<ClassOption[]>([]);
const apartments = ref<ApartmentOption[]>([]);
async function loadOptions() {
  const [classResult, apartmentResult] = await Promise.all([getClassOptions(), getApartmentOptions()]);
  if (classResult.error || apartmentResult.error) return false;
  classes.value = classResult.data;
  apartments.value = apartmentResult.data;
  return true;
}
const classOptions = computed(() =>
  classes.value.map(row => ({
    label: `${row.college} / ${row.className}${row.isEnabled ? '' : '（已停用）'}`,
    value: row.id
  }))
);
const editClassOptions = computed(() =>
  classes.value.map(row => ({
    label: `${row.college} / ${row.className}${row.isEnabled ? '' : '（已停用）'}`,
    value: row.id,
    disabled: !row.isEnabled && row.id !== selected.value?.classId
  }))
);
const statusOptions = [
  { label: '启用', value: 1 },
  { label: '停用', value: 0 }
];
const genderOptions = [
  { label: '男', value: 'male' },
  { label: '女', value: 'female' }
];
const drawerVisible = ref(false);
const editing = ref(false);
const saving = ref(false);
const opening = ref(false);
const changingId = ref<number | null>(null);
const formRef = ref<FormInst | null>(null);
const model = reactive<StudentProfile>({
  name: '',
  gender: null,
  phone: '',
  classId: null,
  parentName: '',
  parentPhone: '',
  teacherName: '',
  apartmentId: null,
  dormitoryNumber: ''
});
const schoolClass = computed(() =>
  classes.value.find(row => row.id === (editing.value ? model.classId : selected.value?.classId))
);
const apartmentOptions = computed(() =>
  apartments.value
    .filter(row => row.gender === model.gender)
    .map(row => ({
      label: `${row.name}${row.isEnabled ? '' : '（已停用）'}`,
      value: row.id,
      disabled: !row.isEnabled && row.id !== selected.value?.apartmentId
    }))
);
watch(
  () => model.gender,
  gender => {
    if (
      model.apartmentId !== null &&
      !apartments.value.some(row => row.id === model.apartmentId && row.gender === gender)
    )
      model.apartmentId = null;
  }
);
const nameRule = {
  required: true,
  validator: (_rule: unknown, value: string) => Boolean(value?.trim()) && value.trim().length <= 64,
  message: '请输入64字符以内的姓名',
  trigger: ['blur', 'input']
};
const phoneRule = {
  required: true,
  validator: (_rule: unknown, value: string) => /^1[3-9][0-9]{9}$/.test(value?.trim()),
  message: '请输入有效的11位手机号码',
  trigger: ['blur', 'input']
};
const rules: FormRules = {
  name: nameRule,
  parentName: nameRule,
  teacherName: nameRule,
  phone: phoneRule,
  parentPhone: phoneRule,
  gender: { required: true, message: '请选择性别', trigger: 'change' },
  classId: { required: true, type: 'number', message: '请选择班级', trigger: 'change' },
  dormitoryNumber: { max: 32, message: '宿舍号最多32字符', trigger: 'input' }
};
async function openDrawer(row: Pick<StudentListItem, 'id'>, edit: boolean) {
  if (opening.value || saving.value || changingId.value !== null) return;
  opening.value = true;
  try {
    const { data: student, error } = await getStudent(row.id);
    if (error || !(await loadOptions())) return;
    selected.value = student;
    const data = selected.value;
    Object.assign(model, {
      name: data.name,
      gender: data.gender,
      phone: data.phone,
      classId: data.classId,
      parentName: data.parentName,
      parentPhone: data.parentPhone,
      teacherName: data.teacherName,
      apartmentId: data.apartmentId,
      dormitoryNumber: data.dormitoryNumber
    });
    editing.value = edit;
    drawerVisible.value = true;
    await nextTick();
    formRef.value?.restoreValidation();
  } catch {
    window.$message?.error('加载学生资料失败');
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
    const { data, error } = await saveStudent(selected.value.id, {
      ...model,
      name: model.name.trim(),
      phone: model.phone.trim(),
      parentName: model.parentName.trim(),
      parentPhone: model.parentPhone.trim(),
      teacherName: model.teacherName.trim(),
      dormitoryNumber: model.dormitoryNumber.trim()
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
function confirmStatus(row: Pick<StudentListItem, 'id' | 'studentId' | 'status'>) {
  if (editing.value || opening.value || saving.value || changingId.value !== null) return;
  const status = row.status === 1 ? 0 : 1;
  const label = status === 1 ? '启用' : '停用';
  window.$dialog?.warning({
    title: `${label}确认`,
    content: `确认${label}学生账号“${row.studentId}”吗？`,
    positiveText: label,
    negativeText: '取消',
    onPositiveClick: async () => {
      if (changingId.value !== null) return false;
      changingId.value = row.id;
      try {
        const { data: updated, error } = await setStudentStatus(row.id, status);
        if (error) return false;
        if (selected.value?.id === row.id) selected.value = updated;
        await reload();
        window.$message?.success(`${label}成功`);
        return true;
      } catch {
        window.$message?.error('操作失败，请重试');
        return false;
      } finally {
        changingId.value = null;
      }
    }
  });
}
const columns: DataTableColumns<StudentListItem> = [
  { key: 'studentId', title: '学号', width: 130, className: 'table-id' },
  {
    key: 'name',
    title: '姓名',
    width: 120,
    render: row => h('span', { class: 'table-person__name' }, row.name || '未完善资料')
  },
  {
    key: 'classId',
    title: '班级',
    width: 160,
    ellipsis: { tooltip: true },
    render: row => classes.value.find(item => item.id === row.classId)?.className || '—'
  },
  { key: 'gender', title: '性别', width: 60, render: row => genderLabel(row.gender) },
  { key: 'phone', title: '手机号', width: 130, render: row => row.phone || '—' },
  {
    key: 'status',
    title: '账号状态',
    width: 90,
    render: row =>
      h(NTag, { type: row.status === 1 ? 'success' : 'error', bordered: false }, () =>
        row.status === 1 ? '启用' : '停用'
      )
  },
  {
    key: 'appVersion',
    title: '使用版本',
    width: 190,
    render: row => h(StudentVersion, { student: row })
  },
  { key: 'createdAt', title: '注册时间', width: 130, render: row => renderTime(row.createdAt) },
  {
    key: 'lastSeenAt',
    title: '最近活跃时间',
    width: 140,
    render: row => renderTime(row.lastSeenAt, '暂无')
  },
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
void loadOptions();
void reload();
</script>

<template>
  <div class="management-page management-page--student" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon" aria-hidden="true"><SvgIcon icon="mdi:account-school-outline" /></span>
        <div>
          <p class="management-header__eyebrow">教务管理 / 学生档案</p>
          <h1 class="management-header__title">学生管理</h1>
          <p class="management-header__description">集中查看学生档案，维护个人资料与账号状态。</p>
        </div>
      </div>
      <div class="management-header__count" aria-live="polite">
        <span class="management-header__count-label">当前查询结果</span>
        <span>
          <strong>{{ total }}</strong>
          位学生
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
          <NSelect v-model:value="search.classId" :options="classOptions" filterable clearable placeholder="全部班级" />
        </NFormItem>
        <NFormItem label="账号状态">
          <NSelect v-model:value="search.status" :options="statusOptions" clearable placeholder="全部状态" />
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
          <span>学生列表</span>
          <span class="section-heading__badge">{{ total }}</span>
        </div>
      </template>
      <template #header-extra>
        <span class="record-count">按注册时间正序</span>
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
        :scroll-x="1160"
      >
        <template #empty>
          <div class="table-empty">
            <SvgIcon icon="mdi:account-search-outline" aria-hidden="true" />
            <strong>没有找到匹配的学生</strong>
            <span>试试调整筛选条件，或重置后查看全部数据。</span>
            <NButton size="small" @click="resetSearch">重置筛选</NButton>
          </div>
        </template>
      </NDataTable>
    </NCard>
    <NDrawer
      v-model:show="drawerVisible"
      :width="appStore.isMobile ? '100%' : 720"
      :close-on-esc="!saving"
      :mask-closable="!saving"
    >
      <NDrawerContent
        class="management-drawer"
        :style="themeStyle"
        :title="editing ? '编辑学生资料' : '学生详情'"
        :closable="!saving"
        :native-scrollbar="false"
      >
        <template v-if="selected">
          <div class="drawer-summary">
            <NAvatar
              class="drawer-summary__avatar"
              :size="48"
              :src="selected.avatarUrl || undefined"
              :img-props="{ alt: `${selected.name || '学生'}的头像` }"
            >
              {{ selected.name?.slice(0, 1) || '学' }}
            </NAvatar>
            <div class="drawer-summary__identity">
              <strong>{{ selected.name || '未完善资料' }}</strong>
              <span>{{ selected.studentId }} · {{ schoolClass?.className || '未设置班级' }}</span>
            </div>
            <NTag :type="selected.status === 1 ? 'success' : 'error'" :bordered="false" round>
              {{ selected.status === 1 ? '账号启用' : '账号停用' }}
            </NTag>
          </div>
          <div v-if="!editing" class="detail-highlights">
            <div>
              <span>最近使用时间</span>
              <strong>{{ selected.lastSeenAt === null ? '暂无' : formatTime(selected.lastSeenAt) }}</strong>
            </div>
            <div>
              <span>使用版本</span>
              <StudentVersion :student="selected" />
            </div>
          </div>
          <NDescriptions :column="appStore.isMobile ? 1 : 2" label-placement="top" class="detail-section">
            <template #header>
              <div class="detail-section__heading">
                <SvgIcon icon="mdi:shield-account-outline" aria-hidden="true" />
                <span>账号信息</span>
              </div>
            </template>
            <NDescriptionsItem label="学号">{{ selected.studentId }}</NDescriptionsItem>
            <NDescriptionsItem label="账号状态">
              <NTag :type="selected.status === 1 ? 'success' : 'error'" :bordered="false">
                {{ selected.status === 1 ? '启用' : '停用' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="注册时间">{{ formatTime(selected.createdAt) }}</NDescriptionsItem>
            <NDescriptionsItem v-if="editing" label="最近使用时间">
              {{ selected.lastSeenAt === null ? '暂无' : formatTime(selected.lastSeenAt) }}
            </NDescriptionsItem>
            <NDescriptionsItem label="活跃设备">{{ selected.lastSeenDevice || '暂无' }}</NDescriptionsItem>
            <NDescriptionsItem v-if="editing" label="使用版本">
              <StudentVersion :student="selected" />
            </NDescriptionsItem>
          </NDescriptions>
          <NForm v-if="editing" ref="formRef" :model="model" :rules="rules" label-placement="top" class="edit-form">
            <div class="form-section-heading">
              <span>01</span>
              <h2>个人与班级资料</h2>
            </div>
            <NFormItem label="姓名" path="name"><NInput v-model:value="model.name" :maxlength="64" /></NFormItem>
            <NFormItem label="性别" path="gender">
              <NSelect v-model:value="model.gender" :options="genderOptions" />
            </NFormItem>
            <NFormItem label="手机号" path="phone"><NInput v-model:value="model.phone" :maxlength="20" /></NFormItem>
            <NFormItem label="班级" path="classId">
              <NSelect v-model:value="model.classId" :options="editClassOptions" filterable />
            </NFormItem>
            <NFormItem label="学院"><NInput :value="schoolClass?.college || ''" readonly /></NFormItem>
            <NFormItem label="专业"><NInput :value="schoolClass?.major || ''" readonly /></NFormItem>
            <div class="form-section-heading">
              <span>02</span>
              <h2>联系与住宿信息</h2>
            </div>
            <NFormItem label="家长姓名" path="parentName">
              <NInput v-model:value="model.parentName" :maxlength="64" />
            </NFormItem>
            <NFormItem label="家长电话" path="parentPhone">
              <NInput v-model:value="model.parentPhone" :maxlength="20" />
            </NFormItem>
            <NFormItem label="辅导员" path="teacherName">
              <NInput v-model:value="model.teacherName" :maxlength="64" />
            </NFormItem>
            <NFormItem label="宿舍楼">
              <NSelect
                v-model:value="model.apartmentId"
                :options="apartmentOptions"
                clearable
                :disabled="model.gender === null"
                placeholder="请选择宿舍楼"
              />
            </NFormItem>
            <NFormItem label="宿舍号" path="dormitoryNumber">
              <NInput v-model:value="model.dormitoryNumber" :maxlength="32" />
            </NFormItem>
          </NForm>
          <template v-else>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:account-school-outline" aria-hidden="true" />
                  <span>个人与班级资料</span>
                </div>
              </template>
              <NDescriptionsItem label="姓名">{{ selected.name || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="性别">{{ genderLabel(selected.gender) }}</NDescriptionsItem>
              <NDescriptionsItem label="手机号">
                <strong class="detail-value">{{ selected.phone || '—' }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="班级">{{ schoolClass?.className || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="学院">{{ schoolClass?.college || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="专业">{{ schoolClass?.major || '—' }}</NDescriptionsItem>
            </NDescriptions>
            <NDescriptions :column="appStore.isMobile ? 1 : 2" class="detail-section" label-placement="top">
              <template #header>
                <div class="detail-section__heading">
                  <SvgIcon icon="mdi:home-account" aria-hidden="true" />
                  <span>家长与住宿信息</span>
                </div>
              </template>
              <NDescriptionsItem label="家长姓名">{{ selected.parentName || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="家长电话">
                <strong class="detail-value">{{ selected.parentPhone || '—' }}</strong>
              </NDescriptionsItem>
              <NDescriptionsItem label="辅导员">{{ selected.teacherName || '—' }}</NDescriptionsItem>
              <NDescriptionsItem label="宿舍楼">
                {{ apartments.find(row => row.id === selected?.apartmentId)?.name || '—' }}
              </NDescriptionsItem>
              <NDescriptionsItem label="宿舍号">{{ selected.dormitoryNumber || '—' }}</NDescriptionsItem>
            </NDescriptions>
          </template>
        </template>
        <template #footer>
          <div v-if="selected" class="drawer-footer">
            <span class="drawer-footer__hint">{{ editing ? '保存后更新当前资料' : '学生档案' }}</span>
            <div class="drawer-actions">
              <NButton
                v-if="!editing"
                :type="selected.status === 1 ? 'warning' : 'success'"
                secondary
                :loading="changingId === selected.id"
                :disabled="opening || saving || changingId !== null"
                @click="confirmStatus(selected)"
              >
                {{ selected.status === 1 ? '停用账号' : '启用账号' }}
              </NButton>
              <NButton :disabled="saving" @click="editing ? (editing = false) : (drawerVisible = false)">
                {{ editing ? '取消编辑' : '关闭' }}
              </NButton>
              <NButton v-if="editing" type="primary" :loading="saving" @click="submit">保存修改</NButton>
              <NButton
                v-else
                type="primary"
                :disabled="opening || saving || changingId !== null"
                @click="openDrawer(selected, true)"
              >
                编辑资料
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
