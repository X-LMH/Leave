import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function setup() {
  const calls = [];
  const request = config => new Promise(resolve => calls.push({ config, resolve }));
  const context = { exports: {}, require: () => ({ request }) };
  const source = readFileSync(new URL('../src/service/api/student.ts', import.meta.url), 'utf8');
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText;
  vm.runInNewContext(compiled, context);
  return { api: context.exports, calls };
}

const listRow = {
  id: 7,
  student_id: '202600010001',
  name: '',
  class_id: 0,
  gender: '',
  phone: '',
  status: 1,
  app_version: '1.0.0',
  app_version_status: 'latest',
  created_at: '2026-10-07T00:00:00Z',
  last_seen_at: null
};

test('student list accepts only table fields without fabricating detail data', async () => {
  const { api, calls } = setup();
  const pending = api.queryStudents({ page: 1, pageSize: 10, studentId: '', name: '', classId: null, status: 0 });
  assert.equal(calls[0].config.params.status, 0);
  calls[0].resolve({ error: null, data: { items: [listRow], total: 1, page: 1, page_size: 10 } });
  const { data, error } = await pending;
  assert.equal(error, null);
  assert.equal(data.total, 1);
  const row = data.items[0];
  assert.equal(row.studentId, listRow.student_id);
  assert.equal(row.classId, null);
  assert.equal(row.gender, null);
  assert.equal(row.lastSeenAt, null);
  assert.equal(row.createdAt, Date.parse(listRow.created_at));
  assert.equal(row.appVersionStatus, listRow.app_version_status);
  assert.equal(Object.keys(row).length, 11);
  for (const field of ['parentName', 'parentPhone', 'teacherName', 'avatarUrl', 'apartmentId', 'dormitoryNumber', 'lastSeenDevice', 'updatedAt']) {
    assert.equal(Object.hasOwn(row, field), false);
  }
});

test('student detail continues to load complete profile fields by account id', async () => {
  const { api, calls } = setup();
  const pending = api.getStudent(7);
  assert.equal(calls[0].config.url, '/admin/students/7');
  calls[0].resolve({
    error: null,
    data: {
      ...listRow,
      parent_name: '家长',
      parent_phone: '13900139000',
      teacher_name: '老师',
      apartment_id: 3,
      dormitory_number: '101',
      avatar_url: '/uploads/avatars/student.png',
      updated_at: listRow.created_at,
      last_seen_device: 'Android'
    }
  });
  const { data, error } = await pending;
  assert.equal(error, null);
  assert.equal(data.parentName, '家长');
  assert.equal(data.parentPhone, '13900139000');
  assert.equal(data.teacherName, '老师');
  assert.equal(data.apartmentId, 3);
  assert.equal(data.dormitoryNumber, '101');
  assert.equal(data.avatarUrl, '/uploads/avatars/student.png');
  assert.equal(data.lastSeenDevice, 'Android');
  assert.equal(data.updatedAt, Date.parse(listRow.created_at));
});


test('reset password sends only the selected account ID', async () => {
  const { api, calls } = setup();
  const pending = api.resetStudentPassword(7);
  assert.equal(calls[0].config.url, '/admin/students/7/reset-password');
  assert.equal(calls[0].config.method, 'post');
  assert.equal(calls[0].config.data, undefined);
  calls[0].resolve({ error: null, data: null });
  assert.equal((await pending).error, null);
});

function setupReset() {
  const calls = [], dialogs = [], messages = [];
  const source = readFileSync(new URL('../src/views/student/index.vue', import.meta.url), 'utf8');
  const handler = source.slice(source.indexOf('function confirmResetPassword('), source.indexOf('function confirmStatus('));
  const state = Object.fromEntries(['editing', 'opening', 'saving', 'resettingPassword'].map(key => [key, { value: false }]));
  state.changingId = { value: null };
  const context = {
    ...state,
    resetStudentPassword: id => new Promise((resolve, reject) => calls.push({ id, resolve, reject })),
    window: { $dialog: { warning: config => dialogs.push(config) }, $message: {
      success: value => messages.push(value), error: value => messages.push(value)
    } }
  };
  vm.runInNewContext(ts.transpileModule(handler, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText, context);
  return { state, calls, dialogs, messages, confirm: context.confirmResetPassword };
}

test('reset requires confirmation and blocks duplicate requests until completion', async () => {
  const { state, calls, dialogs, messages, confirm } = setupReset();
  const row = { id: 7, studentId: '202600010001' };
  confirm(row);
  assert.equal(calls.length, 0);
  assert.ok(dialogs[0].content.includes(row.studentId));
  assert.ok(dialogs[0].content.includes('12345678'));
  const pending = dialogs[0].onPositiveClick();
  row.id = 8;
  assert.equal(calls[0].id, 7);
  assert.equal(state.resettingPassword.value, true);
  assert.equal(await dialogs[0].onPositiveClick(), false);
  confirm(row);
  assert.equal(dialogs.length, 1);
  calls[0].resolve({ error: null });
  assert.equal(await pending, true);
  assert.equal(state.resettingPassword.value, false);
  assert.equal(messages[0], '\u5bc6\u7801\u5df2\u91cd\u7f6e\u4e3a 12345678');
});

test('reset failures release the busy state and allow retry', async () => {
  for (const throws of [false, true]) {
    const { state, calls, dialogs, confirm } = setupReset();
    confirm({ id: 7, studentId: 'student' });
    const pending = dialogs[0].onPositiveClick();
    if (throws) calls[0].reject(new Error('offline'));
    else calls[0].resolve({ error: {} });
    assert.equal(await pending, false);
    assert.equal(state.resettingPassword.value, false);
    const retry = dialogs[0].onPositiveClick();
    assert.equal(calls.length, 2);
    calls[1].resolve({ error: null });
    assert.equal(await retry, true);
  }
});
