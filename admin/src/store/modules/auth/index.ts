import { computed, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { defineStore } from 'pinia';
import { useLoading } from '@sa/hooks';
import { fetchLogin } from '@/service/api';
import { useRouterPush } from '@/hooks/common/router';
import { localStg } from '@/utils/storage';
import { SetupStoreId } from '@/enum';
import { $t } from '@/locales';
import { useRouteStore } from '../route';
import { useTabStore } from '../tab';
import { clearAuthStorage, getToken } from './shared';

function getUserInfoFromToken(token: string): Api.Auth.UserInfo | null {
  const payload = token.split('.')[1];
  if (!payload) {
    return null;
  }

  try {
    const encodedPayload = payload.replace(/-/g, '+').replace(/_/g, '/');
    const claims = JSON.parse(atob(encodedPayload.padEnd(Math.ceil(encodedPayload.length / 4) * 4, '='))) as {
      student_id?: string;
      role?: string;
      exp?: number;
    };

    if (!claims.student_id || claims.role !== 'admin' || !claims.exp || claims.exp <= Date.now() / 1000) {
      return null;
    }

    return {
      userId: claims.student_id,
      userName: claims.student_id,
      roles: ['R_SUPER'],
      buttons: []
    };
  } catch {
    return null;
  }
}

export const useAuthStore = defineStore(SetupStoreId.Auth, () => {
  const route = useRoute();
  const authStore = useAuthStore();
  const routeStore = useRouteStore();
  const tabStore = useTabStore();
  const { toLogin, redirectFromLogin } = useRouterPush(false);
  const { loading: loginLoading, startLoading, endLoading } = useLoading();

  const token = ref('');

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
    clearAuthStorage();

    authStore.$reset();

    if (!route.meta.constant) {
      await toLogin();
    }

    clearSessionTabs();
    await routeStore.resetStore();
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

    const { data: loginToken, error } = await fetchLogin(userName, password);

    if (!error) {
      const pass = await loginByToken(loginToken);

      if (pass) {
        clearSessionTabs();
        await redirectFromLogin(false);

        window.$notification?.success({
          title: $t('page.login.common.loginSuccess'),
          content: $t('page.login.common.welcomeBack', { userName: userInfo.userName }),
          duration: 4500
        });
      }
    } else {
      resetStore();
    }

    endLoading();
  }

  async function loginByToken(loginToken: Api.Auth.LoginToken) {
    // 1. stored in the localStorage, the later requests need it in headers
    localStg.set('token', loginToken.token);
    if (loginToken.refreshToken) {
      localStg.set('refreshToken', loginToken.refreshToken);
    } else {
      localStg.remove('refreshToken');
    }

    const info = getUserInfoFromToken(loginToken.token);
    if (!info) {
      clearAuthStorage();
      return false;
    }

    Object.assign(userInfo, info);
    token.value = loginToken.token;
    return true;
  }

  async function initUserInfo() {
    const maybeToken = getToken();

    if (maybeToken) {
      token.value = maybeToken;
      const info = getUserInfoFromToken(maybeToken);
      if (info) {
        Object.assign(userInfo, info);
      } else {
        await resetStore();
      }
    }
  }

  return {
    token,
    userInfo,
    isStaticSuper,
    isLogin,
    loginLoading,
    resetStore,
    login,
    initUserInfo
  };
});
