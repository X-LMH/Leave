import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  let currentToken = 'new-session';
  let resets = 0;
  let options;
  const messages = [];
  const context = {
    exports: {},
    testEnv: {},
    require(name) {
      if (name === '@sa/axios')
        return {
          createFlatRequest: (_config, hooks) => {
            options = hooks;
            return { state: hooks.defaultState };
          },
          createRequest: () => ({}),
          BACKEND_ERROR_CODE: 'BACKEND_ERROR'
        };
      if (name === '@/store/modules/auth')
        return {
          useAuthStore: () => ({
            resetStore: async () => {
              resets++;
              currentToken = '';
            }
          })
        };
      if (name === '@/utils/storage') return { localStg: { get: () => currentToken } };
      if (name === '@/utils/service') return { getServiceBaseURL: () => ({ baseURL: '/api/v1', otherBaseURL: {} }) };
      if (name === './shared')
        return {
          getAuthorization: () => (currentToken ? `Bearer ${currentToken}` : null),
          isCurrentAuthorization: value => Boolean(value) && value === `Bearer ${currentToken}`,
          showErrorMsg: (_state, message) => messages.push(message)
        };
      throw new Error(`Unexpected import: ${name}`);
    }
  };
  const source = readFileSync(new URL('../src/service/request/index.ts', import.meta.url), 'utf8').replaceAll(
    'import.meta.env',
    'testEnv'
  );
  vm.runInNewContext(
    ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText,
    context
  );
  return {
    options,
    messages,
    get token() {
      return currentToken;
    },
    get resets() {
      return resets;
    }
  };
}

function response(token, status = 401, code = 1005, url = '/admin/overview') {
  return {
    status,
    data: { code, message: 'test failure' },
    config: { url, headers: { Authorization: `Bearer ${token}` } }
  };
}

test('an old unauthorized response cannot clear the new login', async () => {
  const fixture = setup();
  const old = response('old-session');
  await fixture.options.onBackendFail(old);
  await fixture.options.onError({ response: old, config: old.config });
  assert.equal(fixture.resets, 0);
  assert.equal(fixture.token, 'new-session');
});

test('current unauthorized responses clear login without refresh or retry', async () => {
  const fixture = setup();
  await fixture.options.onBackendFail(response('new-session'));
  assert.equal(fixture.resets, 1);
  assert.equal(fixture.token, '');
});

test('HTTP authentication errors are handled when the backend-fail hook is bypassed', async () => {
  const fixture = setup();
  const invalid = response('new-session');
  await fixture.options.onError({ response: invalid, config: invalid.config });
  assert.equal(fixture.resets, 1);
});

test('authentication infrastructure errors preserve the credential', async () => {
  const fixture = setup();
  const unavailable = response('new-session', 503, 9003);
  await fixture.options.onBackendFail(unavailable);
  await fixture.options.onError({ response: unavailable, config: unavailable.config });
  assert.equal(fixture.resets, 0);
  assert.equal(fixture.token, 'new-session');
  assert.deepEqual(fixture.messages, ['test failure']);
});

test('wrong login credentials cannot invalidate an existing session', async () => {
  const fixture = setup();
  const denied = response('new-session', 401, 1003, '/admin/auth/login');
  await fixture.options.onBackendFail(denied);
  await fixture.options.onError({ response: denied, config: denied.config });
  assert.equal(fixture.resets, 0);
});

test('queued session operations retain their explicitly captured authorization', async () => {
  const fixture = setup();
  const request = await fixture.options.onRequest({ headers: { Authorization: 'Bearer old-session' } });
  assert.equal(request.headers.Authorization, 'Bearer old-session');
  const ordinary = await fixture.options.onRequest({ headers: {} });
  assert.equal(ordinary.headers.Authorization, 'Bearer new-session');
});
