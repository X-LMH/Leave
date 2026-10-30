import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';
import { test } from 'node:test';
import ts from 'typescript';
import { defu } from 'defu';

const source = readFileSync(new URL('../src/plugins/loading.ts', import.meta.url), 'utf8');
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
}).outputText;

const sharedSource = readFileSync(new URL('../src/store/modules/theme/shared.ts', import.meta.url), 'utf8');
const sharedCompiled = ts.transpileModule(sharedSource.replace('import.meta.env.PROD', 'true'), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
}).outputText;

function loadSettings(cachedSettings, overrideThemeSettings = {}) {
  const storage = new Map([['themeSettings', cachedSettings]]);
  const imports = {
    defu: { defu },
    '@sa/color': {},
    '@/constants/app': {},
    '@/utils/common': {},
    '@/utils/storage': { localStg: { get: key => storage.get(key), set: (key, value) => storage.set(key, value) } },
    '@/theme/settings': { themeSettings: { themeScheme: 'auto' }, overrideThemeSettings },
    '@/theme/vars': {}
  };
  const context = { exports: {}, BUILD_TIME: 'test-build', require: name => imports[name] };
  vm.runInNewContext(sharedCompiled, context);
  return { initThemeSettings: context.exports.initThemeSettings, storage };
}

test('production first visit follows system by default', () => {
  const { initThemeSettings } = loadSettings(undefined);
  assert.equal(initThemeSettings().themeScheme, 'auto');
});

test('production preserves existing manual theme preferences', () => {
  for (const themeScheme of ['light', 'dark']) {
    const { initThemeSettings } = loadSettings({ themeScheme });
    assert.equal(initThemeSettings(false).themeScheme, themeScheme);
    assert.equal(initThemeSettings().themeScheme, themeScheme);
  }
});

test('loading preview leaves release overrides available to the store', () => {
  const { initThemeSettings, storage } = loadSettings({ themeScheme: 'light' }, { themeScheme: 'auto' });
  assert.equal(initThemeSettings(false).themeScheme, 'auto');
  assert.equal(storage.has('overrideThemeFlag'), false);
  assert.equal(initThemeSettings().themeScheme, 'auto');
  assert.equal(storage.get('overrideThemeFlag'), 'test-build');
});

for (const [scheme, systemDark, expected] of [
  ['auto', true, true],
  ['auto', false, false],
  ['light', true, false],
  ['dark', false, true]
]) {
  test(`loading: ${scheme}, system dark=${systemDark}`, () => {
    let applied;
    let markOverrideApplied;
    const app = { innerHTML: '' };
    const imports = {
      '@sa/color': {
        getColorPalette: () => new Map(),
        getRgb: () => ({ r: 100, g: 108, b: 255 })
      },
      '@/utils/storage': {
        localStg: {
          get: key => {
            assert.equal(key, 'themeColor', 'must not read stale darkMode cache');
            return null;
          }
        }
      },
      '@/locales': { $t: () => 'Leave' },
      '@/store/modules/theme/shared': {
        initThemeSettings: mark => {
          markOverrideApplied = mark;
          return { themeScheme: scheme };
        },
        toggleCssDarkMode: value => {
          applied = value;
        }
      }
    };
    const context = {
      exports: {},
      require: name => imports[name],
      window: {
        matchMedia: query => {
          assert.equal(query, '(prefers-color-scheme: dark)');
          return { matches: systemDark };
        }
      },
      document: { getElementById: () => app }
    };
    vm.runInNewContext(compiled, context);
    context.exports.setupLoading();
    assert.equal(applied, expected);
    assert.equal(markOverrideApplied, false, 'loading must not consume release overrides');
    assert.match(app.innerHTML, /Leave/);
  });
}
