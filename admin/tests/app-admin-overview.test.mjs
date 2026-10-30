import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  const calls = [];
  let token = 'first';
  let cleanup;
  const context = {
    exports: {},
    uni: { getStorageSync: () => token },
    require: () => ({
      getCurrentUser: () => ({ student_id: 'admin' }),
      registerSessionCleanup: callback => { cleanup = callback; },
      request: path => new Promise((resolve, reject) => calls.push({ path, resolve, reject }))
    })
  };
  const source = readFileSync(new URL('../../frontend/utils/admin-api.uts', import.meta.url), 'utf8');
  vm.runInNewContext(ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText, context);
  return { api: context.exports, calls, switchSession: () => { token = 'second'; }, cleanup: () => cleanup() };
}

function response(count) {
  return {
    statistics: { student_count: 2, record_count: 3, unhandled_feedback_count: count },
    weekly_applications: { weeks: [], total: 0 },
    version_usage: { groups: [], total: 0, latest_count: 0 },
    latest_android_version: null
  };
}

test('App overview and more share one request and the completed snapshot', async () => {
  const { api, calls } = setup();
  const overview = api.adminOverview();
  const more = api.adminOverview();
  assert.equal(calls.length, 1);
  assert.equal(calls[0].path, '/admin/overview');
  calls[0].resolve(response(7));
  const data = await overview;
  assert.equal(data.unhandledFeedbackCount, 7);
  assert.equal(await more, data);
  assert.equal(await api.adminOverview(), data);
  assert.equal(calls.length, 1);
});

test('explicit refresh bypasses completed cache and shares an in-flight refresh', async () => {
  const { api, calls } = setup();
  const initial = api.adminOverview();
  calls[0].resolve(response(7));
  await initial;
  const refresh = api.adminOverview(true);
  const repeated = api.adminOverview(true);
  assert.equal(calls.length, 2);
  calls[1].resolve(response(0));
  assert.equal((await refresh).unhandledFeedbackCount, 0);
  assert.equal((await repeated).unhandledFeedbackCount, 0);
});

test('failed request releases the pending task and retry fetches again', async () => {
  const { api, calls } = setup();
  const failed = api.adminOverview();
  calls[0].reject(new Error('offline'));
  await assert.rejects(failed, /offline/);
  const retry = api.adminOverview();
  calls[1].resolve(response(1));
  assert.equal((await retry).unhandledFeedbackCount, 1);
});

for (const change of ['revision', 'session', 'logout']) {
  test(`late response cannot overwrite cache after ${change} changes`, async () => {
    const { api, calls, switchSession, cleanup } = setup();
    const old = api.adminOverview();
    if (change === 'revision') api.invalidateAdmin();
    if (change === 'session') switchSession();
    if (change === 'logout') cleanup();
    const current = api.adminOverview();
    assert.equal(calls.length, 2);
    calls[0].resolve(response(99));
    await old;
    const shared = api.adminOverview();
    assert.equal(calls.length, 2);
    calls[1].resolve(response(2));
    assert.equal((await current).unhandledFeedbackCount, 2);
    await shared;
    assert.equal((await api.adminOverview()).unhandledFeedbackCount, 2);
  });
}
