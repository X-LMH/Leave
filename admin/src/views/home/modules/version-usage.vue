<script setup lang="ts">
import { computed, onDeactivated, onUnmounted, ref, watch } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import type { OverviewVersionUsage } from '@/service/api/overview';
import { HOME_CHART_ANIMATION_DURATION } from '../motion';
import AnimatedCount from './animated-count.vue';

const props = defineProps<{
  usage: OverviewVersionUsage;
  animate: boolean;
}>();
defineEmits<{ manage: [] }>();
const appStore = useAppStore();

type VersionGroup = OverviewVersionUsage['groups'][number];
// Keep segment identities stable while the entire bar is revealed from left to right.
const groupKeys: VersionGroup['key'][] = ['latest', 'previous', 'older', 'unreported', 'other'];
const barGroups = computed(() =>
  groupKeys.map(
    key => props.usage.groups.find(group => group.key === key) ?? { key, version: '', count: 0, percentage: 0 }
  )
);
const lastVisibleSegment = computed(() => barGroups.value.findLastIndex(group => group.percentage > 0));
const barRef = ref<HTMLElement | null>(null);
const revealed = ref(false);
let revealAnimation: Animation | null = null;

function finishReveal() {
  revealAnimation?.cancel();
  revealAnimation = null;
  revealed.value = true;
}

function startReveal() {
  finishReveal();
  if (!props.animate || !barRef.value) return;
  revealAnimation = barRef.value.animate([{ clipPath: 'inset(0 100% 0 0)' }, { clipPath: 'inset(0 0% 0 0)' }], {
    duration: HOME_CHART_ANIMATION_DURATION,
    easing: 'linear',
    fill: 'both'
  });
}

watch(
  () => props.animate,
  animate => {
    if (!animate) finishReveal();
  }
);
onDeactivated(finishReveal);
onUnmounted(finishReveal);
defineExpose({ startReveal, finishReveal });

function versionLabel(group: VersionGroup) {
  const labels = {
    latest: $t('page.home.versionLatestGroup'),
    previous: $t('page.home.versionPreviousGroup'),
    older: $t('page.home.versionOlderGroup'),
    unreported: $t('page.home.versionUnreported'),
    other: $t('page.home.versionOtherGroup')
  };
  return group.version ? `${labels[group.key]} · v${group.version}` : labels[group.key];
}

function percentage(value: number) {
  return new Intl.NumberFormat(appStore.locale, { maximumFractionDigits: 1 }).format(value);
}

function color(group: VersionGroup) {
  const colors = {
    latest: 'var(--management-accent)',
    previous: 'var(--home-blue)',
    older: 'var(--home-teal)',
    unreported: 'var(--management-muted)',
    other: 'var(--home-orange)'
  };
  return colors[group.key];
}
</script>

<template>
  <div class="version-usage">
    <template v-if="usage.total">
      <div class="version-usage__summary">
        <span>{{ $t('page.home.versionSampleTotal') }}</span>
        <strong>
          <AnimatedCount :value="usage.total" :animate="animate" />
          <small>{{ $t('page.home.peopleUnit') }}</small>
        </strong>
      </div>
      <div
        ref="barRef"
        class="version-usage__bar"
        :class="{ 'version-usage__bar--waiting': animate && !revealed }"
        aria-hidden="true"
      >
        <span
          v-for="(group, index) in barGroups"
          :key="group.key"
          :style="{
            width: `${group.percentage}%`,
            backgroundColor: color(group),
            borderRightWidth: `${index === lastVisibleSegment ? 0 : Math.min(2, group.percentage)}px`
          }"
          :title="`${versionLabel(group)} · ${group.count}`"
        ></span>
      </div>
      <ul class="version-usage__legend">
        <li v-for="group in usage.groups" :key="group.key">
          <span class="version-usage__dot" :style="{ backgroundColor: color(group) }"></span>
          <span class="version-usage__name">{{ versionLabel(group) }}</span>
          <strong>
            <AnimatedCount :value="group.count" :animate="animate" />
            {{ $t('page.home.peopleUnit') }}
          </strong>
          <span class="version-usage__percent">{{ percentage(group.percentage) }}%</span>
        </li>
      </ul>
    </template>
    <NEmpty v-else :description="$t('page.home.noStudents')" size="small" />
    <NButton class="version-usage__link" text type="primary" @click="$emit('manage')">
      {{ $t('page.home.manageVersions') }}
      <span class="version-usage__arrow"><SvgIcon icon="lucide:arrow-right" /></span>
    </NButton>
  </div>
</template>

<style scoped lang="scss">
.version-usage {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  height: 100%;
  &__summary {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
    color: var(--management-muted);
    font-size: 13px;
  }
  &__summary strong {
    color: var(--management-text);
    font-size: 28px;
    font-weight: 650;
    font-variant-numeric: tabular-nums;
  }
  &__summary small {
    margin-left: 6px;
    font-size: 12px;
    color: var(--management-muted);
    font-weight: 400;
  }
  &__bar {
    display: flex;
    width: 100%;
    height: 14px;
    overflow: hidden;
    border-radius: 7px;
    background: var(--management-inset);
  }
  &__bar span {
    flex-shrink: 0;
    border-right: 2px solid var(--management-surface);
    box-sizing: border-box;
  }
  &__bar--waiting {
    clip-path: inset(0 100% 0 0);
  }
  &__legend {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 0;
    margin: 0;
    list-style: none;
    max-height: 190px;
    overflow-y: auto;
  }
  &__legend li {
    display: grid;
    grid-template-columns: 8px minmax(0, 1fr) auto 54px;
    gap: 10px;
    align-items: center;
    font-size: 13px;
  }
  &__legend strong {
    font-weight: 500;
    font-variant-numeric: tabular-nums;
  }
  &__name {
    overflow-wrap: anywhere;
  }
  &__dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  &__percent {
    color: var(--management-muted);
    text-align: right;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }
  &__link {
    margin-top: auto;
    align-self: flex-end;
  }
  &__arrow {
    display: inline-flex;
    margin-left: 7px;
    font-size: 16px;
  }
}
</style>
