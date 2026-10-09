import { request } from '../request';

export type AppVersionStatus = 'published' | 'archived';
export type AppVersionPlatform = 'android' | 'ios' | 'web';

export const appVersionPlatformLabel: Record<AppVersionPlatform, string> = {
  android: 'Android',
  ios: 'iOS',
  web: 'Web'
};

export interface AppVersionRecord {
  id: number;
  platform: AppVersionPlatform;
  versionCode: number;
  versionName: string;
  packageFile: string;
  releaseNotes: string[];
  status: AppVersionStatus;
  publishedAt: number;
  createdAt: number;
  updatedAt: number;
}

export interface AppVersionQuery {
  keyword: string;
  platform: AppVersionPlatform | null;
  status: AppVersionStatus | null;
  page: number;
  pageSize: number;
}

interface AppVersionWire {
  id: number;
  platform: AppVersionPlatform;
  version_code: number;
  version_name: string;
  package_file: string;
  release_notes: string[];
  status: AppVersionStatus;
  published_at: string;
  created_at: string;
  updated_at: string;
}

function versionRecord(row: AppVersionWire): AppVersionRecord {
  return {
    id: row.id,
    platform: row.platform,
    versionCode: row.version_code,
    versionName: row.version_name,
    packageFile: row.package_file,
    releaseNotes: row.release_notes,
    status: row.status,
    publishedAt: new Date(row.published_at).getTime(),
    createdAt: new Date(row.created_at).getTime(),
    updatedAt: new Date(row.updated_at).getTime()
  };
}

export async function queryAppVersions(query: AppVersionQuery) {
  const result = await request<{
    items: AppVersionWire[];
    total: number;
    page: number;
    page_size: number;
  }>({
    url: '/admin/app-versions',
    params: {
      keyword: query.keyword.trim(),
      platform: query.platform ?? undefined,
      status: query.status ?? undefined,
      page: query.page,
      page_size: query.pageSize
    }
  });
  if (result.error) return result;
  return { ...result, data: { items: result.data.items.map(versionRecord), total: result.data.total } };
}

export async function getCurrentAppVersion() {
  const result = await request<AppVersionWire | null>({ url: '/admin/app-versions/current' });
  if (result.error) return result;
  return { ...result, data: result.data ? versionRecord(result.data) : null };
}

export async function getAppVersion(id: number) {
  const result = await request<AppVersionWire>({ url: `/admin/app-versions/${id}` });
  if (result.error) return result;
  return { ...result, data: versionRecord(result.data) };
}

export async function downloadCurrentAppPackage() {
  const result = await request<Blob, 'blob'>({
    url: '/app-download',
    params: { platform: 'android' },
    responseType: 'blob',
    timeout: 0
  });
  if (result.error) return;
  // Never save a JSON error response as an APK.
  if (!(result.data instanceof Blob) || result.data.type.includes('application/json')) {
    window.$message?.error('安装包下载失败，请重试');
    return;
  }
  const disposition = String(result.response.headers['content-disposition'] ?? '');
  const encodedName = /filename\*=UTF-8''([^;]+)/i.exec(disposition)?.[1];
  const plainName = /filename="([^"]+)"|filename=([^;]+)/i.exec(disposition);
  const filename = encodedName ? decodeURIComponent(encodedName) : (plainName?.[1] ?? plainName?.[2]?.trim());
  if (!filename) {
    window.$message?.error('安装包文件名缺失，请重试');
    return;
  }
  const url = URL.createObjectURL(result.data);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.append(link);
  link.click();
  link.remove();
  // Allow the browser to consume the URL before releasing the downloaded data.
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}
