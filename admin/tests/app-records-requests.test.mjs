import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

test('hidden record panel stops restoring later pages and can refresh after returning', async () => {
  const requests = [];
  let setActive;
  let panel;
  const context = {
    exports: {},
    uni: { getSystemInfoSync: () => ({ windowWidth: 400 }) },
    defineProps: () => ({ active: true }),
    defineExpose: exposed => { panel = exposed; },
    require: name => name === 'vue' ? {
      ref: value => ({ value }),
      onUnmounted: () => {},
      watch: (_source, callback) => { setActive = callback; callback(true); }
    } : {
      getCurrentUser: () => ({ student_id: 'student' }),
      getRecords: page => new Promise(resolve => requests.push({ page, resolve }))
    }
  };
  const source = readFileSync(new URL('../../frontend/components/RecordsTabContent.uvue', import.meta.url), 'utf8')
    .split('<script setup lang="uts">')[1].split('</script>')[0];
  vm.runInNewContext(ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText, context);
  vm.runInNewContext('currentPage.value = 3', context);
  panel.refresh();
  assert.equal(requests.length, 1);
  setActive(false);
  requests[0].resolve(Array.from({ length: 20 }, () => ({})));
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(requests.length, 1);
  panel.refresh();
  assert.equal(requests.length, 1);
  setActive(true);
  panel.refresh();
  assert.equal(requests.length, 2);
  assert.equal(requests[1].page, 1);
  requests[1].resolve([]);
  await new Promise(resolve => setImmediate(resolve));
});
