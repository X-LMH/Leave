import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';
import * as vue from 'vue';
import { compileScript, parse } from '@vue/compiler-sfc';

// Supply a frame clock before VueUse captures its browser environment.
let frames = [];
globalThis.window = { requestAnimationFrame: callback => frames.push(callback) };
globalThis.document = {};
const vueuse = await import('@vueuse/core');
delete globalThis.window;
delete globalThis.document;
const context = {
  exports: {},
  require: name => (name === 'vue' ? vue : vueuse)
};
vm.runInNewContext(
  ts.transpileModule(readFileSync(new URL('../src/views/home/motion.ts', import.meta.url), 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
  }).outputText,
  context
);
const { useHomeTransition, HOME_CHART_ANIMATION_DURATION } = context.exports;

function loadComponent(filename, overrides = {}) {
  const { descriptor } = parse(readFileSync(new URL(filename, import.meta.url), 'utf8'));
  const compiled = compileScript(descriptor, { id: 'home-chart-test' });
  const imports = {
    vue,
    '../motion': context.exports,
    '@/store/modules/app': { useAppStore: () => ({ locale: 'zh-CN', isMobile: false }) },
    '@/store/modules/theme': { useThemeStore: () => ({ themeColor: '#1677ff' }) },
    '@/locales': { $t: key => key },
    'naive-ui': { useThemeVars: () => vue.ref({}) },
    '../date-range': { formatOverviewDateRange: () => '' },
    './animated-count.vue': { default: {} },
    ...overrides
  };
  const module = { exports: {}, require: name => imports[name], Date };
  vm.runInNewContext(
    ts.transpileModule(compiled.content, {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
    }).outputText,
    module
  );
  return { ...module.exports.default, render: () => vue.h('span') };
}

function mountComponent(component, initial, events = {}) {
  const props = vue.reactive(initial);
  let instance;
  const observedComponent = {
    ...component,
    setup(componentProps, setupContext) {
      instance = vue.getCurrentInstance();
      return component.setup(componentProps, setupContext);
    }
  };
  const renderer = vue.createRenderer({
    createElement: () => ({}),
    createText: () => ({}),
    createComment: () => ({}),
    insert() {},
    remove() {},
    setText() {},
    setElementText() {},
    patchProp() {},
    parentNode: () => null,
    nextSibling: () => null
  });
  const app = renderer.createApp({ render: () => vue.h(observedComponent, { ...props, ...events }) });
  app.mount({});
  return { app, props, instance };
}

function animation(t, initial) {
  let now = 0;
  frames = [];
  t.mock.method(Date, 'now', () => now);
  const values = vue.ref(initial);
  const enabled = vue.ref(true);
  const scope = vue.effectScope();
  const displayed = scope.run(() => useHomeTransition(values, enabled));
  t.after(() => scope.stop());
  return {
    values,
    enabled,
    displayed,
    async advance(milliseconds) {
      now += milliseconds;
      const pending = frames;
      frames = [];
      pending.forEach(callback => callback());
      await vue.nextTick();
    }
  };
}

test('interrupted updates continue from the displayed number and settle exactly, including zero', async t => {
  const motion = animation(t, [0]);
  motion.values.value = [1200];
  await vue.nextTick();
  await motion.advance(150);
  const intermediate = motion.displayed.value[0];
  assert.ok(intermediate > 0 && intermediate < 1200);
  motion.values.value = [2400];
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], intermediate);
  await motion.advance(HOME_CHART_ANIMATION_DURATION);
  assert.equal(motion.displayed.value[0], 2400);
  motion.values.value = [0];
  await vue.nextTick();
  await motion.advance(HOME_CHART_ANIMATION_DURATION);
  assert.equal(motion.displayed.value[0], 0);
});

test('disabled updates synchronize the baseline before re-enabling motion', async t => {
  const motion = animation(t, [0]);
  motion.enabled.value = false;
  await vue.nextTick();
  motion.values.value = [3500];
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], 3500);
  motion.enabled.value = true;
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], 3500);
  await motion.advance(HOME_CHART_ANIMATION_DURATION);
  motion.values.value = [4000];
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], 3500);
  await motion.advance(100);
  motion.enabled.value = false;
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], 4000);
  await motion.advance(HOME_CHART_ANIMATION_DURATION);
  assert.equal(motion.displayed.value[0], 4000);
  motion.values.value = [4700];
  await vue.nextTick();
  motion.enabled.value = true;
  motion.values.value = [5000];
  await vue.nextTick();
  assert.equal(motion.displayed.value[0], 4700);
  await motion.advance(HOME_CHART_ANIMATION_DURATION);
  assert.equal(motion.displayed.value[0], 5000);
});

test('segment additions and removals keep the same slots and preserve the total width', async t => {
  const motion = animation(t, [70, 0, 30, 0, 0]);
  motion.values.value = [50, 0, 0, 0, 50];
  await vue.nextTick();
  await motion.advance(HOME_CHART_ANIMATION_DURATION / 2);
  const widths = Array.from(motion.displayed.value);
  assert.ok(widths[2] > 0 && widths[2] < 30);
  assert.ok(widths[4] > 0 && widths[4] < 50);
  assert.ok(Math.abs(widths.reduce((sum, value) => sum + value, 0) - 100) < 0.0001);
  await motion.advance(HOME_CHART_ANIMATION_DURATION / 2);
  assert.deepEqual(Array.from(motion.displayed.value), [50, 0, 0, 0, 50]);
  motion.values.value = [0, 0, 0, 0, 0];
  motion.enabled.value = false;
  await vue.nextTick();
  assert.deepEqual(Array.from(motion.displayed.value), [0, 0, 0, 0, 0]);
});

test('version reveal starts with the trend draw, ends on its completion, and stops when disabled', async t => {
  let now = 0;
  t.mock.method(Date, 'now', () => now);
  const version = mountComponent(loadComponent('../src/views/home/modules/version-usage.vue'), {
    usage: { total: 10, latestCount: 10, groups: [{ key: 'latest', version: '1', count: 10, percentage: 100 }] },
    latestVersion: '1',
    animate: true
  });
  t.after(() => version.app.unmount());
  const reveals = [];
  version.instance.setupState.barRef = {
    animate(keyframes, options) {
      const reveal = {
        keyframes,
        options,
        cancelled: false,
        cancel() {
          this.cancelled = true;
        }
      };
      reveals.push(reveal);
      return reveal;
    }
  };
  let hooks;
  let options;
  const trend = mountComponent(
    loadComponent('../src/views/home/modules/line-chart.vue', {
      '@/hooks/common/echarts': {
        useEcharts(factory, chartHooks) {
          hooks = chartHooks;
          options = factory();
          return { domRef: vue.ref(null), updateOptions: async () => {} };
        }
      }
    }),
    { weeks: [], animate: true },
    {
      onDraw: version.instance.exposed.startReveal,
      onFinished: version.instance.exposed.finishReveal
    }
  );
  t.after(() => trend.app.unmount());
  assert.equal(version.instance.setupState.revealed, false);
  const events = {};
  const chart = {
    on: (name, handler) => {
      events[name] = handler;
    },
    hideLoading() {}
  };
  hooks.onRender(chart);
  assert.equal(reveals.length, 0);
  hooks.onUpdated(chart);
  assert.equal(reveals.length, 1);
  assert.equal(options.animationDuration, 1000);
  assert.equal(reveals[0].options.duration, options.animationDuration);
  assert.equal(reveals[0].options.easing, 'linear');
  now = 300;
  assert.equal(hooks.resizeAnimation().duration, 700);
  now = 850;
  assert.equal(hooks.resizeAnimation().duration, 150);
  events.finished();
  assert.equal(hooks.resizeAnimation().duration, 0);
  assert.equal(reveals[0].cancelled, true);
  assert.equal(version.instance.setupState.revealed, true);
  hooks.onUpdated(chart);
  assert.equal(reveals.length, 2);
  version.props.animate = false;
  trend.props.animate = false;
  await vue.nextTick();
  assert.equal(reveals[1].cancelled, true);
  assert.equal(hooks.resizeAnimation().duration, 0);
  hooks.onUpdated(chart);
  assert.equal(reveals.length, 2);
});

test('chart size changes preserve the remaining animation instead of drawing a static frame', async t => {
  const dimensions = { width: vue.ref(0), height: vue.ref(0) };
  const resizeCalls = [];
  let draws = 0;
  const chart = {
    setOption() {
      draws += 1;
    },
    resize: options => resizeCalls.push(options),
    dispose() {}
  };
  const imports = {
    vue,
    '@vueuse/core': { useElementSize: () => dimensions },
    'echarts/core': { use() {}, init: () => chart },
    '@/store/modules/theme': { useThemeStore: () => ({ darkMode: false }) }
  };
  const module = { exports: {}, require: name => imports[name] ?? {} };
  vm.runInNewContext(
    ts.transpileModule(readFileSync(new URL('../src/hooks/common/echarts.ts', import.meta.url), 'utf8'), {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
    }).outputText,
    module
  );
  let remaining = 1000;
  const scope = vue.effectScope();
  t.after(() => scope.stop());
  const hooks = scope.run(() =>
    module.exports.useEcharts(() => ({}), {
      onRender() {},
      onUpdated() {},
      resizeAnimation: () => ({ duration: remaining, easing: 'linear' })
    })
  );
  hooks.domRef.value = {};
  dimensions.width.value = 600;
  dimensions.height.value = 280;
  await vue.nextTick();
  await vue.nextTick();
  assert.equal(draws, 1);
  remaining = 700;
  dimensions.width.value = 620;
  await vue.nextTick();
  assert.equal(resizeCalls[0].animation.duration, 700);
  remaining = 0;
  dimensions.width.value = 640;
  await vue.nextTick();
  assert.equal(resizeCalls[1].animation.duration, 0);
  assert.equal(draws, 1);
});
