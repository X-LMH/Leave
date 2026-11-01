import { request } from '../request';

export function changeAdminPassword(currentPassword: string, newPassword: string, confirmPassword: string) {
  return request<null>({
    url: '/admin/auth/password',
    method: 'post',
    data: {
      password: currentPassword,
      new_password: newPassword,
      re_password: confirmPassword
    }
  });
}

/**
 * Login
 *
 * @param userName User name
 * @param password Password
 */
export function fetchLogin(userName: string, password: string) {
  return request<Api.Auth.LoginToken>({
    url: '/admin/auth/login',
    method: 'post',
    data: {
      user_name: userName,
      password
    }
  });
}

export function fetchAuthIdentity(token: string) {
  return request<Api.Auth.Identity>({
    url: '/admin/auth/me',
    headers: { Authorization: `Bearer ${token}` }
  });
}

export function fetchLogout(token: string) {
  return request<null>({
    url: '/admin/auth/logout',
    method: 'post',
    headers: { Authorization: `Bearer ${token}` }
  });
}

/**
 * return custom backend error
 *
 * @param code error code
 * @param msg error message
 */
export function fetchCustomBackendError(code: string, msg: string) {
  return request({ url: '/auth/error', params: { code, msg } });
}
