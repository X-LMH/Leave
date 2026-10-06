import mockAvatarUrl from '@/assets/svg-icon/avatar.svg?url';

export interface ClassOption {
  id: number;
  college: string;
  major: string;
  className: string;
}
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
}
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
export interface LeaveRecord extends LeaveContent {
  id: number;
  readonly createdAt: number;
  avatarUrl: string;
  studentId: string;
  name: string;
  gender: Gender | null;
  classId: number;
  college: string;
  major: string;
  className: string;
  parentName: string;
  parentPhone: string;
  teacherName: string;
}
export interface StudentQuery {
  studentId: string;
  name: string;
  classId: number | null;
  status: number | null;
}
export interface LeaveQuery {
  studentId: string;
  name: string;
  classId: number | null;
  leaveTypeId: number | null;
  isLeaveSchool: boolean | null;
  dateRange: [number, number] | null;
}
export interface PageQuery {
  page: number;
  pageSize: number;
}
export const classes: ClassOption[] = [
  { id: 1, college: '信息工程学院', major: '软件工程', className: '软件工程2401班' },
  { id: 2, college: '信息工程学院', major: '计算机科学与技术', className: '计算机2402班' },
  { id: 3, college: '经济管理学院', major: '会计学', className: '会计2401班' },
  { id: 4, college: '外国语学院', major: '英语', className: '英语2401班' }
];
export const apartments: { id: number; name: string; gender: Gender }[] = [
  { id: 1, name: '一号宿舍楼', gender: 'male' },
  { id: 2, name: '二号宿舍楼', gender: 'female' },
  { id: 3, name: '三号宿舍楼', gender: 'male' },
  { id: 4, name: '四号宿舍楼', gender: 'female' }
];
export const leaveTypes = [
  { id: 1, name: '病假' },
  { id: 2, name: '事假' },
  { id: 3, name: '其他' }
];
const names = ['张明', '李婷', '王晨', '赵悦', '陈浩', '刘欣', '杨帆', '黄静'];
const baseTime = new Date(2026, 9, 6, 9).getTime();
const students: Student[] = Array.from({ length: 28 }, (_, index) => ({
  id: index + 1,
  avatarUrl: index % 3 === 0 ? '' : mockAvatarUrl,
  studentId: `2024${String(index + 1).padStart(6, '0')}`,
  name: index % 9 === 0 ? '' : names[index % names.length],
  gender: index % 9 === 0 ? null : index % 2 === 0 ? 'male' : 'female',
  phone: index % 9 === 0 ? '' : `138${String(index + 1).padStart(8, '0')}`,
  classId: index % 9 === 0 ? null : (index % 4) + 1,
  parentName: index % 9 === 0 ? '' : `家长${index + 1}`,
  parentPhone: index % 9 === 0 ? '' : `139${String(index + 1).padStart(8, '0')}`,
  teacherName: index % 9 === 0 ? '' : '王老师',
  apartmentId: index % 5 === 0 || index % 9 === 0 ? null : (index % 2) + 1,
  dormitoryNumber: index % 5 === 0 || index % 9 === 0 ? '' : `${301 + index}`,
  status: index % 6 === 0 ? 0 : 1,
  createdAt: baseTime - index * 86400000,
  updatedAt: baseTime - index * 86400000,
  lastSeenAt: index % 7 === 0 ? null : baseTime - index * 3600000,
  lastSeenDevice: index % 7 === 0 ? '' : 'Android 14',
  appVersion: index % 7 === 0 ? '' : '1.0.0'
}));
// Records hold independent snapshots; profile edits must not rewrite historical applications.
const records: LeaveRecord[] = Array.from({ length: 42 }, (_, index) => {
  const student = students.filter(item => item.classId !== null)[index % 24];
  const schoolClass = classes.find(item => item.id === student.classId)!;
  const startTime = baseTime - index * 86400000;
  const isLeaveSchool = index % 2 === 0;
  return {
    id: index + 1,
    avatarUrl: student.avatarUrl,
    studentId: student.studentId,
    name: student.name,
    gender: student.gender,
    classId: schoolClass.id,
    college: schoolClass.college,
    major: schoolClass.major,
    className: schoolClass.className,
    parentName: student.parentName,
    parentPhone: student.parentPhone,
    teacherName: student.teacherName,
    leaveTypeId: (index % 3) + 1,
    leaveReason: index % 3 === 0 ? '身体不适，需要前往医院检查并休息。' : '因个人事务申请请假，返校后及时补上课程。',
    startTime,
    endTime: startTime + (index % 5) * 86400000 + 5400000,
    affectedCourse: index % 4 === 0 ? '' : '大学英语、专业课程',
    isLeaveSchool,
    travelWay: isLeaveSchool ? '公共交通' : '',
    destination: isLeaveSchool ? '市区' : '',
    createdAt: startTime - 7100000,
    appliedAt: startTime - 7200000,
    approvedAt: index % 8 === 0 ? null : startTime - 3600000
  };
});
const delay = () => new Promise<void>(resolve => setTimeout(resolve, 180));
const matches = (value: string, search: string) => value.includes(search.trim());
function paginate<T>(items: T[], query: PageQuery) {
  return { items: items.slice((query.page - 1) * query.pageSize, query.page * query.pageSize), total: items.length };
}
export async function queryStudents(query: StudentQuery & PageQuery) {
  await delay();
  const items = students
    .filter(
      row =>
        matches(row.studentId, query.studentId) &&
        matches(row.name, query.name) &&
        (query.classId === null || row.classId === query.classId) &&
        (query.status === null || row.status === query.status)
    )
    .sort((a, b) => b.createdAt - a.createdAt);
  return paginate(
    items.map(row => ({ ...row })),
    query
  );
}
export async function queryLeaveRecords(query: LeaveQuery & PageQuery) {
  await delay();
  const endDate = query.dateRange ? new Date(query.dateRange[1]) : null;
  endDate?.setHours(23, 59, 59, 999);
  const startDate = query.dateRange ? new Date(query.dateRange[0]) : null;
  startDate?.setHours(0, 0, 0, 0);
  const items = records
    .filter(
      row =>
        matches(row.studentId, query.studentId) &&
        matches(row.name, query.name) &&
        (query.classId === null || row.classId === query.classId) &&
        (query.leaveTypeId === null || row.leaveTypeId === query.leaveTypeId) &&
        (query.isLeaveSchool === null || row.isLeaveSchool === query.isLeaveSchool) &&
        (!startDate ||
          (row.startTime !== null && row.startTime >= startDate.getTime() && row.startTime <= endDate!.getTime()))
    )
    .sort((a, b) => (b.appliedAt ?? 0) - (a.appliedAt ?? 0));
  return paginate(
    items.map(row => ({ ...row })),
    query
  );
}
export async function getStudent(id: number) {
  await delay();
  const row = students.find(item => item.id === id);
  if (!row) throw new Error('学生不存在');
  return { ...row };
}
export async function getLeaveRecord(id: number) {
  await delay();
  const row = records.find(item => item.id === id);
  if (!row) throw new Error('请假记录不存在');
  return { ...row };
}
export async function saveStudent(id: number, profile: StudentProfile) {
  await delay();
  const row = students.find(item => item.id === id);
  if (!row) throw new Error('学生不存在');
  Object.assign(row, profile, { updatedAt: Date.now() });
  return { ...row };
}
export async function setStudentStatus(id: number, status: number) {
  await delay();
  const row = students.find(item => item.id === id);
  if (!row) throw new Error('学生不存在');
  row.status = status;
  row.updatedAt = Date.now();
  return { ...row };
}
export async function saveLeaveRecord(id: number, content: LeaveContent) {
  await delay();
  const row = records.find(item => item.id === id);
  if (!row) throw new Error('请假记录不存在');
  Object.assign(row, content, {
    travelWay: content.isLeaveSchool ? content.travelWay : '',
    destination: content.isLeaveSchool ? content.destination : ''
  });
  return { ...row };
}
export const formatTime = (value: number | null) =>
  value === null ? '—' : new Date(value).toLocaleString('zh-CN', { hour12: false });
export const genderLabel = (value: Gender | null) => (value === null ? '—' : value === 'male' ? '男' : '女');
export function durationLabel(start: number | null, end: number | null) {
  if (start === null || end === null || end <= start) return '—';
  const hours = Math.ceil((end - start) / 3600000);
  return `${Math.floor(hours / 24)}天${hours % 24}小时`;
}
