import { request } from '../request';

export type Gender = 'male' | 'female';
export interface StudentProfile {
  name: string;
  gender: Gender | null;
  phone: string;
  classId: number | null;
  parentName: string;
  parentPhone: string;
  teacherName: string;
  apartmentId: number | null;
  dormitoryNumber: string;
}
export interface Student extends StudentProfile {
  id: number;
  studentId: string;
  avatarUrl: string;
  status: number;
  createdAt: number;
  updatedAt: number;
  lastSeenAt: number | null;
  lastSeenDevice: string;
  appVersion: string;
  appVersionStatus: 'latest' | 'previous' | 'outdated' | 'unknown';
}
export type StudentListItem = Pick<
  Student,
  | 'id'
  | 'studentId'
  | 'name'
  | 'classId'
  | 'gender'
  | 'phone'
  | 'status'
  | 'appVersion'
  | 'appVersionStatus'
  | 'createdAt'
  | 'lastSeenAt'
>;
export interface StudentQuery {
  studentId: string;
  name: string;
  classId: number | null;
  status: number | null;
}
interface StudentListWire {
  id: number;
  student_id: string;
  name: string;
  gender: Gender | '';
  phone: string;
  class_id: number;
  status: number;
  app_version: string;
  app_version_status: Student['appVersionStatus'];
  created_at: string;
  last_seen_at: string | null;
}
interface StudentWire extends StudentListWire {
  parent_name: string;
  parent_phone: string;
  teacher_name: string;
  apartment_id: number;
  dormitory_number: string;
  avatar_url: string;
  updated_at: string;
  last_seen_device: string;
}
function studentListRecord(row: StudentListWire): StudentListItem {
  return {
    id: row.id,
    studentId: row.student_id,
    name: row.name,
    gender: row.gender || null,
    phone: row.phone,
    classId: row.class_id || null,
    status: row.status,
    appVersion: row.app_version,
    appVersionStatus: row.app_version_status,
    createdAt: Date.parse(row.created_at),
    lastSeenAt: row.last_seen_at === null ? null : Date.parse(row.last_seen_at)
  };
}
function studentRecord(row: StudentWire): Student {
  return {
    ...studentListRecord(row),
    parentName: row.parent_name,
    parentPhone: row.parent_phone,
    teacherName: row.teacher_name,
    apartmentId: row.apartment_id || null,
    dormitoryNumber: row.dormitory_number,
    avatarUrl: row.avatar_url,
    updatedAt: Date.parse(row.updated_at),
    lastSeenDevice: row.last_seen_device
  };
}
export async function queryStudents(query: StudentQuery & { page: number; pageSize: number }) {
  const result = await request<{ items: StudentListWire[]; total: number; page: number; page_size: number }>({
    url: '/admin/students',
    params: {
      page: query.page,
      page_size: query.pageSize,
      student_id: query.studentId.trim(),
      name: query.name.trim(),
      class_id: query.classId ?? undefined,
      status: query.status ?? undefined
    }
  });
  if (result.error) return result;
  return { ...result, data: { items: result.data.items.map(studentListRecord), total: result.data.total } };
}
export async function getStudent(id: number) {
  const result = await request<StudentWire>({ url: `/admin/students/${id}` });
  if (result.error) return result;
  return { ...result, data: studentRecord(result.data) };
}
export async function saveStudent(id: number, input: StudentProfile) {
  const result = await request<StudentWire>({
    url: `/admin/students/${id}`,
    method: 'put',
    data: {
      name: input.name,
      gender: input.gender,
      phone: input.phone,
      class_id: input.classId,
      parent_name: input.parentName,
      parent_phone: input.parentPhone,
      teacher_name: input.teacherName,
      apartment_id: input.apartmentId ?? 0,
      dormitory_number: input.dormitoryNumber
    }
  });
  if (result.error) return result;
  return { ...result, data: studentRecord(result.data) };
}
export async function setStudentStatus(id: number, status: number) {
  const result = await request<StudentWire>({ url: `/admin/students/${id}/status`, method: 'put', data: { status } });
  if (result.error) return result;
  return { ...result, data: studentRecord(result.data) };
}
