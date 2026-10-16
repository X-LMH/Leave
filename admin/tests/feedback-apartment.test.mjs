import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';
import * as vue from 'vue';
import { parse } from 'vue/compiler-sfc';

function loadApi(kind, request) {
  const source = readFileSync(new URL(`../src/service/api/${kind}.ts`, import.meta.url), 'utf8');
  const compiled = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022
    }
  }).outputText;
  const context = {
    exports: {},
    require: name => {
      assert.equal(name, '../request');
      return { request };
    }
  };
  vm.runInNewContext(compiled, context);
  return context.exports;
}
const apartmentQuery = {
  name: ' 楼 ',
  gender: null,
  isEnabled: false,
  page: 2,
  pageSize: 10
};
const feedbackQuery = {
  keyword: ' 建议 ',
  studentId: ' 2024 ',
  isAnonymous: false,
  isHandled: false,
  page: 1,
  pageSize: 10
};
const apartmentWire = {
  id: 1,
  name: '测试楼',
  gender: 'male',
  sort_order: 0,
  is_enabled: false,
  created_at: '2026-10-10T00:00:00Z',
  updated_at: '2026-10-10T00:00:00Z'
};
const feedbackWire = {
  id: 1,
  student_id: '2024',
  is_anonymous: true,
  is_handled: false,
  attachments: [],
  content: '一行\n二行',
  created_at: '2026-10-10T00:00:00Z',
  updated_at: '2026-10-10T00:00:00Z'
};

test('apartment adapter maps list and preserves zero and false in save bodies', async () => {
  const calls = [];
  const api = loadApi('apartment', async options => {
    calls.push(options);
    return {
      data: options.method ? apartmentWire : { items: [apartmentWire], total: 1 },
      error: null
    };
  });
  const result = await api.queryApartments(apartmentQuery);
  assert.equal(calls[0].url, '/admin/apartments');
  assert.equal(calls[0].params.name, '楼');
  assert.equal(calls[0].params.is_enabled, false);
  assert.equal(calls[0].params.gender, undefined);
  assert.equal(calls[0].params.page_size, 10);
  assert.equal(result.data.items[0].sortOrder, 0);
  assert.equal(result.data.items[0].isEnabled, false);
  const input = {
    name: ' 楼 ',
    gender: 'male',
    sortOrder: 0,
    isEnabled: false
  };
  await api.saveApartment(input);
  await api.saveApartment(input, 1);
  assert.equal(calls[1].method, 'post');
  assert.equal(calls[2].method, 'put');
  assert.equal(calls[2].data.sort_order, 0);
  assert.equal(calls[2].data.is_enabled, false);
  assert.equal(calls[2].url, '/admin/apartments/1');
});
test('feedback adapter sends filters and masks anonymous identity', async () => {
  const calls = [];
  const api = loadApi('feedback', async options => {
    calls.push(options);
    return {
      data: options.url.endsWith('/1') ? feedbackWire : { items: [feedbackWire], total: 1 },
      error: null
    };
  });
  const result = await api.queryFeedbacks(feedbackQuery);
  const params = calls[0].params;
  assert.equal(params.keyword, '建议');
  assert.equal(params.student_id, '2024');
  assert.equal(params.is_anonymous, false);
  assert.equal(params.is_handled, false);
  assert.equal('created_at_from' in params, false);
  assert.equal('created_at_to' in params, false);
  assert.equal(result.data.items[0].studentId, null);
  const detail = await api.getFeedback(1);
  assert.equal(detail.data.studentId, null);
  assert.equal(detail.data.content, '一行\n二行');
});
for (const kind of ['apartment', 'feedback']) {
  test(`${kind} adapter propagates request failures and delete endpoint`, async () => {
    const failure = { data: null, error: { message: '失败' } };
    let options;
    const api = loadApi(kind, async input => {
      options = input;
      return failure;
    });
    assert.equal(
      await (kind === 'apartment' ? api.queryApartments(apartmentQuery) : api.queryFeedbacks(feedbackQuery)),
      failure
    );
    assert.equal(await (kind === 'apartment' ? api.deleteApartment(9) : api.deleteFeedback(9)), failure);
    assert.equal(options.method, 'delete');
    assert.equal(options.url, `/admin/${kind === 'apartment' ? 'apartments' : 'feedbacks'}/9`);
  });
}
const settle = () => new Promise(resolve => setImmediate(resolve));

function loadPage(kind, api) {
  const source = readFileSync(new URL(`../src/views/${kind}/index.vue`, import.meta.url), 'utf8');
  const script = parse(source).descriptor.scriptSetup.content;
  const exposed =
    kind === 'feedback'
      ? 'rows,total,page,pageSize,pagination,search,applySearch,resetSearch,confirmDelete,openDetail,detail,drawerVisible,loading,toggleHandling,handling,attachmentPreviews,attachmentsLoading'
      : 'rows,total,page,pageSize,pagination,search,applySearch,resetSearch,confirmDelete,openForm,model,formRef,submit,saving,drawerVisible,loading';
  let dialog;
  const errors = [];
  const urls = { created: [], revoked: [] };
  const unmountHooks = [];
  const context = {
    URL: {
      createObjectURL: blob => {
        const url = `blob:test-${urls.created.length}`;
        urls.created.push({ url, blob });
        return url;
      },
      revokeObjectURL: url => urls.revoked.push(url)
    },
    Error,
    exports: {},
    require: name => {
      if (name === 'vue') return { ...vue, onBeforeUnmount: hook => unmountHooks.push(hook) };
      if (name === 'naive-ui') return { useThemeVars: () => vue.ref({}) };
      if (name.includes('/service/api/')) return api;
      if (name.includes('/store/modules/app')) return { useAppStore: () => ({ isMobile: false }) };
      throw new Error(`Unexpected import: ${name}`);
    },
    window: {
      $message: { success: () => {}, error: message => errors.push(message) },
      $dialog: {
        warning: options => {
          dialog = options;
        }
      }
    }
  };
  const compiled = ts.transpileModule(`${script}\nexport { ${exposed} };`, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022
    }
  }).outputText;
  vm.runInNewContext(compiled, context);
  return {
    page: context.exports,
    urls,
    unmount: () => unmountHooks.forEach(hook => hook()),
    errors,
    get dialog() {
      return dialog;
    }
  };
}

function pageApi(kind) {
  const rows = Array.from({ length: 3 }, (_, i) =>
    kind === 'apartment'
      ? {
          id: i + 1,
          name: `楼${i}`,
          gender: 'male',
          sortOrder: i,
          isEnabled: true
        }
      : { id: i + 1, studentId: null, isAnonymous: i === 0, isHandled: i === 1, attachments: [], content: '建议' }
  );
  const state = { fail: false, saves: 0, query: null };
  const failure = { data: null, error: { message: '失败' } };
  const query = async q => {
    state.query = q;
    if (state.fail) return failure;
    const filtered = rows.filter(
      row =>
        (q.isAnonymous == null || row.isAnonymous === q.isAnonymous) &&
        (q.isHandled == null || row.isHandled === q.isHandled)
    );
    return {
      data: {
        items: filtered.slice((q.page - 1) * q.pageSize, q.page * q.pageSize),
        total: filtered.length
      },
      error: null
    };
  };
  const remove = async id => {
    if (state.fail) return failure;
    rows.splice(
      rows.findIndex(row => row.id === id),
      1
    );
    return { data: null, error: null };
  };
  return {
    state,
    queryApartments: query,
    queryFeedbacks: query,
    deleteApartment: remove,
    deleteFeedback: remove,
    getFeedback: async id => (state.fail ? failure : { data: { ...rows.find(row => row.id === id) }, error: null }),
    getFeedbackAttachment: async () => new Blob(['image'], { type: 'image/png' }),
    updateFeedbackHandling: async (id, handled) => {
      state.saves++;
      if (state.fail) return failure;
      const row = rows.find(item => item.id === id);
      row.isHandled = handled;
      return { data: { ...row }, error: null };
    },
    saveApartment: async () => {
      state.saves++;
      return state.fail ? failure : { data: {}, error: null };
    }
  };
}
for (const kind of ['apartment', 'feedback']) {
  test(`${kind} page retains confirmation on failure and returns from empty last page`, async () => {
    const api = pageApi(kind);
    const instance = loadPage(kind, api);
    const page = instance.page;
    await settle();
    page.pageSize.value = 1;
    page.pagination.value.onUpdatePage(3);
    await settle();
    const row = page.rows.value[0];
    page.confirmDelete(row);
    api.state.fail = true;
    assert.equal(await instance.dialog.onPositiveClick(), false);
    assert.equal(page.total.value, 3);
    api.state.fail = false;
    assert.equal(await instance.dialog.onPositiveClick(), true);
    assert.equal(page.page.value, 2);
    assert.equal(page.total.value, 2);
    assert.equal(page.rows.value.length, 1);
    assert.equal(instance.errors.length, 0);
  });
  test(`${kind} page clears loading on request failure`, async () => {
    const api = pageApi(kind);
    api.state.fail = true;
    const { page, errors } = loadPage(kind, api);
    await settle();
    assert.equal(page.loading.value, false);
    assert.equal(page.total.value, 0);
    assert.equal(errors.length, 0);
  });
}
test('feedback page applies filters only on query, resets and closes failed details', async () => {
  const api = pageApi('feedback');
  const { page, errors } = loadPage('feedback', api);
  await settle();
  page.search.isAnonymous = true;
  assert.equal(page.total.value, 3);
  page.applySearch();
  await settle();
  assert.equal(page.total.value, 1);
  page.resetSearch();
  await settle();
  assert.equal(page.total.value, 3);
  api.state.fail = true;
  await page.openDetail(page.rows.value[0]);
  assert.equal(page.drawerVisible.value, false);
  assert.equal(errors.length, 0);
});
test('apartment page retains failed form and prevents duplicate saves', async () => {
  const api = pageApi('apartment');
  const { page, errors } = loadPage('apartment', api);
  await settle();
  await page.openForm();
  page.formRef.value = {
    validate: async () => {},
    restoreValidation: () => {}
  };
  Object.assign(page.model, { name: '测试楼', gender: 'male' });
  api.state.fail = true;
  await page.submit();
  assert.equal(page.drawerVisible.value, true);
  assert.equal(page.saving.value, false);
  assert.equal(errors.length, 0);
  api.state.fail = false;
  const before = api.state.saves;
  await Promise.all([page.submit(), page.submit()]);
  assert.equal(api.state.saves, before + 1);
  assert.equal(page.drawerVisible.value, false);
});

for (const kind of ['apartment', 'feedback']) {
  test(`${kind} page ignores older list responses`, async () => {
    const api = pageApi(kind);
    const pending = [];
    const query = () => new Promise(resolve => pending.push(resolve));
    api.queryApartments = query;
    api.queryFeedbacks = query;
    const { page } = loadPage(kind, api);
    page.pagination.value.onUpdatePage(2);
    pending[1]({ data: { items: [{ id: 2 }], total: 20 }, error: null });
    await settle();
    pending[0]({ data: { items: [{ id: 1 }], total: 10 }, error: null });
    await settle();
    assert.equal(page.rows.value[0].id, 2);
    assert.equal(page.total.value, 20);
    assert.equal(page.loading.value, false);
  });
}
test('feedback detail ignores older responses', async () => {
  const api = pageApi('feedback');
  const pending = [];
  api.getFeedback = () => new Promise(resolve => pending.push(resolve));
  const { page } = loadPage('feedback', api);
  await settle();
  const first = page.openDetail({ id: 1 });
  const second = page.openDetail({ id: 2 });
  pending[1]({ data: { id: 2, attachments: [] }, error: null });
  await second;
  pending[0]({ data: { id: 1, attachments: [] }, error: null });
  await first;
  assert.equal(page.detail.value.id, 2);
});

test('feedback handling sends explicit false and maps state and detail attachments', async () => {
  let options;
  const api = loadApi('feedback', async input => {
    options = input;
    return {
      data: { ...feedbackWire, attachments: [{ index: 0, url: '/admin/feedbacks/1/attachments/0' }] },
      error: null
    };
  });
  const result = await api.updateFeedbackHandling(1, false);
  assert.equal(options.method, 'patch');
  assert.equal(options.data.is_handled, false);
  assert.equal(result.data.isHandled, false);
  assert.equal(result.data.attachments[0].index, 0);
});

test('feedback handling retains failure state and prevents duplicate mutations', async () => {
  const api = pageApi('feedback');
  const { page } = loadPage('feedback', api);
  await settle();
  await page.openDetail(page.rows.value[0]);
  api.state.fail = true;
  await page.toggleHandling();
  assert.equal(page.detail.value.isHandled, false);
  assert.equal(page.handling.value, false);
  api.state.fail = false;
  const before = api.state.saves;
  await Promise.all([page.toggleHandling(), page.toggleHandling()]);
  assert.equal(api.state.saves, before + 1);
  assert.equal(page.detail.value.isHandled, true);
  assert.equal(page.rows.value[0].isHandled, true);
  await page.toggleHandling();
  assert.equal(page.detail.value.isHandled, false);
});

test('feedback ignores old attachments and releases preview URLs when closed', async () => {
  const api = pageApi('feedback');
  const pending = [];
  api.getFeedback = async id => ({ data: { id, attachments: [{ index: 0 }] }, error: null });
  api.getFeedbackAttachment = () => new Promise(resolve => pending.push(resolve));
  const { page, urls } = loadPage('feedback', api);
  await settle();
  await page.openDetail({ id: 1 });
  await page.openDetail({ id: 2 });
  pending[1]('new image');
  await settle();
  assert.equal(page.attachmentPreviews.value.length, 1);
  const url = page.attachmentPreviews.value[0].url;
  pending[0]('old image');
  await settle();
  assert.equal(urls.created.length, 1);
  assert.equal(page.attachmentPreviews.value[0].url, url);
  page.drawerVisible.value = false;
  await settle();
  assert.equal(page.attachmentPreviews.value.length, 0);
  assert.ok(urls.revoked.includes(url));
});

test('feedback attachment failures allow details to remain readable', async () => {
  const api = pageApi('feedback');
  api.getFeedback = async () => ({ data: { id: 1, attachments: [{ index: 0 }] }, error: null });
  api.getFeedbackAttachment = async () => {
    throw new Error('failed');
  };
  const { page } = loadPage('feedback', api);
  await settle();
  await page.openDetail({ id: 1 });
  await settle();
  assert.equal(page.detail.value.id, 1);
  assert.equal(page.attachmentPreviews.value[0].failed, true);
  assert.equal(page.attachmentsLoading.value, false);
});

test('feedback does not create URLs for partially loaded details after unmount', async () => {
  const api = pageApi('feedback');
  const pending = [];
  api.getFeedback = async () => ({ data: { id: 1, attachments: [{ index: 0 }, { index: 1 }] }, error: null });
  api.getFeedbackAttachment = () => new Promise(resolve => pending.push(resolve));
  const { page, urls, unmount } = loadPage('feedback', api);
  await settle();
  await page.openDetail({ id: 1 });
  pending[0]('first image');
  await settle();
  unmount();
  pending[1]('second image');
  await settle();
  assert.equal(urls.created.length, 0);
});


test('feedback handling filter applies on query and resets to all statuses', async () => {
  const { page } = loadPage('feedback', pageApi('feedback'));
  await settle();
  page.search.isHandled = false;
  assert.equal(page.total.value, 3);
  page.applySearch();
  await settle();
  assert.equal(page.total.value, 2);
  assert.ok(page.rows.value.every(row => !row.isHandled));
  page.search.isHandled = true;
  page.applySearch();
  await settle();
  assert.equal(page.total.value, 1);
  assert.ok(page.rows.value.every(row => row.isHandled));
  page.resetSearch();
  await settle();
  assert.equal(page.search.isHandled, null);
  assert.equal(page.total.value, 3);
});
