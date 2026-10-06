import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';
import { parse } from 'vue/compiler-sfc';

const require = createRequire(import.meta.url);
const vue = require('vue');
const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8');
const compile = source =>
  ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText;
const settle = () => new Promise(resolve => setImmediate(resolve));

// Run the real API adapters and page logic with deferred HTTP responses.
function setup(kind) {
  const calls = [];
  const messages = [];
  let dialog;
  const request = config => new Promise(resolve => calls.push({ config, resolve }));
  const apiContext = { exports: {}, require: () => ({ request }) };
  vm.runInNewContext(compile(read('src/service/api/management.ts')), apiContext);
  const script = parse(read(`src/views/${kind}/index.vue`)).descriptor.scriptSetup.content;
  const exposed =
    'rows,total,loading,page,search,pagination,model,drawerVisible,formRef,applySearch,resetSearch,openForm,submit,confirmDelete';
  const context = {
    exports: {},
    require: name => {
      if (name === 'vue') return vue;
      if (name === 'naive-ui') return {};
      if (name.includes('/service/api/management')) return apiContext.exports;
      if (name.includes('/store/modules/app')) return { useAppStore: () => ({ isMobile: false }) };
      throw new Error(`unexpected import ${name}`);
    },
    window: {
      $message: { success: message => messages.push(message) },
      $dialog: {
        warning: options => {
          dialog = options;
        }
      }
    }
  };
  vm.runInNewContext(compile(`${script}\nexport { ${exposed} };`), context);
  return {
    page: context.exports,
    calls,
    messages,
    get dialog() {
      return dialog;
    }
  };
}
function row(kind, id = 1) {
  const base = { id, is_enabled: true, created_at: '2026-10-06T00:00:00Z', updated_at: '2026-10-06T00:00:00Z' };
  return kind === 'class'
    ? { ...base, college: '学院', major: '专业', class_name: '一班' }
    : { ...base, name: '病假', sort_order: 0 };
}
function list(call, kind, total = 1, id = 1) {
  call.resolve({ error: null, data: { items: total ? [row(kind, id)] : [], total, page: 1, page_size: 10 } });
}
test('management HTTP 401 clears authentication without refreshing tokens', async () => {
  let options;
  let resets = 0;
  let refreshes = 0;
  const messages = [];
  const context = {
    exports: {},
    testEnv: { DEV: false },
    window: {},
    require: name => {
      if (name === '@sa/axios')
        return {
          BACKEND_ERROR_CODE: 'BACKEND_ERROR',
          createFlatRequest: (_config, opts) => {
            options = opts;
            return { state: opts.defaultState };
          },
          createRequest: () => ({})
        };
      if (name.includes('/store/modules/auth'))
        return {
          useAuthStore: () => ({
            resetStore: async () => {
              resets++;
            }
          })
        };
      if (name.includes('/utils/service')) return { getServiceBaseURL: () => ({ baseURL: '', otherBaseURL: {} }) };
      if (name.includes('/utils/storage')) return {};
      if (name.includes('/locales')) return { $t: value => value };
      if (name === './shared')
        return {
          getAuthorization: () => null,
          handleExpiredRequest: async () => {
            refreshes++;
            return true;
          },
          showErrorMsg: (_state, message) => messages.push(message)
        };
      throw new Error(`unexpected import ${name}`);
    }
  };
  vm.runInNewContext(compile(read('src/service/request/index.ts').replaceAll('import.meta.env', 'testEnv')), context);
  await options.onError({
    message: 'HTTP 401',
    config: { url: '/admin/classes' },
    response: { status: 401, data: { code: 1005, message: 'Token expired' } }
  });
  assert.equal(resets, 1);
  assert.equal(refreshes, 0);
  assert.equal(messages[0], 'Token expired');
  await options.onError({
    message: 'HTTP 401',
    config: { url: '/admin/auth/login' },
    response: { status: 401, data: { code: 1003, message: 'Invalid password' } }
  });
  assert.equal(resets, 1);
  assert.equal(messages[1], 'Invalid password');
});
for (const kind of ['class', 'leave-type']) {
  test(`${kind}: pagination, filters and stale responses`, async () => {
    const app = setup(kind),
      page = app.page;
    assert.equal(app.calls[0].config.params.page_size, 10);
    list(app.calls[0], kind, 12);
    await settle();
    assert.equal(page.total.value, 12);
    assert.equal(page.rows.value.length, 1);
    assert.equal(page.rows.value[0].isEnabled, true);
    page.search.isEnabled = false;
    if (kind === 'class') page.search.className = ' 二班 ';
    else page.search.name = ' 事假 ';
    page.applySearch();
    assert.equal(app.calls[1].config.params.is_enabled, false);
    assert.equal(
      app.calls[1].config.params[kind === 'class' ? 'class_name' : 'name'],
      kind === 'class' ? '二班' : '事假'
    );
    page.pagination.value.onUpdatePage(2);
    assert.equal(app.calls[2].config.params.page, 2);
    list(app.calls[2], kind, 12, 3);
    await settle();
    list(app.calls[1], kind, 1, 2);
    await settle();
    assert.equal(page.rows.value[0].id, 3);
    page.resetSearch();
    assert.equal(app.calls[3].config.params.page, 1);
    assert.equal(app.calls[3].config.params.is_enabled, undefined);
    list(app.calls[3], kind, 0);
    await settle();
    assert.equal(page.total.value, 0);
    assert.equal(page.rows.value.length, 0);
    assert.equal(page.loading.value, false);
  });
  test(`${kind}: save failures, repeated submission and last-page deletion`, async () => {
    const app = setup(kind),
      page = app.page;
    list(app.calls[0], kind, 11);
    await settle();
    await page.openForm(page.rows.value[0]);
    const original = kind === 'class' ? page.rows.value[0].className : page.rows.value[0].name;
    if (kind === 'class') page.model.className = 'Edited';
    else page.model.name = 'Edited';
    assert.equal(kind === 'class' ? page.rows.value[0].className : page.rows.value[0].name, original);
    page.formRef.value = { validate: async () => {}, restoreValidation() {} };
    const pendingSave = page.submit();
    const repeatedSave = page.submit();
    await settle();
    await repeatedSave;
    assert.equal(app.calls[1].config.method, 'put');
    await page.submit();
    assert.equal(app.calls.length, 2);
    app.calls[1].resolve({ data: null, error: new Error('conflict') });
    await pendingSave;
    assert.equal(page.drawerVisible.value, true);
    assert.equal(app.messages.length, 0);
    page.model.isEnabled = false;
    const successfulSave = page.submit();
    await settle();
    assert.equal(app.calls[2].config.data.is_enabled, false);
    app.calls[2].resolve({ data: row(kind), error: null });
    await settle();
    list(app.calls[3], kind, 11);
    await successfulSave;
    assert.equal(page.drawerVisible.value, false);
    page.pagination.value.onUpdatePage(2);
    list(app.calls[4], kind, 11);
    await settle();
    page.confirmDelete(page.rows.value[0]);
    let pendingDelete = app.dialog.onPositiveClick();
    app.calls[5].resolve({ data: null, error: new Error('in use') });
    assert.equal(await pendingDelete, false);
    pendingDelete = app.dialog.onPositiveClick();
    app.calls[6].resolve({ data: null, error: null });
    await settle();
    list(app.calls[7], kind, 10);
    await settle();
    assert.equal(app.calls[8].config.params.page, 1);
    list(app.calls[8], kind, 10);
    await pendingDelete;
    assert.equal(page.page.value, 1);
    assert.equal(page.total.value, 10);
    assert.equal(page.loading.value, false);
  });
}
