<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { useHomeTransition } from '../motion';

const props = defineProps<{
  value: number;
  animate: boolean;
  format?: (count: string) => string;
}>();
const appStore = useAppStore();
const target = ref(0);
const displayed = useHomeTransition(
  () => [target.value],
  () => props.animate
);

function formatValue(value: number) {
  const count = Math.round(value).toLocaleString(appStore.locale);
  return props.format ? props.format(count) : count;
}

onMounted(() => {
  target.value = props.value;
});
watch(
  () => props.value,
  value => {
    target.value = value;
  }
);
</script>

<template>
  <span :aria-label="formatValue(value)">
    <span aria-hidden="true">{{ formatValue(displayed[0]) }}</span>
  </span>
</template>
