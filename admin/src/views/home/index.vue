<script setup lang="ts">
import { computed, h, onActivated, onDeactivated, onMounted, onUnmounted, reactive } from 'vue';
import { useNow } from '@vueuse/core';
import { NTag, useThemeVars } from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import type { RouteKey } from '@elegant-router/types';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { useRouterPush } from '@/hooks/common/router';
import { $t } from '@/locales';
import { getOverview } from '@/service/api/overview';
import type { OverviewRecord } from '@/service/api/overview';
import { createOverviewLoader } from './data';
import { formatOverviewDateRange } from './date-range';
import type { OverviewLoadState } from './data';
import LineChart from './modules/line-chart.vue';
import VersionUsage from './modules/version-usage.vue';

const appStore = useAppStore();
const themeStore = useThemeStore();
const themeVars = useThemeVars();
const { routerPushByKey } = useRouterPush();
const clock = useNow({ interval: 60_000 });
const state = reactive<OverviewLoadState>({ data: null, loading: false, failed: false });
const overview = computed(() => state.data);
const weekly = computed(() => overview.value?.weeklyApplications);
const { reload, invalidate } = createOverviewLoader(state, async () => {
  const result = await getOverview();
  return result.error ? null : result.data;
});
onMounted(() => void reload());
onActivated(() => void reload());
onDeactivated(invalidate);
onUnmounted(invalidate);
const themeStyle = computed(() => ({
  '--management-accent': themeVars.value.primaryColor,
  '--management-text': themeVars.value.textColor1,
  '--management-muted': themeVars.value.textColor3,
  '--management-border': themeVars.value.dividerColor,
  '--management-surface': themeVars.value.cardColor,
  '--management-inset': themeVars.value.tableHeaderColor,
  '--home-blue': themeStore.darkMode ? '#79b8ff' : '#3b82f6',
  '--home-teal': themeStore.darkMode ? '#58d3c4' : '#14b8a6',
  '--home-orange': themeStore.darkMode ? '#f5b56b' : '#ed8b2c'
}));
const dateFormatter = computed(
  () =>
    new Intl.DateTimeFormat(appStore.locale, {
      timeZone: 'Asia/Shanghai',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit'
    })
);
const weekRange = computed(
  () =>
    weekly.value
      ? formatOverviewDateRange(
          weekly.value.weeks[0].start,
          weekly.value.weeks[weekly.value.weeks.length - 1].end,
          appStore.locale
        )
      : ''
);
const generatedAt = computed(() =>
  overview.value
    ? new Intl.DateTimeFormat(appStore.locale, {
        timeZone: 'Asia/Shanghai',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
      }).format(overview.value.generatedAt)
    : ''
);
const statistics = computed(() => [
  {
    key: 'students',
    title: $t('page.home.students'),
    value: overview.value?.statistics.studentCount ?? 0,
    unit: $t('page.home.peopleUnit'),
    icon: 'lucide:users',
    tone: 'blue',
    route: 'student'
  },
  {
    key: 'classes',
    title: $t('page.home.classes'),
    value: overview.value?.statistics.classCount ?? 0,
    unit: $t('page.home.classesUnit'),
    icon: 'lucide:school',
    tone: 'teal',
    route: 'class'
  },
  {
    key: 'records',
    title: $t('page.home.records'),
    value: overview.value?.statistics.recordCount ?? 0,
    unit: $t('page.home.recordsUnit'),
    icon: 'lucide:clipboard-list',
    tone: 'orange',
    route: 'leave-record'
  },
  {
    key: 'feedback',
    title: $t('page.home.feedback'),
    value: overview.value?.statistics.feedbackCount ?? 0,
    unit: $t('page.home.recordsUnit'),
    icon: 'lucide:message-square',
    tone: 'primary',
    route: 'feedback'
  }
]);
const shortcuts = computed(() => [
  {
    route: 'student',
    title: $t('route.student'),
    icon: 'lucide:users',
    tone: 'blue',
    description: $t('page.home.studentShortcut')
  },
  {
    route: 'leave-record',
    title: $t('route.leave-record'),
    icon: 'lucide:clipboard-list',
    tone: 'orange',
    description: $t('page.home.recordShortcut')
  },
  {
    route: 'feedback',
    title: $t('route.feedback'),
    icon: 'lucide:message-square',
    tone: 'primary',
    description: $t('page.home.feedbackShortcut')
  },
  {
    route: 'app-version',
    title: $t('route.app-version'),
    icon: 'lucide:package',
    tone: 'primary',
    description: $t('page.home.versionShortcut')
  }
]);
const columns = computed<DataTableColumns<OverviewRecord>>(() => [
  {
    key: 'name',
    title: $t('page.home.student'),
    width: 160,
    render: row =>
      h('div', { class: 'home-student' }, [
        h('span', { class: 'table-person__name' }, row.name),
        h('span', { class: 'home-student__id' }, row.studentId)
      ])
  },
  { key: 'className', title: $t('page.home.className'), width: 100 },
  {
    key: 'leaveTypeName',
    title: $t('page.home.leaveType'),
    width: 140,
    render: row =>
      h(
        NTag,
        {
          size: 'small',
          bordered: false,
          type: row.leaveTypeName === '病假-本科生' ? 'success' : 'info'
        },
        () => row.leaveTypeName
      )
  },
  {
    key: 'appliedAt',
    title: $t('page.home.appliedAt'),
    width: 150,
    render: row =>
      h('div', { class: 'table-time' }, [
        h('span', dateFormatter.value.format(row.appliedAt)),
        h(
          'span',
          { class: 'table-time__clock' },
          new Intl.DateTimeFormat(appStore.locale, {
            timeZone: 'Asia/Shanghai',
            hour: '2-digit',
            minute: '2-digit',
            hour12: false
          }).format(row.appliedAt)
        )
      ])
  }
]);

function navigate(route: string) {
  void routerPushByKey(route as RouteKey);
}
</script>

<template>
  <div class="management-page home-page" :style="themeStyle">
    <header class="management-header">
      <div>
        <h1 class="management-header__title">{{ $t('page.home.title') }}</h1>
        <p class="management-header__description">{{ $t('page.home.description') }}</p>
      </div>
      <div class="home-page__date">
        <span class="home-page__calendar"><SvgIcon icon="lucide:calendar-days" /></span>
        <time>{{ dateFormatter.format(clock) }}</time>
        <NButton size="small" :loading="state.loading" @click="reload">{{ $t('page.home.refresh') }}</NButton>
      </div>
    </header>

    <div v-if="!overview && state.loading" class="home-loading" role="status" :aria-label="$t('page.home.loading')">
      <div class="home-statistics">
        <NSkeleton v-for="item in 4" :key="item" height="116px" :sharp="false" />
      </div>
      <div class="home-grid">
        <NSkeleton v-for="item in 2" :key="item" height="360px" :sharp="false" />
      </div>
    </div>
    <NAlert v-if="state.failed" type="error" :title="$t(overview ? 'page.home.refreshFailed' : 'page.home.loadFailed')" role="alert">
      <p v-if="overview">{{ $t('page.home.updatedAt', { time: generatedAt }) }}</p>
      <NButton size="small" :loading="state.loading" @click="reload">{{ $t('page.home.retry') }}</NButton>
    </NAlert>
    <template v-if="overview && weekly">
      <div class="home-statistics">
        <button
          v-for="item in statistics"
          :key="item.key"
          type="button"
          class="home-statistic"
          :class="`home-tone--${item.tone}`"
          @click="navigate(item.route)"
        >
          <span class="home-statistic__icon"><SvgIcon :icon="item.icon" /></span>
          <span class="home-statistic__body">
            <span class="home-statistic__label">{{ item.title }}</span>
            <span class="home-statistic__number">
              {{ item.value.toLocaleString(appStore.locale) }}
              <small>{{ item.unit }}</small>
            </span>
          </span>
        </button>
      </div>

      <div class="home-grid">
        <NCard :bordered="false" class="management-card home-panel">
          <template #header>
            <div class="home-panel__heading">
              <h2>{{ $t('page.home.trend') }}</h2>
              <span class="home-panel__caption">{{ weekRange }}</span>
            </div>
          </template>
          <template #header-extra>
            <span class="home-panel__caption">{{ $t('page.home.recentWeeks') }}</span>
          </template>
          <div class="home-week-toolbar">
            <div class="home-week-toolbar__current">
              <span>{{ $t('page.home.currentApplications') }}</span>
              <strong>
                {{ weekly.weeks[8].count.toLocaleString(appStore.locale) }}
                <small>{{ $t('page.home.recordsUnit') }}</small>
              </strong>
            </div>
            <span class="home-week-toolbar__total">
              {{ $t('page.home.weekTotal', { count: weekly.total.toLocaleString(appStore.locale) }) }}
            </span>
          </div>
          <LineChart :weeks="weekly.weeks" />
          <p class="home-panel__note">{{ $t('page.home.trendNote') }}</p>
        </NCard>
        <NCard :bordered="false" class="management-card home-panel">
          <template #header>
            <div class="home-panel__heading">
              <h2>{{ $t('page.home.versionUsage') }}</h2>
              <span class="home-panel__caption">{{ $t('page.home.versionUsageCaption') }}</span>
            </div>
          </template>
          <VersionUsage
            :usage="overview.versionUsage"
            :latest-version="overview.latestVersion?.versionName ?? null"
            @manage="navigate('app-version')"
          />
        </NCard>
      </div>

      <div class="home-grid home-grid--bottom">
        <NCard :bordered="false" class="management-card management-card--table home-panel">
          <template #header>
            <h2 class="home-panel__title">{{ $t('page.home.recentRecords') }}</h2>
          </template>
          <template #header-extra>
            <NButton text type="primary" @click="navigate('leave-record')">
              {{ $t('page.home.viewAll') }}
              <span class="home-link-arrow"><SvgIcon icon="lucide:arrow-right" /></span>
            </NButton>
          </template>
          <NDataTable
            :columns="columns"
            :data="overview.recentRecords"
            :row-key="row => row.id"
            :bordered="false"
            :scroll-x="600"
          />
        </NCard>
        <div class="home-side">
          <NCard :bordered="false" class="management-card home-panel">
            <template #header>
              <h2 class="home-panel__title">{{ $t('page.home.shortcuts') }}</h2>
            </template>
            <div class="home-shortcuts">
              <button
                v-for="item in shortcuts"
                :key="item.route"
                class="home-shortcut"
                :class="`home-tone--${item.tone}`"
                type="button"
                @click="navigate(item.route)"
              >
                <span class="home-shortcut__icon"><SvgIcon :icon="item.icon" /></span>
                <span class="home-shortcut__body">
                  <strong>{{ item.title }}</strong>
                  <small>{{ item.description }}</small>
                </span>
                <span class="home-shortcut__arrow"><SvgIcon icon="lucide:chevron-right" /></span>
              </button>
            </div>
          </NCard>
          <NCard :bordered="false" class="management-card home-panel">
            <template #header>
              <h2 class="home-panel__title">{{ $t('page.home.latestVersion') }}</h2>
            </template>
            <div v-if="overview.latestVersion" class="home-version">
              <span class="home-version__icon"><SvgIcon icon="ant-design:android-filled" /></span>
              <div>
                <strong>
                  Android
                  <span class="home-version__number">v{{ overview.latestVersion.versionName }}</span>
                </strong>
                <span>
                  {{ $t('page.home.publishedAt') }} {{ dateFormatter.format(overview.latestVersion.publishedAt) }}
                </span>
              </div>
              <NButton text type="primary" @click="navigate('app-version')">
                {{ $t('page.home.viewVersion') }}
                <span class="home-link-arrow"><SvgIcon icon="lucide:arrow-right" /></span>
              </NButton>
            </div>
            <p v-if="overview.latestVersion?.releaseNotes.length" class="home-version__notes">
              {{ overview.latestVersion.releaseNotes[0] }}
            </p>
            <NEmpty v-if="!overview.latestVersion" :description="$t('page.home.noVersion')" size="small">
              <template #extra>
                <NButton text type="primary" @click="navigate('app-version')">{{ $t('page.home.viewVersion') }}</NButton>
              </template>
            </NEmpty>
          </NCard>
        </div>
      </div>
      <p class="home-page__sample">{{ $t('page.home.updatedAt', { time: generatedAt }) }}</p>
    </template>
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';

.home-page {
  --home-icon-color: var(--management-accent);
  gap: 22px;

  .home-loading {
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .management-header {
    padding: 8px 0 0;
  }

  .management-header__title {
    font-size: 28px;
    letter-spacing: 0;
  }

  &__calendar {
    display: inline-flex;
    font-size: 16px;
  }

  &__date {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--management-muted);
    font-size: 13px;
  }
  &__sample {
    margin: 0;
    color: var(--management-muted);
    font-size: 12px;
  }
}
.home-tone--blue {
  --home-icon-color: var(--home-blue);
}
.home-tone--teal {
  --home-icon-color: var(--home-teal);
}
.home-tone--orange {
  --home-icon-color: var(--home-orange);
}
.home-link-arrow {
  display: inline-flex;
  margin-left: 7px;
  font-size: 16px;
}
.home-statistics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 20px;
}
.home-statistic {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 116px;
  padding: 24px;
  border: 1px solid var(--management-border);
  border-radius: 14px;
  background: var(--management-surface);
  color: var(--management-text);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 160ms,
    box-shadow 160ms;

  &:hover {
    border-color: var(--management-accent);
    box-shadow: 0 4px 16px color-mix(in srgb, var(--management-accent) 6%, transparent);
  }
  &:focus-visible {
    outline: 2px solid var(--management-accent);
    outline-offset: 3px;
  }
  &__icon {
    display: grid;
    width: 48px;
    height: 48px;
    flex-shrink: 0;
    place-items: center;
    border-radius: 12px;
    background: color-mix(in srgb, var(--home-icon-color) 9%, var(--management-surface));
    color: var(--home-icon-color);
    font-size: 25px;
  }
  &__body {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 8px;
  }
  &__label {
    color: var(--management-muted);
    font-size: 13px;
  }
  &__number {
    font-size: 32px;
    font-weight: 650;
    font-variant-numeric: tabular-nums;
    line-height: 1.2;
    white-space: nowrap;
  }
  small {
    margin-left: 6px;
    color: var(--management-muted);
    font-size: 12px;
    font-weight: 400;
  }
}
.home-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(340px, 1fr);
  align-items: stretch;
  gap: 20px;
}
.home-panel {
  min-width: 0;

  :deep(.n-card-header) {
    gap: 16px;
    align-items: flex-start;
  }

  :deep(.n-card-header__main) {
    min-width: 0;
  }
  &__heading {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  &__heading h2,
  &__title {
    margin: 0;
    font-size: 17px;
    font-weight: 600;
  }
  &__caption,
  &__note {
    color: var(--management-muted);
    font-size: 12px;
    font-weight: 400;
  }
  &__note {
    margin: 6px 0 0;
    line-height: 1.6;
  }
}
.home-week-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;

  &__current {
    display: flex;
    align-items: baseline;
    gap: 12px;
    color: var(--management-muted);
    font-size: 13px;
    strong {
      color: var(--management-text);
      font-size: 28px;
      font-weight: 650;
      font-variant-numeric: tabular-nums;
    }
    small {
      margin-left: 6px;
      color: var(--management-muted);
      font-size: 12px;
      font-weight: 400;
    }
  }

  &__total {
    color: var(--management-muted);
    font-size: 13px;
    font-variant-numeric: tabular-nums;
  }
}
.home-side {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.home-shortcuts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}
.home-shortcut {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  padding: 16px 12px;
  border: 1px solid var(--management-border);
  border-radius: 10px;
  background: var(--management-surface);
  color: var(--management-text);
  text-align: left;
  cursor: pointer;
  transition: background-color 160ms;

  &:hover {
    background: var(--management-inset);
  }
  &:focus-visible {
    outline: 2px solid var(--management-accent);
    outline-offset: 2px;
  }
  &__icon {
    display: grid;
    width: 38px;
    height: 38px;
    flex-shrink: 0;
    place-items: center;
    border-radius: 10px;
    background: color-mix(in srgb, var(--home-icon-color) 8%, var(--management-surface));
    color: var(--home-icon-color);
    font-size: 23px;
  }
  &__body {
    min-width: 0;
  }
  strong {
    display: block;
    font-size: 13px;
    font-weight: 500;
  }
  small {
    display: block;
    margin-top: 4px;
    color: var(--management-muted);
    font-size: 11px;
  }
  &__arrow {
    display: inline-flex;
    flex-shrink: 0;
    margin-left: auto;
    color: var(--management-muted);
    font-size: 13px;
  }
}
.home-version {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;

  &__icon {
    display: inline-flex;
    flex-shrink: 0;
    color: var(--home-teal);
    font-size: 32px;
  }
  strong,
  > div > span {
    display: block;
  }
  strong {
    font-size: 14px;
    font-weight: 500;
  }
  > div > span {
    margin-top: 4px;
    color: var(--management-muted);
    font-size: 13px;
  }
  :deep(.n-button) {
    margin-left: auto;
    flex-shrink: 0;
  }
  > div {
    min-width: 0;
  }
  &__number {
    color: var(--management-text);
    white-space: nowrap;
  }
  &__notes {
    margin: 14px 0 0;
    padding-top: 12px;
    border-top: 1px solid var(--management-border);
    color: var(--management-muted);
    font-size: 12px;
    line-height: 1.7;
  }
}
:deep(.home-student__id) {
  display: block;
  margin-top: 4px;
  color: var(--management-muted);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}
@media (max-width: 1100px) {
  .home-statistics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .home-grid {
    grid-template-columns: minmax(0, 1.5fr) minmax(320px, 1fr);
  }
  .home-shortcuts {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 960px) {
  .home-grid {
    grid-template-columns: minmax(0, 1fr);
  }
  .home-shortcuts {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 480px) {
  .home-statistics {
    gap: 10px;
  }
  .home-statistic {
    gap: 8px;
    padding: 18px 12px;
  }
  .home-statistic__icon {
    width: 30px;
    height: 30px;
    font-size: 18px;
  }
  .home-statistic__number {
    font-size: 24px;
  }
  .home-statistic__number small {
    display: block;
    margin: 4px 0 0;
  }
  .home-page .management-header__title {
    font-size: 24px;
  }
  .home-panel :deep(.n-card-header) {
    flex-wrap: wrap;
  }
  .home-panel :deep(.n-card-header__extra) {
    margin-left: 0;
  }
  .home-shortcuts {
    grid-template-columns: 1fr;
  }
  .home-panel :deep(.n-card-header) {
    padding: 18px 14px;
  }
  .home-panel :deep(.n-card__content) {
    padding: 0 14px 18px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .home-statistic,
  .home-shortcut {
    transition: none;
  }
}
</style>
