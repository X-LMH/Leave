<script setup lang="ts">
import { computed, h, onActivated, onDeactivated, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { useNow, usePreferredReducedMotion } from '@vueuse/core';
import { NSkeleton, NTag, useThemeVars } from 'naive-ui';
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
import { HOME_CHART_ANIMATION_DURATION } from './motion';
import LineChart from './modules/line-chart.vue';
import VersionUsage from './modules/version-usage.vue';
import AnimatedCount from './modules/animated-count.vue';

const appStore = useAppStore();
const themeStore = useThemeStore();
const themeVars = useThemeVars();
const { routerPushByKey } = useRouterPush();
const clock = useNow({ interval: 60_000 });
const state = reactive<OverviewLoadState>({ data: null, loading: false, failed: false });
const overview = computed(() => state.data);
const weekly = computed(() => overview.value?.weeklyApplications);
const pageRef = ref<HTMLElement | null>(null);
const versionUsageRef = ref<InstanceType<typeof VersionUsage> | null>(null);
const active = ref(true);
const reducedMotion = usePreferredReducedMotion();
const motionEnabled = computed(() => active.value && themeStore.page.animate && reducedMotion.value !== 'reduce');
const presented = ref(false);
let animatedOverview: typeof state.data = null;
let contentAnimations: Animation[] = [];

function stopMotion() {
  contentAnimations.forEach(animation => animation.cancel());
  contentAnimations = [];
}

function startHomeMotion() {
  versionUsageRef.value?.startReveal();
  const data = overview.value;
  if (!data || data === animatedOverview) return;
  const first = !presented.value;
  presented.value = true;
  animatedOverview = data;
  stopMotion();
  if (!motionEnabled.value) return;

  // Refresh content in place; theme and layout redraws do not replay the whole page.
  const selector = first
    ? '.management-header, .home-statistic, .home-panel, .home-page__sample'
    : '.management-header, .home-statistic__body, .home-statistic__icon, .home-panel .n-card-header, .home-panel .n-card__content, .home-page__sample';
  const targets = pageRef.value?.querySelectorAll<HTMLElement>(selector);
  contentAnimations = Array.from(targets ?? [], element =>
    element.animate(
      [
        { opacity: first ? 0 : 0.8, translate: first ? '-6px 0' : '0 0' },
        { opacity: 1, translate: '0 0' }
      ],
      { duration: HOME_CHART_ANIMATION_DURATION, easing: 'linear' }
    )
  );
}

function finishHomeMotion() {
  versionUsageRef.value?.finishReveal();
  stopMotion();
}

watch(motionEnabled, enabled => {
  if (!enabled) stopMotion();
});
const { reload, invalidate } = createOverviewLoader(state, async () => {
  const result = await getOverview();
  return result.error ? null : result.data;
});
onMounted(() => void reload());
onActivated(() => {
  active.value = true;
  void reload();
});
onDeactivated(() => {
  active.value = false;
  stopMotion();
  invalidate();
});
onUnmounted(() => {
  stopMotion();
  invalidate();
});
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
const weekRange = computed(() =>
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
    value: overview.value?.statistics.unhandledFeedbackCount ?? 0,
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
const skeletonRows = Array.from({ length: 5 }, (_, id) => ({ id }));
const skeletonColumns = computed(() =>
  columns.value.map(column => ({
    ...column,
    render: () => {
      const lines = [h(NSkeleton, { height: '18px', width: '70%', animated: motionEnabled.value })];
      if ('key' in column && (column.key === 'name' || column.key === 'appliedAt')) {
        lines.push(h(NSkeleton, { height: '18px', width: '55%', animated: motionEnabled.value }));
      }
      return h('div', { class: 'home-table-placeholder', 'aria-hidden': 'true' }, lines);
    }
  }))
);

function navigate(route: string) {
  void routerPushByKey(route as RouteKey);
}
</script>

<template>
  <div
    ref="pageRef"
    class="management-page home-page"
    :class="{ 'home-page--motion': motionEnabled }"
    :style="themeStyle"
    :aria-busy="state.loading"
  >
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon" aria-hidden="true"><SvgIcon icon="mdi:monitor-dashboard" /></span>
        <div>
          <h1 class="management-header__title">{{ $t('page.home.title') }}</h1>
          <p class="management-header__description">{{ $t('page.home.description') }}</p>
        </div>
      </div>
      <div class="home-page__date">
        <span class="home-page__calendar"><SvgIcon icon="lucide:calendar-days" /></span>
        <time>{{ dateFormatter.format(clock) }}</time>
      </div>
    </header>

    <span v-if="!overview && state.loading" class="home-loading-status" role="status">
      {{ $t('page.home.loading') }}
    </span>
    <NAlert
      v-if="state.failed"
      type="error"
      :title="$t(overview ? 'page.home.refreshFailed' : 'page.home.loadFailed')"
      role="alert"
    >
      <p v-if="overview">{{ $t('page.home.updatedAt', { time: generatedAt }) }}</p>
      <NButton size="small" :loading="state.loading" @click="reload">{{ $t('page.home.retry') }}</NButton>
    </NAlert>
    <template v-if="overview || state.loading">
      <div class="home-statistics">
        <button
          v-for="item in statistics"
          :key="item.key"
          type="button"
          class="home-statistic"
          :class="`home-tone--${item.tone}`"
          :disabled="!overview"
          @click="navigate(item.route)"
        >
          <span class="home-statistic__icon"><SvgIcon :icon="item.icon" /></span>
          <span class="home-statistic__body">
            <span class="home-statistic__label">{{ item.title }}</span>
            <span class="home-statistic__number">
              <AnimatedCount v-if="overview" :value="item.value" :animate="motionEnabled" />
              <NSkeleton v-else class="home-number-placeholder" width="75px" height="32px" :animated="motionEnabled" />
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
              <span v-if="weekly" class="home-panel__caption">{{ weekRange }}</span>
              <NSkeleton v-else width="190px" height="18px" :animated="motionEnabled" aria-hidden="true" />
            </div>
          </template>
          <template #header-extra>
            <span class="home-panel__caption">{{ $t('page.home.recentWeeks') }}</span>
          </template>
          <div class="home-week-toolbar">
            <div class="home-week-toolbar__current">
              <span>{{ $t('page.home.currentApplications') }}</span>
              <strong>
                <AnimatedCount v-if="weekly" :value="weekly.weeks[8].count" :animate="motionEnabled" />
                <NSkeleton
                  v-else
                  class="home-number-placeholder"
                  width="64px"
                  height="28px"
                  :animated="motionEnabled"
                />
                <small>{{ $t('page.home.recordsUnit') }}</small>
              </strong>
            </div>
            <span class="home-week-toolbar__total">
              <AnimatedCount
                v-if="weekly"
                :value="weekly.total"
                :animate="motionEnabled"
                :format="count => $t('page.home.weekTotal', { count })"
              />
              <NSkeleton v-else width="170px" height="18px" :animated="motionEnabled" aria-hidden="true" />
            </span>
          </div>
          <LineChart
            v-if="weekly"
            :weeks="weekly.weeks"
            :animate="motionEnabled"
            @draw="startHomeMotion"
            @finished="finishHomeMotion"
          />
          <div v-else class="home-chart-placeholder" aria-hidden="true">
            <NSkeleton height="100%" :sharp="false" :animated="motionEnabled" />
          </div>
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
            v-if="overview"
            ref="versionUsageRef"
            :usage="overview.versionUsage"
            :latest-version="overview.latestVersion?.versionName ?? null"
            :animate="motionEnabled"
            @manage="navigate('app-version')"
          />
          <div v-else class="home-usage-placeholder">
            <div class="home-usage-placeholder__summary">
              <span>{{ $t('page.home.versionSampleTotal') }}</span>
              <NSkeleton width="80px" height="42px" :animated="motionEnabled" />
            </div>
            <NSkeleton height="14px" :sharp="false" :animated="motionEnabled" />
            <div class="home-usage-placeholder__legend" aria-hidden="true">
              <div v-for="item in 4" :key="item" class="home-usage-placeholder__group">
                <NSkeleton circle width="8px" height="8px" :animated="motionEnabled" />
                <NSkeleton width="80%" height="19.5px" :animated="motionEnabled" />
                <NSkeleton width="50px" height="19.5px" :animated="motionEnabled" />
                <NSkeleton width="42px" height="18px" :animated="motionEnabled" />
              </div>
            </div>
            <div class="home-usage-placeholder__latest">
              <NSkeleton width="170px" height="18px" :animated="motionEnabled" />
              <NSkeleton width="120px" height="18px" :animated="motionEnabled" />
            </div>
            <p class="home-usage-placeholder__note">{{ $t('page.home.versionUsageNote') }}</p>
            <NButton class="home-usage-placeholder__link" text type="primary" @click="navigate('app-version')">
              {{ $t('page.home.manageVersions') }}
              <span class="home-link-arrow"><SvgIcon icon="lucide:arrow-right" /></span>
            </NButton>
          </div>
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
            :columns="overview ? columns : skeletonColumns"
            :data="overview?.recentRecords ?? skeletonRows"
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
            <div v-if="overview?.latestVersion" class="home-version">
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
            <div v-if="!overview" class="home-version">
              <span class="home-version__icon"><SvgIcon icon="ant-design:android-filled" /></span>
              <div class="home-version-placeholder">
                <NSkeleton width="100px" height="21px" :animated="motionEnabled" />
                <NSkeleton width="170px" height="19px" :animated="motionEnabled" />
              </div>
              <NButton text type="primary" @click="navigate('app-version')">
                {{ $t('page.home.viewVersion') }}
                <span class="home-link-arrow"><SvgIcon icon="lucide:arrow-right" /></span>
              </NButton>
            </div>
            <p v-if="overview?.latestVersion?.releaseNotes.length" class="home-version__notes">
              {{ overview.latestVersion.releaseNotes[0] }}
            </p>
            <NEmpty v-if="overview && !overview.latestVersion" :description="$t('page.home.noVersion')" size="small">
              <template #extra>
                <NButton text type="primary" @click="navigate('app-version')">
                  {{ $t('page.home.viewVersion') }}
                </NButton>
              </template>
            </NEmpty>
          </NCard>
        </div>
      </div>
      <p class="home-page__sample">
        <template v-if="overview">{{ $t('page.home.updatedAt', { time: generatedAt }) }}</template>
        <NSkeleton v-else width="190px" height="18px" :animated="motionEnabled" aria-hidden="true" />
      </p>
    </template>
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';

.home-page {
  --home-icon-color: var(--management-accent);
  gap: 22px;

  .management-header,
  .management-card {
    animation: none;
  }

  :deep(.n-skeleton) {
    max-width: 100%;
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
    box-shadow 160ms,
    transform 160ms ease-out;

  &:enabled:hover {
    border-color: var(--management-accent);
    box-shadow: 0 4px 16px color-mix(in srgb, var(--management-accent) 6%, transparent);
  }
  &:focus-visible {
    outline: 2px solid var(--management-accent);
    outline-offset: 3px;
  }
  &:disabled {
    cursor: default;
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
  transition:
    background-color 160ms,
    border-color 160ms,
    transform 160ms ease-out;

  &:hover {
    background: var(--management-inset);
    border-color: var(--management-accent);
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
    transition: transform 160ms ease-out;
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
.home-loading-status {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
.home-number-placeholder {
  display: inline-block;
  vertical-align: middle;
}
.home-chart-placeholder {
  height: 280px;
  padding: 32px 20px 48px 36px;
  box-sizing: border-box;
}
.home-usage-placeholder {
  display: flex;
  flex-direction: column;
  gap: 20px;

  &__summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    color: var(--management-muted);
    font-size: 13px;
  }
  &__legend {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }
  &__group {
    display: grid;
    grid-template-columns: 8px minmax(0, 1fr) auto 54px;
    align-items: center;
    gap: 10px;
  }
  &__latest {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 8px;
    padding: 12px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--management-accent) 7%, var(--management-surface));
  }
  &__link {
    align-self: flex-end;
  }
  &__note {
    margin: 0;
    color: var(--management-muted);
    font-size: 12px;
    line-height: 1.7;
  }
}
.home-version-placeholder,
:deep(.home-table-placeholder) {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
@media (prefers-reduced-motion: no-preference) {
  @media (hover: hover) and (pointer: fine) {
    .home-page--motion {
      .home-statistic:enabled:hover,
      .home-shortcut:hover {
        transform: translateY(-2px);
      }
      .home-shortcut:hover .home-shortcut__arrow {
        transform: translateX(2px);
      }
    }
  }

  .home-page--motion {
    .home-statistic:enabled:active,
    .home-shortcut:active {
      transform: translateY(0);
      transition-duration: 80ms;
    }
    .home-shortcut:active .home-shortcut__arrow {
      transform: translateX(0);
      transition-duration: 80ms;
    }
  }
}

.home-page:not(.home-page--motion) {
  .home-statistic,
  .home-shortcut,
  .home-shortcut__arrow {
    transition: none;
  }
}
</style>
