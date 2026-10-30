import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function loadModule(path, require = () => ({})) {
  const context = { exports: {}, require };
  vm.runInNewContext(
    ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
    }).outputText,
    context
  );
  return context.exports;
}
const { adaptOverview } = loadModule('../src/service/api/overview.ts');
const { createOverviewLoader } = loadModule('../src/views/home/data.ts');
const { formatOverviewDateRange } = loadModule('../src/views/home/date-range.ts');
const serialize = value => JSON.parse(JSON.stringify(value));

function fixture() {
  return {
    generated_at: '2026-10-11T12:00:00+08:00',
    statistics: { student_count: 72, class_count: 3, record_count: 125, unhandled_feedback_count: 7 },
    weekly_applications: {
      weeks: Array.from({ length: 9 }, (_, index) => ({
        start: new Date(Date.parse('2026-08-10T00:00:00+08:00') + index * 604800000).toISOString(),
        end: new Date(Date.parse('2026-08-16T00:00:00+08:00') + index * 604800000).toISOString(),
        count: index === 8 ? 2 : 0,
        is_current: index === 8
      })),
      total: 2
    },
    version_usage: { total: 72, latest_count: 48, groups: [{ key: 'latest', version: '0.0.10', count: 48, percentage: 66.666666 }] },
    recent_records: [{ id: 3, name: '李欣然', student_id: '202624000001', class_name: '计24-1', leave_type_id: 1, leave_type_name: '事假-本科生', applied_at: '2026-10-11T11:00:00+08:00' }],
    latest_android_version: { version_name: '0.0.10', version_code: 10, published_at: '2026-10-05T08:00:00+08:00', release_notes: [] }
  };
}

test('adapts real response fields and timestamps without recalculating business data', () => {
  const wire = fixture();
  const data = adaptOverview(wire);
  assert.equal(data.generatedAt, Date.parse(wire.generated_at));
  assert.deepEqual(serialize(data.statistics), { studentCount: 72, classCount: 3, recordCount: 125, unhandledFeedbackCount: 7 });
  assert.equal(data.weeklyApplications.weeks.length, 9);
  assert.equal(data.weeklyApplications.weeks[8].isCurrent, true);
  assert.equal(data.weeklyApplications.weeks[8].start, Date.parse('2026-10-05T00:00:00+08:00'));
  assert.equal(data.weeklyApplications.total, 2);
  assert.equal(data.recentRecords[0].studentId, '202624000001');
  assert.equal(data.recentRecords[0].appliedAt, Date.parse(wire.recent_records[0].applied_at));
  assert.equal(data.latestVersion.versionCode, 10);
  assert.equal(data.latestVersion.publishedAt, Date.parse(wire.latest_android_version.published_at));
  assert.equal(data.versionUsage.latestCount, 48);
  assert.deepEqual(serialize(data.versionUsage.groups), wire.version_usage.groups);
});

test('preserves empty database, zero weeks and unavailable release', () => {
  const wire = fixture();
  wire.statistics = { student_count: 0, class_count: 0, record_count: 0, unhandled_feedback_count: 0 };
  wire.weekly_applications.weeks.forEach(week => { week.count = 0; });
  wire.weekly_applications.total = 0;
  wire.version_usage = { total: 0, latest_count: 0, groups: [] };
  wire.recent_records = [];
  wire.latest_android_version = null;
  const data = adaptOverview(wire);
  assert.equal(data.latestVersion, null);
  assert.equal(data.weeklyApplications.weeks.length, 9);
  assert.equal(data.weeklyApplications.weeks.every(week => week.count === 0), true);
  assert.deepEqual(serialize(data.recentRecords), []);
  assert.deepEqual(serialize(data.versionUsage.groups), []);
});

test('request uses overview endpoint and passes errors without fake data', async () => {
  const calls = [];
  let response = { data: fixture(), error: null };
  const { getOverview } = loadModule('../src/service/api/overview.ts', () => ({ request: async options => { calls.push(options); return response; } }));
  assert.equal((await getOverview()).data.statistics.recordCount, 125);
  assert.equal(calls[0].url, '/admin/overview');
  response = { data: null, error: new Error('failed') };
  assert.equal(await getOverview(), response);
});

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}
const initialState = () => ({ data: null, loading: false, failed: false });

test('range captions use readable month and day; tooltips include the starting year', () => {
  const start = Date.parse('2026-08-10T00:00:00+08:00');
  const end = Date.parse('2026-10-11T00:00:00+08:00');
  assert.equal(formatOverviewDateRange(start, end, 'zh-CN'), '8月10日 — 10月11日');
  assert.equal(formatOverviewDateRange(start, end, 'zh-CN', true), '2026年8月10日 — 10月11日');
  assert.equal(formatOverviewDateRange(start, end, 'en-US'), 'Aug 10 — Oct 11');
  assert.equal(formatOverviewDateRange(start, end, 'en-US', true), 'Aug 10, 2026 — Oct 11');
});

test('cross-year ranges show both years using Beijing dates regardless of UTC date', () => {
  const start = Date.parse('2026-12-28T00:00:00+08:00');
  const end = Date.parse('2026-12-31T16:00:00Z');
  assert.equal(formatOverviewDateRange(start, end, 'zh-CN'), '2026年12月28日 — 2027年1月1日');
  assert.equal(formatOverviewDateRange(start, end, 'zh-CN', true), '2026年12月28日 — 2027年1月1日');
  assert.equal(formatOverviewDateRange(start, end, 'en-US'), 'Dec 28, 2026 — Jan 1, 2027');
});

test('deduplicates mount and initial activation and replaces data on refresh', async () => {
  const state = initialState();
  const pending = deferred();
  let calls = 0;
  const loader = createOverviewLoader(state, () => { calls += 1; return pending.promise; });
  const first = loader.reload();
  await loader.reload();
  assert.equal(calls, 1);
  assert.equal(state.loading, true);
  const data = adaptOverview(fixture());
  pending.resolve(data);
  await first;
  assert.equal(state.loading, false);
  assert.equal(state.failed, false);
  assert.equal(state.data, data);
  await loader.reload();
  assert.equal(calls, 2);
});

test('initial failure stays null and retry can succeed', async () => {
  const state = initialState();
  let response = null;
  const loader = createOverviewLoader(state, async () => response);
  await loader.reload();
  assert.equal(state.failed, true);
  assert.equal(state.data, null);
  assert.equal(state.loading, false);
  response = adaptOverview(fixture());
  await loader.reload();
  assert.equal(state.failed, false);
  assert.equal(state.data, response);
});

test('refresh failure keeps the last successful snapshot and statistics time', async () => {
  const snapshot = adaptOverview(fixture());
  const state = { ...initialState(), data: snapshot };
  const loader = createOverviewLoader(state, async () => { throw new Error('network'); });
  await loader.reload();
  assert.equal(state.failed, true);
  assert.equal(state.data, snapshot);
  assert.equal(state.data.generatedAt, snapshot.generatedAt);
  assert.equal(state.loading, false);
});

test('leaving and returning ignores old successes, errors and loading updates', async () => {
  for (const oldFails of [false, true]) {
    const state = initialState();
    const old = deferred();
    const fresh = deferred();
    let calls = 0;
    const loader = createOverviewLoader(state, () => (++calls === 1 ? old.promise : fresh.promise));
    const first = loader.reload();
    loader.invalidate();
    const second = loader.reload();
    if (oldFails) old.reject(new Error('stale failure'));
    else old.resolve(adaptOverview(fixture()));
    await first;
    assert.equal(state.data, null);
    assert.equal(state.failed, false);
    assert.equal(state.loading, true);
    const snapshot = adaptOverview(fixture());
    fresh.resolve(snapshot);
    await second;
    assert.equal(state.data, snapshot);
    assert.equal(state.loading, false);
  }
});

test('late response after new response cannot overwrite the new snapshot', async () => {
  const state = initialState();
  const old = deferred();
  const fresh = deferred();
  let calls = 0;
  const loader = createOverviewLoader(state, () => (++calls === 1 ? old.promise : fresh.promise));
  const first = loader.reload();
  loader.invalidate();
  const second = loader.reload();
  const snapshot = adaptOverview(fixture());
  fresh.resolve(snapshot);
  await second;
  old.resolve({ ...snapshot, generatedAt: 0 });
  await first;
  assert.equal(state.data, snapshot);
});
