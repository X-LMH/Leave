<script setup lang="ts">
import { useAppStore } from '@/store/modules/app';
import { $t } from '@/locales';
import type { OverviewVersionUsage } from '@/service/api/overview';

defineProps<{
  usage: OverviewVersionUsage;
  latestVersion: string | null;
}>();
defineEmits<{ manage: [] }>();
const appStore = useAppStore();

type VersionGroup = OverviewVersionUsage['groups'][number];
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
          {{ usage.total.toLocaleString(appStore.locale) }}
          <small>{{ $t('page.home.peopleUnit') }}</small>
        </strong>
      </div>
      <div class="version-usage__bar" aria-hidden="true">
        <span
          v-for="group in usage.groups.filter(item => item.count > 0)"
          :key="group.key"
          :style="{ width: `${group.percentage}%`, backgroundColor: color(group) }"
          :title="`${versionLabel(group)} · ${group.count}`"
        ></span>
      </div>
      <ul class="version-usage__legend">
        <li v-for="group in usage.groups" :key="group.key">
          <span class="version-usage__dot" :style="{ backgroundColor: color(group) }"></span>
          <span class="version-usage__name">{{ versionLabel(group) }}</span>
          <strong>{{ group.count.toLocaleString(appStore.locale) }} {{ $t('page.home.peopleUnit') }}</strong>
          <span class="version-usage__percent">{{ percentage(group.percentage) }}%</span>
        </li>
      </ul>
    </template>
    <NEmpty v-else :description="$t('page.home.noStudents')" size="small" />
    <div class="version-usage__latest">
      <template v-if="latestVersion">
        <span>
          {{ $t('page.home.latestAndroidVersion') }}
          <strong>v{{ latestVersion }}</strong>
        </span>
        <span class="version-usage__adoption">
          {{ $t('page.home.versionAdoption', { count: usage.latestCount.toLocaleString(appStore.locale) }) }}
        </span>
      </template>
      <span v-else>{{ $t('page.home.noVersion') }}</span>
    </div>
    <p class="version-usage__note">{{ $t('page.home.versionUsageNote') }}</p>
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
  gap: 20px;
  min-width: 0;
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
  &__bar span:last-child {
    border-right: 0;
  }
  &__legend {
    display: flex;
    flex-direction: column;
    gap: 18px;
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
  &__latest {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 8px;
    padding: 12px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--management-accent) 7%, var(--management-surface));
    font-size: 12px;
  }
  &__latest strong {
    color: var(--management-accent);
    font-weight: 600;
    overflow-wrap: anywhere;
  }
  &__adoption {
    color: var(--management-muted);
  }
  &__note {
    margin: 0;
    color: var(--management-muted);
    font-size: 12px;
    line-height: 1.7;
  }
  &__link {
    align-self: flex-end;
  }
  &__arrow {
    display: inline-flex;
    margin-left: 7px;
    font-size: 16px;
  }
}
</style>
