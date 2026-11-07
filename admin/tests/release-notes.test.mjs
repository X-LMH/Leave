import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function loadModule(path, additions = '', globals = {}) {
  const context = { exports: {}, require: () => ({}), process: { env: {} }, ...globals };
  const source = readFileSync(new URL(path, import.meta.url), 'utf8') + additions;
  vm.runInNewContext(
    ts.transpileModule(source, {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
    }).outputText,
    context
  );
  return context.exports;
}
const admin = loadModule('../src/service/api/app-version.ts');
const storage = new Map();
const app = loadModule('../../frontend/utils/api.uts', '\nexport { appVersionInfoFromJSON };', {
  uni: { setStorageSync: (key, value) => storage.set(key, value), getStorageSync: key => storage.get(key) }
});
const plain = value => JSON.parse(JSON.stringify(value));
const grouped = { features: ['新增'], improvements: ['优化'], bug_fixes: ['修复'], notices: ['重新登录'] };
function payload(notes, categories) {
  return {
    platform: 'android',
    latest_version_code: 17,
    latest_version_name: '1.2.0',
    updates: [
      {
        version_code: 17,
        version_name: '1.2.0',
        release_notes: notes,
        ...(categories === undefined ? {} : { release_notes_grouped: categories })
      }
    ]
  };
}

test('App and admin display categorized entries once in a fixed order', () => {
  const update = app.appVersionInfoFromJSON(payload(['展开兼容内容'], grouped)).updates[0];
  const sections = plain(app.releaseNoteSections(update));
  assert.deepEqual(
    sections.map(s => s.title),
    ['新增功能', '优化', '问题修复', '更新须知']
  );
  assert.deepEqual(
    sections.flatMap(s => s.items),
    ['新增', '优化', '修复', '重新登录']
  );
  assert.deepEqual(plain(admin.releaseNoteSections(['展开兼容内容'], grouped)), sections);
});

test('historical wire payload and cached versions retain legacy notes', () => {
  const version = app.appVersionInfoFromJSON(payload(['旧说明']));
  app.saveUpdateInfo(version);
  const restored = app.getStoredUpdateInfo();
  assert.deepEqual(plain(app.releaseNoteSections(restored.updates[0])), [{ title: '更新内容', items: ['旧说明'] }]);
  storage.set('leave_update_info', JSON.stringify(payload(['旧缓存'])));
  assert.equal(app.getStoredUpdateInfo().updates[0].release_notes_grouped, null);
  assert.deepEqual(plain(admin.releaseNoteSections(['旧说明'])), [{ title: '更新内容', items: ['旧说明'] }]);
});

test('categorized cache round-trips and missing categories normalize to empty lists', () => {
  const version = app.appVersionInfoFromJSON(payload([], { notices: ['重新登录'] }));
  app.saveUpdateInfo(version);
  const restored = app.getStoredUpdateInfo();
  assert.deepEqual(plain(app.releaseNoteSections(restored.updates[0])), [{ title: '更新须知', items: ['重新登录'] }]);
});

test('empty notes hide all sections rather than claiming fixes', () => {
  assert.deepEqual(plain(app.releaseNoteSections(app.appVersionInfoFromJSON(payload([])).updates[0])), []);
  assert.deepEqual(plain(app.releaseNoteSections(app.appVersionInfoFromJSON(payload([], {})).updates[0])), []);
  assert.deepEqual(
    plain(admin.releaseNoteSections([], { features: [], improvements: [], bug_fixes: [], notices: [] })),
    []
  );
});
