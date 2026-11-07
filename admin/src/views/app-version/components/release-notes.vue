<script setup lang="ts">
import { computed } from 'vue';
import { releaseNoteSections } from '@/service/api/app-version';
import type { ReleaseNotesGrouped } from '@/service/api/app-version';
const props = defineProps<{ notes: string[]; grouped?: ReleaseNotesGrouped | null }>();
const sections = computed(() => releaseNoteSections(props.notes, props.grouped));
</script>

<template>
  <div v-if="sections.length" class="release-note-groups">
    <section v-for="section in sections" :key="section.title">
      <h4>{{ section.title }}</h4>
      <ul>
        <li v-for="(note, index) in section.items" :key="index">{{ note }}</li>
      </ul>
    </section>
  </div>
  <p v-else>该版本暂无更新说明。</p>
</template>

<style scoped>
.release-note-groups section + section {
  margin-top: 16px;
}
.release-note-groups h4 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
}
.release-note-groups ul {
  margin: 0;
  padding-left: 20px;
}
.release-note-groups li {
  line-height: 1.7;
  overflow-wrap: anywhere;
}
</style>
