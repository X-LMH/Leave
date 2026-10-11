import { request } from '../request';

export type Gender = 'male' | 'female';

export interface ApartmentRecord {
  id: number;
  name: string;
  gender: Gender;
  sortOrder: number;
  isEnabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ApartmentQuery {
  name: string;
  gender: Gender | null;
  isEnabled: boolean | null;
  page: number;
  pageSize: number;
}

interface ApartmentWire {
  id: number;
  name: string;
  gender: Gender;
  sort_order: number;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}
const record = (row: ApartmentWire): ApartmentRecord => ({
  id: row.id,
  name: row.name,
  gender: row.gender,
  sortOrder: row.sort_order,
  isEnabled: row.is_enabled,
  createdAt: row.created_at,
  updatedAt: row.updated_at
});
export async function queryApartments(query: ApartmentQuery) {
  const result = await request<{
    items: ApartmentWire[];
    total: number;
    page: number;
    page_size: number;
  }>({
    url: '/admin/apartments',
    params: {
      page: query.page,
      page_size: query.pageSize,
      name: query.name.trim(),
      gender: query.gender ?? undefined,
      is_enabled: query.isEnabled ?? undefined
    }
  });
  if (result.error) return result;
  return {
    ...result,
    data: { items: result.data.items.map(record), total: result.data.total }
  };
}
export async function saveApartment(
  input: Pick<ApartmentRecord, 'name' | 'gender' | 'sortOrder' | 'isEnabled'>,
  id?: number
) {
  const result = await request<ApartmentWire>({
    url: id === undefined ? '/admin/apartments' : `/admin/apartments/${id}`,
    method: id === undefined ? 'post' : 'put',
    data: {
      name: input.name.trim(),
      gender: input.gender,
      sort_order: input.sortOrder,
      is_enabled: input.isEnabled
    }
  });
  if (result.error) return result;
  return { ...result, data: record(result.data) };
}
export function deleteApartment(id: number) {
  return request<null>({ url: `/admin/apartments/${id}`, method: 'delete' });
}
