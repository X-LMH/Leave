import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function script(name) {
  return readFileSync(new URL(`../../frontend/components/${name}.uvue`, import.meta.url), 'utf8')
    .split('<script setup lang="uts">')[1].split('</script>')[0];
}

function setup() {
  const props = { kind: 'students', active: true, height: 600, pendingKey: 0 };
  const calls = [];
  let status = '启用';
  let activate;
  const context = {
    exports: {},
    defineProps: () => props,
    defineExpose() {},
    require: name => name === 'vue' ? {
      ref: value => ({ value }),
      computed: fn => ({ get value() { return fn(); } }),
      onUnmounted() {},
      nextTick: fn => fn(),
      watch: (_source, callback, options) => {
        if (options?.immediate) activate = callback;
      }
    } : {
      adminRevision: 0,
      adminSession: () => 'admin-session',
      requireAdmin: async () => true,
      adminOptions: async () => [{ id: 1, label: '计24-1' }],
      adminList: async (_kind, query) => {
        calls.push(query);
        const visible = !query.includes('status=1') || status === '启用';
        return {
          rows: visible ? [{ id: 1, classID: 1, status }] : [],
          total: visible ? 1 : 0
        };
      }
    }
  };
  const source = script('AdminListContent') + '\nglobalThis.panel = { rows, total, draftState, applyFilters, rememberScroll, scrollTop };';
  vm.runInNewContext(ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText, context);
  return {
    panel: context.panel,
    calls,
    setStatus(value) { status = value; },
    async active(value) {
      props.active = value;
      activate(value);
      await new Promise(resolve => setImmediate(resolve));
    }
  };
}

test('returning from student detail reloads status without manual refresh and preserves scroll', async () => {
  const app = setup();
  await app.active(true);
  assert.equal(app.panel.rows.value[0].status, '启用');
  app.panel.rememberScroll({ detail: { scrollTop: 180 } });
  await app.active(false);
  app.setStatus('禁用');
  await app.active(true);
  assert.equal(app.panel.rows.value[0].status, '禁用');
  assert.equal(app.panel.scrollTop.value, 180);
  assert.equal(app.calls.length, 2);
  await app.active(false);
  app.setStatus('启用');
  await app.active(true);
  assert.equal(app.panel.rows.value[0].status, '启用');
});

test('returning from detail reapplies account filter and removes a newly disabled student', async () => {
  const app = setup();
  await app.active(true);
  app.panel.draftState.value = 1;
  app.panel.applyFilters();
  await new Promise(resolve => setImmediate(resolve));
  await app.active(false);
  app.setStatus('禁用');
  await app.active(true);
  assert.ok(app.calls.at(-1).includes('status=1'));
  assert.equal(app.panel.rows.value.length, 0);
  assert.equal(app.panel.total.value, 0);
});

test('student detail renders profile, residence, family and usage data with short class names', () => {
  const context = {
    defineProps: () => ({ kind: 'students', active: true }),
    ref: value => ({ value }),
    computed: fn => ({ get value() { return fn(); } }),
    watch() {},
    onUnmounted() {},
    stringValue: (data, key) => data[key] == null ? '' : String(data[key]),
    displayDate: value => value
  };
  const source = script('AdminDetailContent').replace(/import[\s\S]*?from\s+['"][^'"]+['"]/g, '')
    + '\nglobalThis.detail = { showDetail, options, apartments, detailGroups };';
  vm.runInNewContext(ts.transpileModule(source, {
    compilerOptions: { target: ts.ScriptTarget.ES2022 }
  }).outputText, context);
  const detail = context.detail;
  detail.options.value = [{ id: 1, label: '计24-1' }];
  detail.apartments.value = [{ id: 2, label: '一号公寓' }];
  detail.showDetail({
    status: 1, class_id: 1, gender: 'female', teacher_name: '老师',
    phone: '123', apartment_id: 2, dormitory_number: '201',
    parent_name: '家长', parent_phone: '456', app_version: '1.2.3',
    app_version_status: 'latest', last_seen_device: 'Android',
    last_seen_at: '最近使用时间', created_at: '注册时间', updated_at: '更新时间'
  });
  const groups = detail.detailGroups.value;
  assert.deepEqual(Array.from(groups, group => group.title), ['基本信息', '联系与住宿', '使用信息']);
  const fields = Object.fromEntries(groups.flatMap(group => Array.from(group.fields, field => [field.label, field.value])));
  assert.equal(fields['班级'], '计24-1');
  assert.equal(fields['性别'], '女');
  assert.equal(fields['辅导员'], '老师');
  assert.equal(fields['公寓'], '一号公寓');
  assert.equal(fields['宿舍号'], '201');
  assert.equal(fields['家长电话'], '456');
  assert.equal(fields['联系电话'], '123');
  assert.equal(fields['最近设备'], 'Android');
});
