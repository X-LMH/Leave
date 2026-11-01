import { computed, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { defineStore } from 'pinia';
import { useLoading } from '@sa/hooks';
import { fetchLogin, fetchAuthIdentity, fetchLogout } from '@/service/api';
import { useRouterPush } from '@/hooks/common/router';
import { localStg } from '@/utils/storage';
import { SetupStoreId } from '@/enum';
import { $t } from '@/locales';
import { useRouteStore } from '../route';
import { useTabStore } from '../tab';
import { clearAuthStorage, getToken } from './shared';

export const useAuthStore = defineStore(SetupStoreId.Auth, () => {
  const route = useRoute();
  const authStore = useAuthStore();
  const routeStore = useRouteStore();
  const tabStore = useTabStore();
  const { toLogin, redirectFromLogin } = useRouterPush(false);
  const { loading: loginLoading, startLoading, endLoading } = useLoading();

  const token = ref('');
  let loginAttempt = 0;

  const userInfo: Api.Auth.UserInfo = reactive({
    userId: '',
    userName: '',
    roles: [],
    buttons: []
  });

  /** is super role in static route */
  const isStaticSuper = computed(() => {
    const { VITE_AUTH_ROUTE_MODE, VITE_STATIC_SUPER_ROLE } = import.meta.env;

    return VITE_AUTH_ROUTE_MODE === 'static' && userInfo.roles.includes(VITE_STATIC_SUPER_ROLE);
  });

  /** Is login */
  const isLogin = computed(() => Boolean(token.value));

  /** Reset auth store */
  async function resetStore() {
    const attempt = ++loginAttempt;
    clearAuthStorage();

    authStore.$reset();

    clearSessionTabs();
    await routeStore.resetStore();

    // 清理期间已开始新登录时，不再执行旧会话的页面跳转。
    if (attempt === loginAttempt && !getToken() && !route.meta.constant) {
      await toLogin();
    }
  }

  /** 新登录会话不恢复上次会话的标签，包括固定标签。 */
  function clearSessionTabs() {
    localStg.remove('globalTabs');
    localStg.remove('lastLoginUserId');
    tabStore.resetTabs();
  }

  /**
   * Login
   *
   * @param userName User name
   * @param password Password
   */
  async function login(userName: string, password: string) {
    startLoading();

    const attempt = ++loginAttempt;
    try {
      const { data: loginToken, error } = await fetchLogin(userName, password);
      if (error || attempt !== loginAttempt) return;
      const pass = await loginByToken(loginToken);
      if (pass && attempt === loginAttempt) {
        clearSessionTabs();
        await redirectFromLogin(false);
        window.$notification?.success({
          title: $t('page.login.common.loginSuccess'),
          content: $t('page.login.common.welcomeBack', { userName: userInfo.userName }),
          duration: 4500
        });
      }
    } finally {
      if (attempt === loginAttempt) endLoading();
    }
  }

  async function loginByToken(loginToken: Api.Auth.LoginToken) {
    Object.assign(userInfo, { userId: '', userName: '', roles: [], buttons: [] });
    localStg.set('token', loginToken.token);
    localStg.remove('refreshToken');
    token.value = loginToken.token;
    return initUserInfo();
  }

  async function initUserInfo() {
    const expectedToken = getToken();
    if (!expectedToken) return false;
    token.value = expectedToken;
    const { data, error } = await fetchAuthIdentity(expectedToken);
    // Old responses must not overwrite a newer login or restore a logged-out identity.
    if (getToken() !== expectedToken) return false;
    if (error) return false;
    if (data.role !== 'admin') {
      await resetStore();
      return false;
    }
    Object.assign(userInfo, {
      userId: data.student_id,
      userName: data.student_id,
      roles: ['R_SUPER'],
      buttons: []
    });
    return true;
  }

  async function logout() {
    const expectedToken = getToken();
    if (!expectedToken) {
      await resetStore();
      return;
    }
    const { error } = await fetchLogout(expectedToken);
    if (!error && getToken() === expectedToken) await resetStore();
  }

  return {
    token,
    userInfo,
    isStaticSuper,
    isLogin,
    loginLoading,
    resetStore,
    logout,
    login,
    initUserInfo
  };
});
