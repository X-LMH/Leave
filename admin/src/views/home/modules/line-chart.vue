<script setup lang="ts">
import { watch } from 'vue';
import { useThemeVars } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { useEcharts } from '@/hooks/common/echarts';
import { $t } from '@/locales';
import type { WeeklyApplicationPoint } from '@/service/api/overview';
import { formatOverviewDateRange } from '../date-range';
import { HOME_CHART_ANIMATION_DURATION } from '../motion';

const props = defineProps<{ weeks: WeeklyApplicationPoint[]; animate: boolean }>();
const emit = defineEmits<{ draw: []; finished: [] }>();
const appStore = useAppStore();
const themeStore = useThemeStore();
const themeVars = useThemeVars();
let animationEndsAt = 0;

function options() {
  const dates = new Intl.DateTimeFormat(appStore.locale, {
    timeZone: 'Asia/Shanghai',
    month: '2-digit',
    day: '2-digit'
  });
  return {
    animation: props.animate,
    animationDuration: HOME_CHART_ANIMATION_DURATION,
    tooltip: {
      trigger: 'axis' as const,
      valueFormatter: (value: unknown) => `${value} ${$t('page.home.recordsUnit')}`
    },
    grid: { left: 36, right: 20, top: 32, bottom: 48 },
    xAxis: {
      type: 'category' as const,
      boundaryGap: false,
      data: props.weeks.map(week => {
        const range = formatOverviewDateRange(week.start, week.end, appStore.locale, true);
        return week.isCurrent ? `${range} (${$t('page.home.currentWeek')} · ${$t('page.home.asOfNow')})` : range;
      }),
      axisTick: { show: false },
      axisLine: { lineStyle: { color: themeVars.value.dividerColor } },
      axisLabel: {
        color: themeVars.value.textColor3,
        interval: appStore.isMobile ? ('auto' as const) : 0,
        fontSize: 11,
        lineHeight: 18,
        formatter: (_value: string, index: number) => {
          const week = props.weeks[index];
          return `${dates.format(week.start)}${week.isCurrent ? `\n${$t('page.home.currentWeek')}` : ''}`;
        }
      }
    },
    yAxis: {
      type: 'value' as const,
      min: 0,
      minInterval: 1,
      splitNumber: 4,
      axisLabel: { color: themeVars.value.textColor3 },
      splitLine: { lineStyle: { color: themeVars.value.dividerColor, type: 'dashed' as const } }
    },
    series: [
      {
        name: $t('page.home.applications'),
        type: 'line' as const,
        data: props.weeks.map(week => ({
          value: week.count,
          itemStyle: {
            color: week.isCurrent ? themeVars.value.cardColor : themeVars.value.primaryColor,
            borderColor: themeVars.value.primaryColor,
            borderWidth: 2
          }
        })),
        connectNulls: false,
        symbol: 'circle',
        symbolSize: 7,
        lineStyle: { width: 2.5, color: themeVars.value.primaryColor },
        itemStyle: {
          color: themeVars.value.cardColor,
          borderColor: themeVars.value.primaryColor,
          borderWidth: 2
        },
        label: {
          show: !appStore.isMobile,
          position: 'top' as const,
          color: themeVars.value.primaryColor,
          fontSize: 11
        },
        areaStyle: { color: themeVars.value.primaryColor, opacity: 0.06 }
      }
    ]
  };
}

const { domRef, updateOptions } = useEcharts(options, {
  onRender: instance => {
    animationEndsAt = props.animate ? Date.now() + HOME_CHART_ANIMATION_DURATION : 0;
    instance.on('finished', () => {
      animationEndsAt = 0;
      emit('finished');
    });
  },
  onUpdated: instance => {
    instance.hideLoading();
    animationEndsAt = props.animate ? Date.now() + HOME_CHART_ANIMATION_DURATION : 0;
    emit('draw');
  },
  resizeAnimation: () => {
    // Route transitions can resize the canvas mid-reveal; ECharts otherwise stops it immediately.
    return {
      duration: props.animate ? Math.max(0, animationEndsAt - Date.now()) : 0,
      easing: 'linear'
    };
  }
});
watch(
  [
    () => props.weeks,
    () => props.animate,
    () => appStore.locale,
    () => appStore.isMobile,
    () => themeStore.themeColor,
    themeVars
  ],
  () => {
    void updateOptions(() => options());
  }
);
</script>

<template>
  <div ref="domRef" class="h-280px min-w-0" role="img" :aria-label="$t('page.home.trend')"></div>
</template>
