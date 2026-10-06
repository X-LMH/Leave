import { localStg } from '@/utils/storage';

export interface LeaveTypeRecord {
  id: number;
  name: string;
  sortOrder: number;
  isEnabled: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
}

export interface ClassRecord {
  id: number;
  college: string;
  major: string;
  className: string;
  isEnabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface MockStore<T> {
  nextId: number;
  records: T[];
}

export type LeaveTypeInput = Pick<LeaveTypeRecord, 'name' | 'sortOrder' | 'isEnabled'>;
export type ClassInput = Pick<ClassRecord, 'college' | 'major' | 'className' | 'isEnabled'>;

function seedLeaveTypes(): LeaveTypeRecord[] {
  const names = [
    '事假-本科生',
    '病假-本科生',
    '事假-研究生',
    '病假-研究生',
    '家庭事务',
    '就医检查',
    '实习面试',
    '考试',
    '校外活动',
    '交通延误',
    '其他事由',
    '临时事务'
  ];
  const now = new Date().toISOString();
  return names.map((name, index) => ({
    id: index + 1,
    name,
    sortOrder: index + 1,
    isEnabled: index < 10,
    createdAt: now,
    updatedAt: now,
    deletedAt: null
  }));
}

function seedClasses(): ClassRecord[] {
  const now = new Date().toISOString();
  return Array.from({ length: 12 }, (_, index) => ({
    id: index + 1,
    college: index < 6 ? '智能科学与技术学院（网络空间安全学院）' : '经济管理学院',
    major: index < 6 ? '计算机科学与技术' : '工商管理',
    className: index < 6 ? `计24-${index + 1}` : `工商24-${index - 5}`,
    isEnabled: index < 10,
    createdAt: now,
    updatedAt: now
  }));
}

function validRecord(value: unknown, kind: 'mockLeaveTypes' | 'mockClasses'): boolean {
  if (!value || typeof value !== 'object') return false;
  const row = value as Record<string, unknown>;
  const text = (key: string, max: number) =>
    typeof row[key] === 'string' && String(row[key]).trim().length > 0 && String(row[key]).length <= max;
  const timestamps = ['createdAt', 'updatedAt'].every(
    key => typeof row[key] === 'string' && Number.isFinite(Date.parse(String(row[key])))
  );
  if (!Number.isSafeInteger(row.id) || Number(row.id) < 1 || typeof row.isEnabled !== 'boolean' || !timestamps)
    return false;
  if (kind === 'mockLeaveTypes') {
    return (
      text('name', 32) &&
      Number.isSafeInteger(row.sortOrder) &&
      Number(row.sortOrder) >= 0 &&
      (row.deletedAt === null || (typeof row.deletedAt === 'string' && Number.isFinite(Date.parse(row.deletedAt))))
    );
  }
  return text('college', 100) && text('major', 100) && text('className', 64);
}

function load<K extends 'mockLeaveTypes' | 'mockClasses'>(key: K): StorageType.Local[K] {
  // Inspect presence before the storage helper removes malformed JSON.
  const raw = window.localStorage.getItem(`${import.meta.env.VITE_STORAGE_PREFIX || ''}${key}`);
  const stored = localStg.get(key);
  if (
    stored &&
    Number.isSafeInteger(stored.nextId) &&
    stored.nextId > 0 &&
    Array.isArray(stored.records) &&
    stored.records.every(row => validRecord(row, key) && row.id < stored.nextId) &&
    new Set(stored.records.map(row => row.id)).size === stored.records.length
  ) {
    return stored;
  }
  const records = key === 'mockLeaveTypes' ? seedLeaveTypes() : seedClasses();
  const initial = { nextId: records.length + 1, records } as StorageType.Local[K];
  localStg.set(key, initial);
  if (raw !== null) window.$message?.warning('本地模拟数据损坏，已恢复初始数据');
  return initial;
}

function requiredText(value: string, label: string, max: number) {
  const text = value.trim();
  if (!text || text.length > max) throw new Error(`${label}不能为空，且不能超过 ${max} 个字符`);
  return text;
}

export function queryLeaveTypes() {
  return load('mockLeaveTypes')
    .records.filter(row => !row.deletedAt)
    .sort((a, b) => a.sortOrder - b.sortOrder || a.id - b.id);
}

export function saveLeaveType(input: LeaveTypeInput, id?: number) {
  const store = load('mockLeaveTypes');
  const name = requiredText(input.name, '原因名称', 32);
  if (!Number.isSafeInteger(input.sortOrder) || input.sortOrder < 0) throw new Error('排序值必须为非负整数');
  if (store.records.some(row => row.id !== id && row.name.toLocaleLowerCase() === name.toLocaleLowerCase())) {
    throw new Error('原因名称已存在（包括已删除的记录），请使用其他名称');
  }
  const now = new Date().toISOString();
  if (id !== undefined) {
    const row = store.records.find(item => item.id === id && !item.deletedAt);
    if (!row) throw new Error('记录不存在，请刷新列表');
    Object.assign(row, input, { name, updatedAt: now });
  } else {
    store.records.push({ ...input, name, id: store.nextId++, createdAt: now, updatedAt: now, deletedAt: null });
  }
  localStg.set('mockLeaveTypes', store);
}

export function deleteLeaveType(id: number) {
  const store = load('mockLeaveTypes');
  const row = store.records.find(item => item.id === id && !item.deletedAt);
  if (!row) throw new Error('记录不存在，请刷新列表');
  row.deletedAt = new Date().toISOString();
  row.updatedAt = row.deletedAt;
  localStg.set('mockLeaveTypes', store);
}

export function queryClasses() {
  return load('mockClasses').records.sort((a, b) => a.id - b.id);
}

export function saveClass(input: ClassInput, id?: number) {
  const store = load('mockClasses');
  const data = {
    college: requiredText(input.college, '学院', 100),
    major: requiredText(input.major, '专业', 100),
    className: requiredText(input.className, '班级名称', 64),
    isEnabled: input.isEnabled
  };
  if (
    store.records.some(
      row =>
        row.id !== id &&
        ['college', 'major', 'className'].every(
          key =>
            row[key as keyof ClassInput].toString().toLocaleLowerCase() ===
            data[key as keyof ClassInput].toString().toLocaleLowerCase()
        )
    )
  ) {
    throw new Error('该学院、专业下的班级名称已存在');
  }
  const now = new Date().toISOString();
  if (id !== undefined) {
    const row = store.records.find(item => item.id === id);
    if (!row) throw new Error('记录不存在，请刷新列表');
    Object.assign(row, data, { updatedAt: now });
  } else {
    store.records.push({ ...data, id: store.nextId++, createdAt: now, updatedAt: now });
  }
  localStg.set('mockClasses', store);
}

export function deleteClass(id: number) {
  const store = load('mockClasses');
  if (!store.records.some(row => row.id === id)) throw new Error('记录不存在，请刷新列表');
  store.records = store.records.filter(row => row.id !== id);
  localStg.set('mockClasses', store);
}
