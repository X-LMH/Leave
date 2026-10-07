import { request } from '../request';

export interface ClassOption {
  id: number;
  college: string;
  major: string;
  className: string;
  isEnabled: boolean;
}

export interface ApartmentOption {
  id: number;
  name: string;
  gender: 'male' | 'female';
  isEnabled: boolean;
}

export interface LeaveTypeOption {
  id: number;
  name: string;
  isEnabled: boolean;
}

export async function getClassOptions() {
  const result = await request<{
    id: number;
    college: string;
    major: string;
    class_name: string;
    is_enabled: boolean;
  }[]>({ url: '/admin/class-options' });
  if (result.error) return result;
  return {
    ...result,
    data: result.data.map((row): ClassOption => ({
      id: row.id,
      college: row.college,
      major: row.major,
      className: row.class_name,
      isEnabled: row.is_enabled
    }))
  };
}

export async function getApartmentOptions(gender?: ApartmentOption['gender']) {
  const result = await request<{ id: number; name: string; gender: ApartmentOption['gender']; is_enabled: boolean }[]>({
    url: '/admin/apartment-options',
    params: { gender }
  });
  if (result.error) return result;
  return {
    ...result,
    data: result.data.map((row): ApartmentOption => ({
      id: row.id,
      name: row.name,
      gender: row.gender,
      isEnabled: row.is_enabled
    }))
  };
}

export async function getLeaveTypeOptions() {
  const result = await request<{ id: number; name: string; is_enabled: boolean }[]>({ url: '/admin/leave-type-options' });
  if (result.error) return result;
  return {
    ...result,
    data: result.data.map((row): LeaveTypeOption => ({
      id: row.id,
      name: row.name,
      isEnabled: row.is_enabled
    }))
  };
}
