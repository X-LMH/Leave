import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import vm from 'node:vm';
import ts from 'typescript';

function nativeSource(source) {
  let web = false;
  return source
    .split('\n')
    .filter(line => {
      if (line.includes('// #ifdef WEB')) {
        web = true;
        return false;
      }
      if (line.includes('// #endif') && web) {
        web = false;
        return false;
      }
      return !web;
    })
    .join('\n');
}

function setup(mode = 'auto', system = 'light', stored = null) {
  const storage = new Map([['appearance_mode', mode]]);
  let appTheme = mode;
  let osTheme = system;
  let themeChange;
  const downloads = [];
  const installations = [];
  const hooks = {};
  let versionRequest = async () => {
    throw new Error('版本服务暂不可用');
  };
  const uni = {
    getStorageSync: key => storage.get(key) ?? '',
    setStorageSync: (key, value) => storage.set(key, value),
    getDeviceInfo: () => ({ osTheme }),
    getAppBaseInfo: () => ({ appTheme }),
    onAppThemeChange: callback => {
      themeChange = callback;
    },
    setAppTheme: options => {
      appTheme = options.theme;
      themeChange?.({ appTheme: appTheme === 'auto' ? osTheme : appTheme });
      options.success?.({});
    },
    showToast: () => {},
    getElementById: () => null,
    downloadFile: options => {
      const request = { options, progress: null };
      downloads.push(request);
      return {
        onProgressUpdate: callback => {
          request.progress = callback;
        }
      };
    },
    installApk: options => installations.push(options)
  };
  const vue = {
    reactive: value => value,
    ref: value => ({ value }),
    computed: getter => ({
      get value() {
        return getter();
      }
    }),
    nextTick: callback => {
      callback?.();
    },
    watch: () => {}
  };
  function evaluate(source, dependencies, globals = {}) {
    const context = {
      exports: {},
      uni,
      ...globals,
      require: name => {
        if (name === 'vue') return vue;
        if (name in dependencies) return dependencies[name];
        throw new Error(`Unexpected module: ${name}`);
      }
    };
    vm.runInNewContext(
      ts.transpileModule(source, {
        compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 }
      }).outputText,
      context
    );
    return context.exports;
  }
  const theme = evaluate(
    nativeSource(readFileSync(new URL('../../frontend/utils/app-theme.uts', import.meta.url), 'utf8')),
    {
      '@/utils/navigation-preferences.uts': { systemIncreasesContrast: () => false }
    }
  );
  theme.initializeAppTheme();
  const pageSource = readFileSync(new URL('../../frontend/pages/update/index.uvue', import.meta.url), 'utf8')
    .split('<script setup lang="uts">')[1]
    .split('</script>')[0];
  const page = evaluate(
    pageSource +
      '\nexport { updateGroups, loadVersion, downloadAndInstall, redownloadApk, latestVersion, loadingVersion, loadError, downloading, installing, downloadedApk, downloadProgressDisplay, actionText, actionError, themeClass, themeImage };',
    {
      '@/utils/app-theme.uts': theme,
      '@/utils/navigation-preferences.uts': { configureNavigationRadio() {} },
      '@/utils/device.uts': { getAppVersion: () => '1.1.0' },
      '@/utils/api.uts': {
        getStoredUpdateInfo: () => stored,
        getCurrentAppVersion: () => versionRequest(),
        saveUpdateInfo: info => storage.set('leave_update_info', JSON.stringify(info)),
        getAppDownloadURL: () => 'http://127.0.0.1:8080/api/v1/app-download?platform=android',
        releaseNoteSections: () => []
      }
    },
    Object.fromEntries(
      ['onReady', 'onResize', 'onShow', 'onLoad', 'onBackPress'].map(name => [
        name,
        callback => {
          hooks[name] = callback;
        }
      ])
    )
  );
  return {
    theme,
    page,
    storage,
    downloads,
    installations,
    hooks,
    request: callback => {
      versionRequest = callback;
    },
    systemTheme: value => {
      osTheme = value;
      if (appTheme === 'auto') themeChange?.({ appTheme: value });
    }
  };
}

test('update page follows global appearance without its own selector', () => {
  const { page, theme, systemTheme } = setup('auto', 'dark');
  assert.match(page.themeClass('update-page'), /dark-update-page/);
  theme.setAppearanceMode('light');
  assert.equal(page.themeClass('update-page'), 'update-page');
  theme.setAppearanceMode('auto');
  systemTheme('dark');
  assert.match(page.themeClass('update-page'), /dark-update-page/);
  const source = readFileSync(new URL('../../frontend/pages/update/index.uvue', import.meta.url), 'utf8');
  assert.doesNotMatch(source, /appearance-options|chooseAppearance|update-mode-/);
});

test('updates sort numerically newest first without changing historical cache order', () => {
  const versions = [2, 11, 3].map(version_code => ({
    version_code,
    version_name: String(version_code),
    release_notes: []
  }));
  const { page } = setup('auto', 'light', { updates: versions });
  assert.deepEqual(
    Array.from(page.updateGroups.value, version => version.version_code),
    [11, 3, 2]
  );
  assert.deepEqual(
    versions.map(version => version.version_code),
    [2, 11, 3]
  );
});

test('update page restores explicit appearance on cold start and refreshes on return', () => {
  const { page, hooks } = setup('dark', 'light');
  hooks.onShow();
  assert.match(page.themeClass('update-page'), /dark-update-page/);
  assert.equal(hooks.onBackPress(), true);
});

test('failed version fetch can retry and cache the successful response', async () => {
  const { page, request, storage } = setup();
  await page.loadVersion();
  assert.equal(page.loadingVersion.value, false);
  assert.equal(page.loadError.value, '版本服务暂不可用');
  const version = { platform: 'android', latest_version_code: 17, latest_version_name: '1.2.0', updates: [] };
  request(async () => version);
  await page.loadVersion();
  assert.equal(page.loadError.value, '');
  assert.equal(page.latestVersion.value, version);
  assert.equal(JSON.parse(storage.get('leave_update_info')).latest_version_code, 17);
});

test('download handles duplicate clicks, progress, HTTP failure and retry', () => {
  const { page, downloads } = setup();
  page.downloadAndInstall();
  page.downloadAndInstall();
  assert.equal(downloads.length, 1);
  downloads[0].progress({ progress: 37.9 });
  assert.equal(page.downloadProgressDisplay.value, 37);
  assert.equal(page.actionText.value, '正在下载');
  downloads[0].options.success({ statusCode: 503 });
  assert.equal(page.downloading.value, false);
  assert.equal(page.actionText.value, '重新下载');
  page.downloadAndInstall();
  assert.equal(downloads.length, 2);
  downloads[1].options.fail({ errMsg: '网络不可用' });
  assert.equal(page.actionError.value, '网络不可用');
});

test('install failure retries existing APK and explicit redownload replaces it', () => {
  const { page, downloads, installations } = setup();
  page.downloadAndInstall();
  downloads[0].options.success({ statusCode: 200, tempFilePath: '/cache/update.apk' });
  assert.equal(page.installing.value, true);
  page.downloadAndInstall();
  assert.equal(installations.length, 1);
  installations[0].fail({ errMsg: '安装器未打开' });
  installations[0].complete();
  assert.equal(page.actionText.value, '重试安装');
  page.downloadAndInstall();
  assert.equal(downloads.length, 1);
  assert.equal(installations.length, 2);
  assert.equal(installations[1].filePath, '/cache/update.apk');
  installations[1].complete();
  page.redownloadApk();
  assert.equal(page.downloadedApk.value, '');
  assert.equal(downloads.length, 2);
});
