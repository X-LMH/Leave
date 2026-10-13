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

/**
 * Refresh token
 *
 * @param refreshToken Refresh token
 */
export function fetchRefreshToken(refreshToken: string) {
  return request<Api.Auth.LoginToken>({
    url: '/auth/refreshToken',
    method: 'post',
    data: {
      refreshToken
    }
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
