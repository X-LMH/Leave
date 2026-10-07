<script setup lang="ts">
import { computed } from 'vue';
import { NTag, useThemeVars } from 'naive-ui';
import type { StudentListItem } from '@/service/api/student';

const props = defineProps<{
  student: Pick<StudentListItem, 'appVersion' | 'appVersionStatus'>;
}>();
const theme = useThemeVars();
const status = computed(() => props.student.appVersionStatus || 'unknown');
const label = computed(
  () => ({ latest: '最新版', previous: '旧版本', outdated: '版本过旧', unknown: '版本未知' })[status.value]
);
const style = computed(() => ({
  '--version-text': status.value === 'outdated' ? theme.value.errorColor : theme.value.textColor1
}));
</script>

<template>
  <div v-if="student.appVersion.trim()" class="student-version" :style="style">
    <div class="student-version__main">
      <span class="student-version__number">{{ student.appVersion }}</span>
      <NTag
        :type="
          status === 'latest'
            ? 'success'
            : status === 'outdated'
              ? 'error'
              : status === 'previous'
                ? 'warning'
                : 'default'
        "
        :bordered="false"
        size="small"
        round
      >
        {{ label }}
      </NTag>
    </div>
  </div>
  <span v-else :style="{ color: theme.textColor3 }">暂无</span>
</template>

<style scoped>
.student-version {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-width: 0;
}
.student-version__main {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.student-version__number {
  color: var(--version-text);
  font-weight: 600;
  overflow-wrap: anywhere;
  font-variant-numeric: tabular-nums;
}
</style>
