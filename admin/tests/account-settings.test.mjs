import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  const calls = [];
  const hooks = [];
  const messages = [];
  const logouts = [];
  const context = {
    exports: {},
    TextEncoder,
    window: { $message: { success: message => messages.push(message), error: message => messages.push(message) } },
    require(name) {
      if (name === 'vue') {
        return {
          computed: fn => ({ get value() { return fn(); } }),
          reactive: value => value,
          ref: value => ({ value }),
          onDeactivated: fn => hooks.push(fn),
          onUnmounted: fn => hooks.push(fn)
        };
      }
      if (name === 'vue-router') return { onBeforeRouteLeave: fn => hooks.push(fn) };
      if (name === '@vueuse/core') return { useMediaQuery: () => ({ value: false }) };
      if (name === 'naive-ui') return { useThemeVars: () => ({ value: {} }) };
      if (name.includes('store/modules/auth')) return {
        useAuthStore: () => ({ userInfo: { userName: 'admin' }, resetStore: async () => logouts.push('logout') })
      };
      if (name.includes('service/api/auth')) {
        return { changeAdminPassword: (...args) => new Promise(resolve => calls.push({ args, resolve })) };
      }
      return {};
    }
  };
  const source = readFileSync(new URL('../src/views/account-settings/index.vue', import.meta.url), 'utf8')
    .match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1];
  vm.runInNewContext(ts.transpileModule(`${source}\nexports.state = { model, saving, rules, formRef, submit, resetForm };`, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText, context);
  const state = context.exports.state;
  let clears = 0;
  state.formRef.value = {
    async validate() {
      for (const [key, rules] of Object.entries(state.rules)) {
        for (const rule of Array.isArray(rules) ? rules : [rules]) {
          if (!rule.validator(rule, state.model[key])) throw new Error('invalid');
        }
      }
    },
    restoreValidation() { clears++; }
  };
  return { state, calls, hooks, messages, logouts, get clears() { return clears; } };
}

const tick = () => new Promise(resolve => setImmediate(resolve));
const fill = state => Object.assign(state.model, { currentPassword: 'old-pass', newPassword: ' pass ', confirmPassword: ' pass ' });
const assertEmpty = state => assert.ok(Object.values(state.model).every(value => value === ''));

test('API uses the protected admin endpoint and preserves password bytes', async () => {
  let config;
  const context = { exports: {}, require: () => ({ request: input => { config = input; return Promise.resolve({ error: null, data: null }); } }) };
  const source = readFileSync(new URL('../src/service/api/auth.ts', import.meta.url), 'utf8');
  vm.runInNewContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, context);
  await context.exports.changeAdminPassword(' old ', ' 密码 ', ' 密码 ');
  assert.deepEqual(JSON.parse(JSON.stringify(config)), {
    url: '/admin/auth/password', method: 'post', data: { password: ' old ', new_password: ' 密码 ', re_password: ' 密码 ' }
  });
  await context.exports.fetchLogin('admin', ' password ');
  assert.deepEqual(JSON.parse(JSON.stringify(config)), {
    url: '/admin/auth/login', method: 'post', data: { user_name: 'admin', password: ' password ' }
  });
});

test('password boundaries use UTF-8 bytes and Go-compatible whitespace rules', async () => {
  for (const [password, valid] of [
    ['12345', false], ['123456', true], ['a'.repeat(30), true], ['a'.repeat(31), false],
    ['密码', true], ['密'.repeat(10), true], ['密'.repeat(11), false],
    ['      ', false], ['\u3000\u3000', false], ['\u0085'.repeat(3), false], ['\ufeff\ufeff', true], [' pass ', true]
  ]) {
    const { state, calls } = setup();
    Object.assign(state.model, { currentPassword: 'old-pass', newPassword: password, confirmPassword: password });
    const pending = state.submit();
    await tick();
    assert.equal(calls.length, valid ? 1 : 0);
    if (valid) calls[0].resolve({ error: null });
    await pending;
  }
  for (const values of [
    ['', '123456', '123456'], ['old', '', ''], ['old', '123456', ''],
    ['old', '123456', 'different'], ['123456', '123456', '123456']
  ]) {
    const { state, calls } = setup();
    [state.model.currentPassword, state.model.newPassword, state.model.confirmPassword] = values;
    await state.submit();
    assert.equal(calls.length, 0);
    assert.equal(state.saving.value, false);
  }
});

test('success clears inputs, duplicate submission is blocked, failure preserves inputs for retry', async () => {
  const { state, calls, messages, logouts } = setup();
  fill(state);
  let pending = state.submit();
  await tick();
  await state.submit();
  assert.equal(calls.length, 1);
  assert.deepEqual(calls[0].args, ['old-pass', ' pass ', ' pass ']);
  calls[0].resolve({ error: { message: 'request wrapper already notified' } });
  await pending;
  assert.equal(state.model.newPassword, ' pass ');
  assert.equal(state.saving.value, false);
  assert.equal(messages.length, 0);
  assert.equal(logouts.length, 0);
  pending = state.submit();
  await tick();
  calls[1].resolve({ error: null });
  await pending;
  assertEmpty(state);
  assert.equal(state.saving.value, false);
  assert.deepEqual(messages, ['密码修改成功，请重新登录']);
  assert.equal(logouts.length, 1);
});

test('reset and navigation clear credentials but successful late responses still log out', async () => {
  const fixture = setup();
  const { state, hooks, calls, messages, logouts } = fixture;
  fill(state);
  state.resetForm();
  assertEmpty(state);
  for (const leave of [state.resetForm, ...hooks]) {
    fill(state);
    const pending = state.submit();
    await tick();
    leave();
    assertEmpty(state);
    calls.at(-1).resolve({ error: null });
    await pending;
    assert.equal(messages.at(-1), '密码修改成功，请重新登录');
    assert.equal(logouts.length, messages.length);
    assert.equal(state.saving.value, false);
  }
  assert.ok(fixture.clears > 0);
  assert.equal(logouts.length, hooks.length + 1);
});

test('leaving during validation prevents a request, and late failures do not log out', async () => {
  const { state, hooks, calls, messages, logouts } = setup();
  fill(state);
  const validating = state.submit();
  hooks[0]();
  await validating;
  assert.equal(calls.length, 0);

  fill(state);
  const pending = state.submit();
  await tick();
  hooks[0]();
  calls[0].resolve({ error: { message: 'request failed' } });
  await pending;
  assertEmpty(state);
  assert.equal(state.saving.value, false);
  assert.equal(messages.length, 0);
  assert.equal(logouts.length, 0);
});
