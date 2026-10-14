import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  const storage = new Map([['globalTabs', ['student', 'fixed-tab']], ['lastLoginUserId', 'admin']]);
  const tabs = { items: ['student', 'fixed-tab'], resetTabs() { this.items = []; } };
  const redirects = [];
  let failed = false;
  const store = { $reset() {} };
  let initialized = false;
  const token = `header.${Buffer.from(JSON.stringify({ student_id: 'admin', role: 'admin', exp: 9999999999 })).toString('base64url')}.signature`;
  const context = {
    exports: {}, atob, testEnv: {},
    window: { $notification: { success() {} } },
    require(name) {
      if (name === 'vue') return { computed: fn => ({ get value() { return fn(); } }), reactive: value => value, ref: value => ({ value }) };
      if (name === 'vue-router') return { useRoute: () => ({ meta: {} }) };
      if (name === 'pinia') return { defineStore: (_id, create) => () => {
        if (!initialized) {
          initialized = true;
          Object.assign(store, create());
        }
        return store;
      } };
      if (name === '@sa/hooks') return { useLoading: () => ({ startLoading() {}, endLoading() {} }) };
      if (name === '@/service/api') return { fetchLogin: async () => ({ data: { token }, error: failed ? {} : null }) };
      if (name === '@/hooks/common/router') return { useRouterPush: () => ({ toLogin: async () => redirects.push('login'), redirectFromLogin: async redirect => redirects.push(redirect) }) };
      if (name === '@/utils/storage') return { localStg: { get: key => storage.get(key), set: (key, value) => storage.set(key, value), remove: key => storage.delete(key) } };
      if (name === '@/enum') return { SetupStoreId: { Auth: 'auth' } };
      if (name === '@/locales') return { $t: value => value };
      if (name === '../route') return { useRouteStore: () => ({ resetStore: async () => {} }) };
      if (name === '../tab') return { useTabStore: () => tabs };
      if (name === './shared') return { clearAuthStorage: () => { storage.delete('token'); storage.delete('refreshToken'); }, getToken: () => storage.get('token') };
      throw new Error(`Unexpected import: ${name}`);
    }
  };
  const source = readFileSync(new URL('../src/store/modules/auth/index.ts', import.meta.url), 'utf8').replaceAll('import.meta.env', 'testEnv');
  vm.runInNewContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, context);
  context.exports.useAuthStore();
  return { store, storage, tabs, redirects, failLogin() { failed = true; } };
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
