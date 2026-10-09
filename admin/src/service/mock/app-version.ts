import type { PageQuery } from './management';

export type AppVersionStatus = 'published' | 'archived';

export interface AppVersionRecord {
  id: number;
  platform: string;
  versionCode: number;
  versionName: string;
  packageFile: string;
  releaseNotes: string[] | null;
  status: AppVersionStatus;
  publishedAt: number;
  createdAt: number;
  updatedAt: number;
}

export interface AppVersionQuery {
  keyword: string;
  platform: string | null;
  status: AppVersionStatus | null;
}

const versions: AppVersionRecord[] = Array.from({ length: 16 }, (_, index) => {
  const versionCode = 16 - index;
  const versionName = `0.0.${versionCode}`;
  const publishedAt = new Date(2026, 9, 8 - index, 10, 30).getTime();
  return {
    id: versionCode,
    platform: 'android',
    versionCode,
    versionName,
    packageFile: `leave-android-${versionCode}-${versionName}.apk`,
    releaseNotes:
      index % 5 === 4 ? null : ['优化请假申请页面交互', '修复个人资料显示问题', '提升 Android 客户端运行稳定性'],
    status: index === 0 ? 'published' : 'archived',
    publishedAt,
    createdAt: publishedAt,
    updatedAt: publishedAt
  };
});

const delay = () => new Promise<void>(resolve => window.setTimeout(resolve, 180));
const copy = (row: AppVersionRecord): AppVersionRecord => ({
  ...row,
  releaseNotes: row.releaseNotes === null ? null : [...row.releaseNotes]
});

export async function getAppVersionOverview() {
  await delay();
  const published = versions
    .filter(row => row.platform === 'android' && row.status === 'published')
    .sort((a, b) => b.versionCode - a.versionCode)[0];
  return {
    current: published ? copy(published) : null,
    total: versions.length
  };
}

export async function queryAppVersions(query: AppVersionQuery & PageQuery) {
  await delay();
  const keyword = query.keyword.trim().toLowerCase();
  const items = versions
    .filter(
      row =>
        (!keyword || [row.versionName, String(row.versionCode)].some(value => value.toLowerCase().includes(keyword))) &&
        (query.platform === null || row.platform === query.platform) &&
        (query.status === null || row.status === query.status)
    )
    .sort((a, b) => b.versionCode - a.versionCode || b.id - a.id);
  const offset = (query.page - 1) * query.pageSize;
  return { items: items.slice(offset, offset + query.pageSize).map(copy), total: items.length };
}

export async function getAppVersion(id: number) {
  await delay();
  const row = versions.find(item => item.id === id);
  if (!row) throw new Error('版本记录不存在');
  return copy(row);
}
