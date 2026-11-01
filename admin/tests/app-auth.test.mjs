import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

// Exercise shared request logic with uni callbacks; this does not replace Android compilation.
function setup() {
  const storage = new Map([
    ['leave_token', 'old-session'],
    ['leave_student_id', 'student']
  ]);
  const requests = [];
  const uploads = [];
  const redirects = [];
  const context = {
    exports: {},
    process: { env: { NODE_ENV: 'development' } },
    uni: {
      getStorageSync: key => storage.get(key) ?? '',
      setStorageSync: (key, value) => storage.set(key, value),
      removeStorageSync: key => storage.delete(key),
      request: options => requests.push(options),
      uploadFile: options => uploads.push(options),
      reLaunch: options => redirects.push(options.url)
    },
    require(name) {
      if (name === '@/utils/device.uts') return { getAppVersionCode: () => 1 };
      if (name === '@/utils/avatar-cache.uts') return { cacheAvatarFile() {}, clearAvatarFileCache() {} };
      throw new Error(`Unexpected import: ${name}`);
    }
  };
  const source = readFileSync(new URL('../../frontend/utils/api.uts', import.meta.url), 'utf8');
  vm.runInNewContext(
    ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText,
    context
  );
  return { api: context.exports, storage, requests, uploads, redirects };
}

for (const path of ['/profile', '/admin/overview']) {
  test(`late authentication failure at ${path} preserves a newer App session`, async () => {
    const fixture = setup();
    const pending = fixture.api.request(path, 'GET', null, true);
    fixture.storage.set('leave_token', 'new-session');
    fixture.requests[0].success({ statusCode: 401, data: { code: 1005, message: 'invalid' } });
    await assert.rejects(pending);
    assert.equal(fixture.storage.get('leave_token'), 'new-session');
    assert.deepEqual(fixture.redirects, []);
  });
}

test('current App authentication failure clears login', async () => {
  const fixture = setup();
  const pending = fixture.api.request('/profile', 'GET', null, true);
  fixture.requests[0].success({ statusCode: 401, data: { code: 1005, message: 'invalid' } });
  await assert.rejects(pending);
  assert.equal(fixture.storage.has('leave_token'), false);
  assert.deepEqual(fixture.redirects, ['/pages/auth/login']);
});

test('App infrastructure failures retain the session', async () => {
  const fixture = setup();
  const pending = fixture.api.request('/profile', 'GET', null, true);
  fixture.requests[0].success({ statusCode: 503, data: { code: 9003, message: 'unavailable' } });
  await assert.rejects(pending);
  assert.equal(fixture.storage.get('leave_token'), 'old-session');
  assert.deepEqual(fixture.redirects, []);
});

test('late upload authentication failure preserves a newer session', async () => {
  const fixture = setup();
  const pending = fixture.api.uploadAvatar('/test-avatar.jpg');
  fixture.storage.set('leave_token', 'new-session');
  fixture.uploads[0].success({ statusCode: 401, data: JSON.stringify({ code: 1005, message: 'invalid' }) });
  await assert.rejects(pending);
  assert.equal(fixture.storage.get('leave_token'), 'new-session');
  assert.deepEqual(fixture.redirects, []);
});

test('partial registration success is exposed without saving credentials', async () => {
  const fixture = setup();
  fixture.storage.delete('leave_token');
  const pending = fixture.api.register('202410201016', '123456', '123456', 'test', '1.0.0');
  fixture.requests[0].success({ statusCode: 503, data: { code: 1008, message: '注册成功，请重新登录' } });
  await assert.rejects(pending, error => error.code === 1008);
  assert.equal(fixture.storage.has('leave_token'), false);
  assert.equal(fixture.requests.length, 1);
});
