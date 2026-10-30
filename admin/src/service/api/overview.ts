import { request } from '../request';

export interface OverviewRecord {
  id: number;
  name: string;
  studentId: string;
  className: string;
  leaveTypeId: number;
  leaveTypeName: string;
  appliedAt: number;
}

export interface WeeklyApplicationPoint {
  start: number;
  end: number;
  count: number;
  isCurrent: boolean;
}

export interface OverviewVersionUsage {
  total: number;
  latestCount: number;
  groups: {
    key: 'latest' | 'previous' | 'older' | 'unreported' | 'other';
    version: string;
    count: number;
    percentage: number;
  }[];
}

export interface Overview {
  generatedAt: number;
  statistics: {
    studentCount: number;
    classCount: number;
    recordCount: number;
    unhandledFeedbackCount: number;
  };
  weeklyApplications: { weeks: WeeklyApplicationPoint[]; total: number };
  versionUsage: OverviewVersionUsage;
  recentRecords: OverviewRecord[];
  latestVersion: {
    versionName: string;
    versionCode: number;
    publishedAt: number;
    releaseNotes: string[];
  } | null;
}

interface OverviewWire {
  generated_at: string;
  statistics: {
    student_count: number;
    class_count: number;
    record_count: number;
    unhandled_feedback_count: number;
  };
  weekly_applications: {
    weeks: { start: string; end: string; count: number; is_current: boolean }[];
    total: number;
  };
  version_usage: {
    total: number;
    latest_count: number;
    groups: OverviewVersionUsage['groups'];
  };
  recent_records: {
    id: number;
    name: string;
    student_id: string;
    class_name: string;
    leave_type_id: number;
    leave_type_name: string;
    applied_at: string;
  }[];
  latest_android_version: {
    version_name: string;
    version_code: number;
    published_at: string;
    release_notes: string[];
  } | null;
}

export function adaptOverview(row: OverviewWire): Overview {
  return {
    generatedAt: Date.parse(row.generated_at),
    statistics: {
      studentCount: row.statistics.student_count,
      classCount: row.statistics.class_count,
      recordCount: row.statistics.record_count,
      unhandledFeedbackCount: row.statistics.unhandled_feedback_count
    },
    weeklyApplications: {
      weeks: row.weekly_applications.weeks.map(week => ({
        start: Date.parse(week.start),
        end: Date.parse(week.end),
        count: week.count,
        isCurrent: week.is_current
      })),
      total: row.weekly_applications.total
    },
    versionUsage: {
      total: row.version_usage.total,
      latestCount: row.version_usage.latest_count,
      groups: row.version_usage.groups
    },
    recentRecords: row.recent_records.map(record => ({
      id: record.id,
      name: record.name,
      studentId: record.student_id,
      className: record.class_name,
      leaveTypeId: record.leave_type_id,
      leaveTypeName: record.leave_type_name,
      appliedAt: Date.parse(record.applied_at)
    })),
    latestVersion: row.latest_android_version
      ? {
          versionName: row.latest_android_version.version_name,
          versionCode: row.latest_android_version.version_code,
          publishedAt: Date.parse(row.latest_android_version.published_at),
          releaseNotes: row.latest_android_version.release_notes
        }
      : null
  };
}

export async function getOverview() {
  const result = await request<OverviewWire>({ url: '/admin/overview' });
  if (result.error) return result;
  return { ...result, data: adaptOverview(result.data) };
}
