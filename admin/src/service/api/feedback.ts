import { request } from '../request';

export interface FeedbackRecord {
  id: number;
  studentId: string | null;
  isAnonymous: boolean;
  isHandled: boolean;
  content: string;
  createdAt: string;
  updatedAt: string;
}

export interface FeedbackQuery {
  keyword: string;
  studentId: string;
  isAnonymous: boolean | null;
  isHandled: boolean | null;
  page: number;
  pageSize: number;
}

export interface FeedbackDetail extends FeedbackRecord {
  attachments: { index: number; url: string }[];
}
interface FeedbackDetailWire extends FeedbackWire {
  attachments: { index: number; url: string }[];
}
interface FeedbackWire {
  id: number;
  student_id: string | null;
  is_anonymous: boolean;
  is_handled: boolean;
  content: string;
  created_at: string;
  updated_at: string;
}
const record = (row: FeedbackWire): FeedbackRecord => ({
  id: row.id,
  studentId: row.is_anonymous ? null : row.student_id,
  isAnonymous: row.is_anonymous,
  isHandled: row.is_handled,
  content: row.content,
  createdAt: row.created_at,
  updatedAt: row.updated_at
});
export async function queryFeedbacks(query: FeedbackQuery) {
  const result = await request<{
    items: FeedbackWire[];
    total: number;
    page: number;
    page_size: number;
  }>({
    url: '/admin/feedbacks',
    params: {
      page: query.page,
      page_size: query.pageSize,
      keyword: query.keyword.trim(),
      student_id: query.studentId.trim(),
      is_anonymous: query.isAnonymous ?? undefined,
      is_handled: query.isHandled ?? undefined
    }
  });
  if (result.error) return result;
  return {
    ...result,
    data: { items: result.data.items.map(record), total: result.data.total }
  };
}
export async function getFeedback(id: number) {
  const result = await request<FeedbackDetailWire>({ url: `/admin/feedbacks/${id}` });
  if (result.error) return result;
  return { ...result, data: { ...record(result.data), attachments: result.data.attachments } };
}
export function deleteFeedback(id: number) {
  return request<null>({ url: `/admin/feedbacks/${id}`, method: 'delete' });
}

export async function updateFeedbackHandling(id: number, isHandled: boolean) {
  const result = await request<FeedbackDetailWire>({
    url: `/admin/feedbacks/${id}/handling`,
    method: 'patch',
    data: { is_handled: isHandled }
  });
  if (result.error) return result;
  return { ...result, data: { ...record(result.data), attachments: result.data.attachments } };
}

export async function getFeedbackAttachment(id: number, index: number): Promise<Blob> {
  const result = await request<Blob, 'blob'>({
    url: `/admin/feedbacks/${id}/attachments/${index}`,
    responseType: 'blob'
  });
  if (result.error) throw new Error('附件加载失败，请重试');
  if (!(result.data instanceof Blob) || !['image/jpeg', 'image/png'].includes(result.data.type)) {
    throw new Error('附件加载失败，请重试');
  }
  return result.data;
}
