import { request } from '../request';

export interface FeedbackRecord {
  id: number;
  studentId: string | null;
  isAnonymous: boolean;
  content: string;
  createdAt: string;
  updatedAt: string;
}

export interface FeedbackQuery {
  keyword: string;
  studentId: string;
  isAnonymous: boolean | null;
  page: number;
  pageSize: number;
}

interface FeedbackWire {
  id: number;
  student_id: string | null;
  is_anonymous: boolean;
  content: string;
  created_at: string;
  updated_at: string;
}
const record = (row: FeedbackWire): FeedbackRecord => ({
  id: row.id,
  studentId: row.is_anonymous ? null : row.student_id,
  isAnonymous: row.is_anonymous,
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
      is_anonymous: query.isAnonymous ?? undefined
    }
  });
  if (result.error) return result;
  return {
    ...result,
    data: { items: result.data.items.map(record), total: result.data.total }
  };
}
export async function getFeedback(id: number) {
  const result = await request<FeedbackWire>({ url: `/admin/feedbacks/${id}` });
  if (result.error) return result;
  return { ...result, data: record(result.data) };
}
export function deleteFeedback(id: number) {
  return request<null>({ url: `/admin/feedbacks/${id}`, method: 'delete' });
}
