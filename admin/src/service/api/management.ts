import { request } from '../request';

export interface LeaveTypeRecord {
  id: number;
  name: string;
  sortOrder: number;
  isEnabled: boolean;
  createdAt: string;
  updatedAt: string;
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
type LeaveTypeWire = Omit<LeaveTypeRecord, 'sortOrder' | 'isEnabled' | 'createdAt' | 'updatedAt'> & {
  sort_order: number;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
};
type ClassWire = Omit<ClassRecord, 'className' | 'isEnabled' | 'createdAt' | 'updatedAt'> & {
  class_name: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
};
interface ListWire<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
export interface ManagementQuery {
  page: number;
  pageSize: number;
  isEnabled: boolean | null;
}
const leaveTypeRecord = (row: LeaveTypeWire): LeaveTypeRecord => ({
  id: row.id,
  name: row.name,
  sortOrder: row.sort_order,
  isEnabled: row.is_enabled,
  createdAt: row.created_at,
  updatedAt: row.updated_at
});
const classRecord = (row: ClassWire): ClassRecord => ({
  id: row.id,
  college: row.college,
  major: row.major,
  className: row.class_name,
  isEnabled: row.is_enabled,
  createdAt: row.created_at,
  updatedAt: row.updated_at
});
export async function queryLeaveTypes(query: ManagementQuery & { name: string }) {
  const result = await request<ListWire<LeaveTypeWire>>({
    url: '/admin/leave-types',
    params: {
      page: query.page,
      page_size: query.pageSize,
      name: query.name.trim(),
      is_enabled: query.isEnabled ?? undefined
    }
  });
  if (result.error) return result;
  return { ...result, data: { items: result.data.items.map(leaveTypeRecord), total: result.data.total } };
}
export async function queryClasses(query: ManagementQuery & { college: string; major: string; className: string }) {
  const result = await request<ListWire<ClassWire>>({
    url: '/admin/classes',
    params: {
      page: query.page,
      page_size: query.pageSize,
      college: query.college.trim(),
      major: query.major.trim(),
      class_name: query.className.trim(),
      is_enabled: query.isEnabled ?? undefined
    }
  });
  if (result.error) return result;
  return { ...result, data: { items: result.data.items.map(classRecord), total: result.data.total } };
}
export async function saveLeaveType(input: Pick<LeaveTypeRecord, 'name' | 'sortOrder' | 'isEnabled'>, id?: number) {
  const result = await request<LeaveTypeWire>({
    url: id === undefined ? '/admin/leave-types' : `/admin/leave-types/${id}`,
    method: id === undefined ? 'post' : 'put',
    data: { name: input.name.trim(), sort_order: input.sortOrder, is_enabled: input.isEnabled }
  });
  if (result.error) return result;
  return { ...result, data: leaveTypeRecord(result.data) };
}
export async function saveClass(
  input: Pick<ClassRecord, 'college' | 'major' | 'className' | 'isEnabled'>,
  id?: number
) {
  const result = await request<ClassWire>({
    url: id === undefined ? '/admin/classes' : `/admin/classes/${id}`,
    method: id === undefined ? 'post' : 'put',
    data: {
      college: input.college.trim(),
      major: input.major.trim(),
      class_name: input.className.trim(),
      is_enabled: input.isEnabled
    }
  });
  if (result.error) return result;
  return { ...result, data: classRecord(result.data) };
}
export function deleteLeaveType(id: number) {
  return request<null>({ url: `/admin/leave-types/${id}`, method: 'delete' });
}
export function deleteClass(id: number) {
  return request<null>({ url: `/admin/classes/${id}`, method: 'delete' });
}
