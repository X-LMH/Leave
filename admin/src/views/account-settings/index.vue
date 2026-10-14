<script setup lang="ts">
import { computed, onDeactivated, onUnmounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useMediaQuery } from '@vueuse/core';
import { useThemeVars } from 'naive-ui';
import type { FormInst, FormRules } from 'naive-ui';
import { useAuthStore } from '@/store/modules/auth';
import { changeAdminPassword } from '@/service/api/auth';
import { $t } from '@/locales';

const authStore = useAuthStore();
const narrowScreen = useMediaQuery('(max-width: 720px)');
const themeVars = useThemeVars();
const themeStyle = computed(() => ({
  '--management-accent': themeVars.value.primaryColor,
  '--management-text': themeVars.value.textColor1,
  '--management-muted': themeVars.value.textColor3,
  '--management-border': themeVars.value.dividerColor,
  '--management-surface': themeVars.value.cardColor
}));
const formRef = ref<FormInst | null>(null);
const model = reactive({ currentPassword: '', newPassword: '', confirmPassword: '' });
const saving = ref(false);
let submitSequence = 0;

const rules: FormRules = {
  currentPassword: {
    required: true,
    validator: (_rule, value: string) => Boolean(value?.length),
    message: '请输入原密码',
    trigger: ['input', 'blur']
  },
  newPassword: [
    {
      required: true,
      validator: (_rule, value: string) => Boolean(value?.length),
      message: '请输入新密码',
      trigger: ['input', 'blur']
    },
    {
      validator: (_rule, value: string) => {
        if (!value) return true;
        const length = new TextEncoder().encode(value).length;
        // Unicode White_Space matches Go's strings.TrimSpace, including U+0085.
        return length >= 6 && length <= 30 && !/^\p{White_Space}*$/u.test(value);
      },
      message: '密码长度至少为6位',
      trigger: ['input', 'blur']
    },
    {
      validator: (_rule, value: string) => !value || value !== model.currentPassword,
      message: '新密码不能与原密码相同',
      trigger: ['input', 'blur']
    }
  ],
  confirmPassword: [
    {
      required: true,
      validator: (_rule, value: string) => Boolean(value?.length),
      message: '请再次输入新密码',
      trigger: ['input', 'blur']
    },
    {
      validator: (_rule, value: string) => !value || value === model.newPassword,
      message: '两次输入的新密码不一致',
      trigger: ['input', 'blur']
    }
  ]
};

function resetForm() {
  submitSequence += 1;
  saving.value = false;
  model.currentPassword = '';
  model.newPassword = '';
  model.confirmPassword = '';
  formRef.value?.restoreValidation();
}

async function submit() {
  if (saving.value || !formRef.value) return;
  const sequence = ++submitSequence;
  saving.value = true;
  try {
    try {
      await formRef.value.validate();
    } catch {
      return;
    }
    if (sequence !== submitSequence) return;
    const { error } = await changeAdminPassword(model.currentPassword, model.newPassword, model.confirmPassword);
    if (error) return;
    // 密码已修改成功，即使表单已重置或页面已离开，也必须退出当前登录。
    resetForm();
    window.$message?.success('密码修改成功，请重新登录');
    await authStore.resetStore();
  } catch {
    if (sequence === submitSequence) window.$message?.error('提交失败，请重试');
  } finally {
    if (sequence === submitSequence) saving.value = false;
  }
}

// Clear sensitive inputs and cancel pending validation when leaving, including cached pages.
onBeforeRouteLeave(() => {
  resetForm();
});
onDeactivated(resetForm);
onUnmounted(resetForm);
</script>

<template>
  <div class="management-page account-settings" :style="themeStyle">
    <header class="management-header">
      <div class="management-header__intro">
        <span class="management-header__icon"><SvgIcon icon="mdi:account-cog-outline" /></span>
        <div>
          <h1 class="management-header__title">{{ $t('route.account-settings') }}</h1>
          <p class="management-header__description">查看当前管理员账号，维护登录密码。</p>
        </div>
      </div>
    </header>

    <NCard :bordered="false" class="management-card account-settings__card">
      <section class="account-settings__account" aria-label="当前账号">
        <dl class="account-settings__identity">
          <dt>登录账号</dt>
          <dd>{{ authStore.userInfo.userName }}</dd>
        </dl>
      </section>

      <section aria-labelledby="account-password-title">
        <h2 id="account-password-title" class="account-settings__section-title">修改密码</h2>
        <p class="account-settings__notice">修改密码后将退出登录，请使用新密码重新登录。</p>
        <NForm
          ref="formRef"
          :model="model"
          :rules="rules"
          :disabled="saving"
          :label-placement="narrowScreen ? 'top' : 'left'"
          :label-width="narrowScreen ? undefined : 104"
          label-align="left"
          require-mark-placement="right"
          class="account-settings__form"
          @submit.prevent="submit"
        >
          <NFormItem label="原密码" path="currentPassword">
            <NInput
              v-model:value="model.currentPassword"
              type="password"
              show-password-on="click"
              autocomplete="current-password"
              placeholder="请输入原密码"
              :input-props="{ 'aria-label': '原密码' }"
            />
          </NFormItem>
          <NFormItem label="新密码" path="newPassword">
            <NInput
              v-model:value="model.newPassword"
              type="password"
              show-password-on="click"
              autocomplete="new-password"
              placeholder="请输入新密码"
              :input-props="{ 'aria-label': '新密码' }"
            />
          </NFormItem>
          <NFormItem label="确认新密码" path="confirmPassword">
            <NInput
              v-model:value="model.confirmPassword"
              type="password"
              show-password-on="click"
              autocomplete="new-password"
              placeholder="请再次输入新密码"
              :input-props="{ 'aria-label': '确认新密码' }"
            />
          </NFormItem>
          <div class="account-settings__actions">
            <NButton type="primary" attr-type="submit" :loading="saving" :disabled="saving">保存修改</NButton>
            <NButton :disabled="saving" @click="resetForm">重置</NButton>
          </div>
        </NForm>
      </section>
    </NCard>
  </div>
</template>

<style scoped lang="scss">
@use '@/styles/scss/management.scss';

.account-settings {
  --account-label-width: 104px;
  width: 100%;
  max-width: 820px;
  margin-inline: auto;
  padding-block: clamp(24px, 8vh, 88px) 32px;

  &__card :deep(.n-card__content) {
    padding: 32px;
  }

  &__account {
    margin-bottom: 28px;
    padding-bottom: 24px;
    border-bottom: 1px solid var(--management-border);
  }

  &__identity {
    display: grid;
    grid-template-columns: 104px minmax(0, 1fr);
    align-items: baseline;
    margin: 0;

    dt {
      color: var(--management-muted);
      font-size: 14px;
      line-height: 1.5;
    }

    dd {
      margin: 0;
      overflow-wrap: anywhere;
      font-size: 16px;
      font-weight: 500;
      line-height: 1.5;
    }
  }

  &__section-title {
    margin: 0;
    font-size: 17px;
    font-weight: 600;
    line-height: 1.5;
  }

  &__notice {
    margin: 8px 0 24px;
    color: var(--management-muted);
    font-size: 13px;
    line-height: 1.7;
  }

  &__form {
    :deep(.n-form-item-label) {
      font-size: 14px;
      font-weight: 500;
    }
  }

  &__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    padding-top: 4px;
    padding-left: var(--account-label-width);
  }
}

@media (max-width: 720px) {
  .account-settings {
    --account-label-width: 0px;
    padding-block: 12px 24px;

    &__card :deep(.n-card__content) {
      padding: 24px 20px;
    }

    &__identity {
      grid-template-columns: minmax(0, 1fr);
      gap: 8px;
    }
  }
}
</style>
