import type { ClassOption } from './options';
import { request } from '../request';

export interface LeaveContent {
  leaveTypeId: number | null;
  leaveReason: string;
  startTime: number | null;
  endTime: number | null;
  affectedCourse: string;
  isLeaveSchool: boolean;
  travelWay: string;
  destination: string;
  appliedAt: number | null;
  approvedAt: number | null;
}
export interface LeaveRecordListItem {
  id: number;
  studentId: string;
  name: string;
  className: string;
  leaveTypeId: number;
  leaveTypeName: string;
  startTime: number;
  endTime: number;
  isLeaveSchool: boolean;
  appliedAt: number;
}
export interface LeaveRecord extends LeaveContent {
  id: number;
  studentId: string;
  name: string;
  className: string;
  leaveTypeName: string;
  gender: 'male' | 'female' | null;
  college: string;
  major: string;
  parentName: string;
  parentPhone: string;
  teacherName: string;
  readonly createdAt: number;
}
export interface LeaveQuery {
  studentId: string;
  name: string;
  classSnapshot: string | null;
  leaveTypeId: number | null;
  isLeaveSchool: boolean | null;
  dateRange: [number, number] | null;
}
interface LeaveRecordListWire {
  id: number;
  student_id: string;
  name: string;
  class_name: string;
  leave_type_id: number;
  leave_type_name: string;
  start_time: string;
  end_time: string;
  is_leave_school: boolean;
  applied_at: string;
}
interface LeaveRecordWire extends LeaveRecordListWire {
  gender: 'male' | 'female' | '';
  college: string;
  major: string;
  parent_name: string;
  parent_phone: string;
  teacher_name: string;
  affected_course: string;
  leave_reason: string;
  travel_way: string;
  destination: string;
  approved_at: string | null;
  created_at: string;
}
function listRecord(row: LeaveRecordListWire): LeaveRecordListItem {
  return {
    id: row.id,
    studentId: row.student_id,
    name: row.name,
    className: row.class_name,
    leaveTypeId: row.leave_type_id,
    leaveTypeName: row.leave_type_name,
    startTime: Date.parse(row.start_time),
    endTime: Date.parse(row.end_time),
    isLeaveSchool: row.is_leave_school,
    appliedAt: Date.parse(row.applied_at)
  };
}
function detailRecord(row: LeaveRecordWire): LeaveRecord {
  return {
    ...listRecord(row),
    gender: row.gender || null,
    college: row.college,
    major: row.major,
    parentName: row.parent_name,
    parentPhone: row.parent_phone,
    teacherName: row.teacher_name,
    affectedCourse: row.affected_course,
    leaveReason: row.leave_reason,
    travelWay: row.travel_way,
    destination: row.destination,
    approvedAt: row.approved_at === null ? null : Date.parse(row.approved_at),
    createdAt: Date.parse(row.created_at)
  };
}
export async function queryLeaveRecords(query: LeaveQuery & { page: number; pageSize: number }) {
  const classFilter: Pick<ClassOption, 'college' | 'major' | 'className'> | null = query.classSnapshot
    ? JSON.parse(query.classSnapshot)
    : null;
  const start = query.dateRange ? new Date(query.dateRange[0]) : null;
  const end = query.dateRange ? new Date(query.dateRange[1]) : null;
  start?.setHours(0, 0, 0, 0);
  if (end) {
    end.setHours(0, 0, 0, 0);
    end.setDate(end.getDate() + 1);
  }
  const result = await request<{ items: LeaveRecordListWire[]; total: number; page: number; page_size: number }>({
    url: '/admin/records',
    params: {
      page: query.page,
      page_size: query.pageSize,
      student_id: query.studentId.trim(),
      name: query.name.trim(),
      college: classFilter?.college,
      major: classFilter?.major,
      class_name: classFilter?.className,
      leave_type_id: query.leaveTypeId ?? undefined,
      is_leave_school: query.isLeaveSchool ?? undefined,
      start_time_from: start?.toISOString(),
      start_time_to: end?.toISOString()
    }
  });
  if (result.error) return result;
  return { ...result, data: { items: result.data.items.map(listRecord), total: result.data.total } };
}
export async function getLeaveRecord(id: number) {
  const result = await request<LeaveRecordWire>({ url: `/admin/records/${id}` });
  if (result.error) return result;
  return { ...result, data: detailRecord(result.data) };
}
export async function saveLeaveRecord(id: number, input: LeaveContent) {
  const timeValue = (value: number | null) => (value === null ? null : new Date(value).toISOString());
  const result = await request<LeaveRecordWire>({
    url: `/admin/records/${id}`,
    method: 'put',
    data: {
      leave_type_id: input.leaveTypeId,
      leave_reason: input.leaveReason,
      start_time: timeValue(input.startTime),
      end_time: timeValue(input.endTime),
      affected_course: input.affectedCourse,
      is_leave_school: input.isLeaveSchool,
      travel_way: input.travelWay,
      destination: input.destination,
      applied_at: timeValue(input.appliedAt),
      approved_at: timeValue(input.approvedAt)
    }
  });
  if (result.error) return result;
  return { ...result, data: detailRecord(result.data) };
}
