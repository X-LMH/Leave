import { computed, toValue, watch } from 'vue';
import type { MaybeRefOrGetter } from 'vue';
import { TransitionPresets, useTransition } from '@vueuse/core';

export const HOME_CHART_ANIMATION_DURATION = 1000;

/** Keep the visible baseline when motion is re-enabled after silent data updates. */
export function useHomeTransition(values: MaybeRefOrGetter<number[]>, enabled: MaybeRefOrGetter<boolean>) {
  // Including the switch makes re-enabling synchronize VueUse's cached output.
  const source = computed(() => [...toValue(values), toValue(enabled) ? 1 : 0]);
  let baseline = source.value;
  let pending = false;
  const animated = useTransition<number[]>(source, {
    duration: HOME_CHART_ANIMATION_DURATION,
    transition: TransitionPresets.linear,
    disabled: computed(() => !toValue(enabled)),
    onStarted: () => {
      pending = false;
    },
    interpolation: (_from, to, progress) =>
      to.map((value, index) => baseline[index] + (value - baseline[index]) * progress)
  });
  watch(
    source,
    (_next, previous) => {
      if (!toValue(enabled)) {
        pending = false;
        return;
      }
      // Multiple changes in one render share the same visible starting point.
      if (pending) return;
      baseline = previous[previous.length - 1] ? [...animated.value] : previous;
      pending = true;
    },
    { flush: 'sync' }
  );
  return computed(() => animated.value.slice(0, -1));
}
