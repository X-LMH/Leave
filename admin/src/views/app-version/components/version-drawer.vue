<script setup lang="ts">
import { computed } from 'vue';
import { useThemeVars } from 'naive-ui';
import { useAppStore } from '@/store/modules/app';
import type { AppVersionRecord } from '@/service/mock/app-version';
import { formatTime } from '@/service/mock/management';

const props = defineProps<{ show: boolean; record: AppVersionRecord | null }>();
const emit = defineEmits<{ 'update:show': [value: boolean] }>();
const appStore = useAppStore();
const theme = useThemeVars();
const themeStyle = computed(() => ({
  '--version-text': theme.value.textColor1,
  '--version-muted': theme.value.textColor3,
  '--version-border': theme.value.dividerColor
}));
async function copyPackage() {
  if (!props.record) return;
  try {
    await navigator.clipboard.writeText(props.record.packageFile);
    window.$message?.success('文件名已复制');
  } catch {
    window.$message?.error('复制失败，请手动复制文件名');
  }
}
</script>

<template>
  <NDrawer :show="show" :width="appStore.isMobile ? '100%' : 660" @update:show="emit('update:show', $event)">
    <NDrawerContent class="version-drawer" :style="themeStyle" title="版本详情" :native-scrollbar="false" closable>
      <template v-if="record">
        <div class="version-identity">
          <span class="version-kicker">ANDROID RELEASE</span>
          <h2>{{ record.versionName }}</h2>
          <div class="version-identity__meta">
            <span>构建 {{ record.versionCode }}</span>
            <NTag :type="record.status === 'published' ? 'success' : 'default'" :bordered="false" size="small">
              {{ record.status === 'published' ? '最新版本' : '历史版本' }}
            </NTag>
          </div>
        </div>
        <section class="version-section">
          <h3>版本信息</h3>
          <dl class="version-facts">
            <div>
              <dt>记录 ID</dt>
              <dd>{{ record.id }}</dd>
            </div>
            <div>
              <dt>客户端平台</dt>
              <dd>Android</dd>
            </div>
            <div>
              <dt>发布时间</dt>
              <dd>{{ formatTime(record.publishedAt) }}</dd>
            </div>
            <div>
              <dt>创建时间</dt>
              <dd>{{ formatTime(record.createdAt) }}</dd>
            </div>
            <div>
              <dt>更新时间</dt>
              <dd>{{ formatTime(record.updatedAt) }}</dd>
            </div>
          </dl>
        </section>
        <section class="version-section">
          <h3>安装包</h3>
          <div class="version-package">
            <code>{{ record.packageFile }}</code>
            <NButton size="small" @click="copyPackage">复制文件名</NButton>
          </div>
        </section>
        <section class="version-section">
          <h3>
            更新说明
            <span class="version-muted">{{ record.releaseNotes?.length ?? 0 }} 条</span>
          </h3>
          <ol v-if="record.releaseNotes?.length" class="version-notes">
            <li v-for="(note, index) in record.releaseNotes" :key="index">{{ note }}</li>
          </ol>
          <p v-else class="version-muted">该版本未填写更新说明。</p>
        </section>
      </template>
      <template #footer>
        <div class="version-footer">
          <span class="version-muted">客户端版本发布记录</span>
          <NButton @click="emit('update:show', false)">关闭</NButton>
        </div>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped lang="scss">
.version-drawer {
  color: var(--version-text);
}
.version-kicker {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 1.5px;
  color: var(--version-muted);
}
.version-identity {
  padding: 4px 0 20px;
  border-bottom: 1px solid var(--version-border);
}
.version-identity h2 {
  margin: 10px 0;
  font-size: 32px;
  overflow-wrap: anywhere;
  font-weight: 650;
  letter-spacing: -1px;
  font-variant-numeric: tabular-nums;
}
.version-identity__meta {
  display: flex;
  align-items: center;
  gap: 16px;
  color: var(--version-muted);
}
.version-section {
  margin: 20px 0;
}
.version-section h3 {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 600;
}
.version-section h3 .version-muted {
  margin-left: auto;
  font-size: 12px;
  font-weight: 400;
}
.version-facts {
  margin: 0;
}
.version-facts div {
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 12px;
  padding: 7px 0;
}
.version-facts dt,
.version-muted {
  color: var(--version-muted);
}
.version-facts dd {
  margin: 0;
  overflow-wrap: anywhere;
}
.version-package {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--version-border);
  border-radius: 8px;
}
.version-package code {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
  font-size: 13px;
}
.version-notes {
  padding-left: 22px;
  margin: 12px 0 0;
}
.version-notes li {
  padding: 5px 0 5px 5px;
  line-height: 1.7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.version-notes li::marker {
  color: var(--version-muted);
  font-size: 12px;
}
.version-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}
.version-footer > span {
  font-size: 12px;
}
@media (max-width: 480px) {
  .version-footer > span {
    display: none;
  }
  .version-footer {
    justify-content: flex-end;
  }
}
</style>
