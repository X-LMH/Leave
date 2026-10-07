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
  assert.equal(Object.keys(row).length, 10);
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
