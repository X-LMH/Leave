import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  const storage = new Map([
    ['globalTabs', ['student', 'fixed-tab']],
    ['lastLoginUserId', 'admin']
  ]);
  const tabs = {
    items: ['student', 'fixed-tab'],
    resetTabs() {
      this.items = [];
    }
  };
  const redirects = [];
  let failed = false;
  const store = { $reset() {} };
  let initialized = false;
  const token = 'opaque-random-session';
  let identity = async () => ({ data: { student_id: 'admin', role: 'admin' }, error: null });
  let logoutError = null;
  let logoutCalls = 0;
  let resetRoutes = async () => {};
  const credentials = [];
  const context = {
    exports: {},
    testEnv: {},
    window: { $notification: { success() {} } },
    require(name) {
      if (name === 'vue')
        return {
          computed: fn => ({
            get value() {
              return fn();
            }
          }),
          reactive: value => value,
          ref: value => ({ value })
        };
      if (name === 'vue-router') return { useRoute: () => ({ meta: {} }) };
      if (name === 'pinia')
        return {
          defineStore: (_id, create) => () => {
            if (!initialized) {
              initialized = true;
              Object.assign(store, create());
            }
            return store;
          }
        };
      if (name === '@sa/hooks') return { useLoading: () => ({ startLoading() {}, endLoading() {} }) };
      if (name === '@/service/api')
        return {
          fetchLogin: async () => ({ data: { token }, error: failed ? {} : null }),
          fetchAuthIdentity: expectedToken => {
            credentials.push(expectedToken);
            return identity();
          },
          fetchLogout: async expectedToken => {
            credentials.push(expectedToken);
            logoutCalls++;
            return { error: logoutError };
          }
        };
      if (name === '@/hooks/common/router')
        return {
          useRouterPush: () => ({
            toLogin: async () => redirects.push('login'),
            redirectFromLogin: async redirect => redirects.push(redirect)
          })
        };
      if (name === '@/utils/storage')
        return {
          localStg: {
            get: key => storage.get(key),
            set: (key, value) => storage.set(key, value),
            remove: key => storage.delete(key)
          }
        };
      if (name === '@/enum') return { SetupStoreId: { Auth: 'auth' } };
      if (name === '@/locales') return { $t: value => value };
      if (name === '../route') return { useRouteStore: () => ({ resetStore: () => resetRoutes() }) };
      if (name === '../tab') return { useTabStore: () => tabs };
      if (name === './shared')
        return {
          clearAuthStorage: () => {
            storage.delete('token');
            storage.delete('refreshToken');
          },
          getToken: () => storage.get('token')
        };
      throw new Error(`Unexpected import: ${name}`);
    }
  };
  const source = readFileSync(new URL('../src/store/modules/auth/index.ts', import.meta.url), 'utf8').replaceAll(
    'import.meta.env',
    'testEnv'
  );
  vm.runInNewContext(
    ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText,
    context
  );
  context.exports.useAuthStore();
  return {
    store,
    storage,
    tabs,
    redirects,
    credentials,
    setRouteReset(fn) {
      resetRoutes = fn;
    },
    failLogin() {
      failed = true;
    },
    setIdentity(fn) {
      identity = fn;
    },
    failLogout() {
      logoutError = {};
    },
    get logoutCalls() {
      return logoutCalls;
    }
  };
}

test('same-account login clears old and fixed tabs and ignores the previous-page redirect', async () => {
  const { store, storage, tabs, redirects } = setup();
  await store.login('admin', 'new-password');
  assert.deepEqual(tabs.items, []);
  assert.equal(storage.has('globalTabs'), false);
  assert.equal(storage.has('lastLoginUserId'), false);
  assert.equal(storage.has('token'), true);
  assert.deepEqual(redirects, [false]);
});

test('logout removes cached tabs and authentication for a fresh login', async () => {
  const { store, storage, tabs, redirects } = setup();
  storage.set('token', 'old-token');
  storage.set('refreshToken', 'old-refresh');
  await store.resetStore();
  assert.deepEqual(tabs.items, []);
  for (const key of ['token', 'refreshToken', 'globalTabs', 'lastLoginUserId']) assert.equal(storage.has(key), false);
  assert.deepEqual(redirects, ['login']);
});

test('failed login does not enter the home page', async () => {
  const fixture = setup();
  fixture.failLogin();
  await fixture.store.login('admin', 'wrong-password');
  assert.equal(fixture.storage.has('token'), false);
  assert.equal(fixture.redirects.includes(false), false);
});

test('opaque tokens restore identity through the server', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'opaque-saved-session');
  assert.equal(await fixture.store.initUserInfo(), true);
  assert.equal(fixture.store.userInfo.userId, 'admin');
  assert.deepEqual(Array.from(fixture.store.userInfo.roles), ['R_SUPER']);
});

test('identity service unavailable preserves the stored credential', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'saved-session');
  fixture.setIdentity(async () => ({ error: { response: { status: 503 } } }));
  assert.equal(await fixture.store.initUserInfo(), false);
  assert.equal(fixture.storage.get('token'), 'saved-session');
  assert.deepEqual(fixture.redirects, []);
});

test('late identity response cannot overwrite a newer session', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'old-session');
  let resolve;
  fixture.setIdentity(
    () =>
      new Promise(done => {
        resolve = done;
      })
  );
  const pending = fixture.store.initUserInfo();
  fixture.storage.set('token', 'new-session');
  resolve({ data: { student_id: 'old-admin', role: 'admin' }, error: null });
  assert.equal(await pending, false);
  assert.equal(fixture.storage.get('token'), 'new-session');
  assert.equal(fixture.store.userInfo.userId, '');
});

test('active logout calls the backend and clears tabs only on success', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'saved-session');
  await fixture.store.logout();
  assert.equal(fixture.logoutCalls, 1);
  assert.equal(fixture.storage.has('token'), false);
  assert.deepEqual(fixture.tabs.items, []);
});

test('failed server logout preserves local login', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'saved-session');
  fixture.failLogout();
  await fixture.store.logout();
  assert.equal(fixture.storage.get('token'), 'saved-session');
  assert.deepEqual(fixture.redirects, []);
});

test('identity and logout calls bind the token captured by the store', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'captured-session');
  await fixture.store.initUserInfo();
  await fixture.store.logout();
  assert.deepEqual(fixture.credentials, ['captured-session', 'captured-session']);
});

test('pending old-session cleanup cannot navigate away from a new login', async () => {
  const fixture = setup();
  fixture.storage.set('token', 'old-session');
  let finishReset;
  fixture.setRouteReset(
    () =>
      new Promise(resolve => {
        finishReset = resolve;
      })
  );
  const pending = fixture.store.resetStore();
  fixture.storage.set('token', 'new-session');
  fixture.tabs.items = ['new-tab'];
  finishReset();
  await pending;
  assert.equal(fixture.storage.get('token'), 'new-session');
  assert.deepEqual(fixture.tabs.items, ['new-tab']);
  assert.deepEqual(fixture.redirects, []);
});
