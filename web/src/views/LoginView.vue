<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';

import { login } from '@/api/auth';
import { useSessionStore } from '@/store/session';

interface LoginForm {
  username: string;
  password: string;
}

const router = useRouter();
const route = useRoute();
const sessionStore = useSessionStore();
const formRef = ref<FormInstance>();
const loading = ref(false);
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';

const form = reactive<LoginForm>({
  username: 'admin',
  password: 'admin123',
});

const redirectTarget = computed(() => {
  const redirect = route.query.redirect;
  if (typeof redirect === 'string' && redirect.trim() !== '' && redirect !== '/login') {
    return redirect;
  }
  return '/dashboard';
});

const rules: FormRules<LoginForm> = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
};

async function onSubmit() {
  if (!formRef.value) {
    return;
  }
  await formRef.value.validate(async (valid) => {
    if (!valid) {
      return;
    }
    loading.value = true;
    try {
      const response = await login({ username: form.username.trim(), password: form.password });
      sessionStore.applyLoginResponse(response);
      ElMessage.success('登录成功');
      await router.replace(redirectTarget.value);
    } catch (error) {
      const message = error instanceof Error ? error.message : '登录失败';
      ElMessage.error(message);
    } finally {
      loading.value = false;
    }
  });
}
</script>

<template>
  <div class="login-page">
    <div class="login-page__backdrop" />

    <el-card class="login-card" shadow="never">
      <section class="login-card__brand">
        <div class="login-card__brand-top">
          <div class="login-card__logo">G</div>
          <div>
            <h1>{{ appTitle }}</h1>
            <p>Frontend Core · Vue 3 + TypeScript + Vite</p>
          </div>
        </div>

        <div class="login-card__brand-body">
          <el-tag effect="plain" round type="success">Phase 11 Auth</el-tag>
          <h2>欢迎使用 GoAdmin</h2>
          <p>
            参考 Gin-vue-admin 的登录页布局，左侧展示品牌信息与系统能力，右侧用于账号登录和快速进入工作台。
          </p>

          <ul class="login-card__highlights">
            <li>JWT 登录与会话管理</li>
            <li>动态菜单与权限驱动</li>
            <li>Element Plus 统一视觉风格</li>
          </ul>

          <div class="login-card__stats">
            <div>
              <strong>{{ apiBaseUrl }}</strong>
              <span>API 基址</span>
            </div>
            <div>
              <strong>admin</strong>
              <span>默认用户名</span>
            </div>
            <div>
              <strong>admin123</strong>
              <span>默认密码</span>
            </div>
          </div>
        </div>
      </section>

      <section class="login-card__panel">
        <div class="login-card__panel-header">
          <h2>登录系统</h2>
          <p>请输入后端创建的账号登录系统。</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="login-form"
          label-position="top"
          @keyup.enter="onSubmit"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="form.username" autocomplete="username" placeholder="请输入用户名" />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              autocomplete="current-password"
              placeholder="请输入密码"
              type="password"
              show-password
            />
          </el-form-item>

          <div class="login-form__meta">
            <el-tag effect="plain" round type="info">API 基址：{{ apiBaseUrl }}</el-tag>
            <span>默认账号：admin / admin123</span>
          </div>

          <el-button class="login-form__submit" type="primary" :loading="loading" @click="onSubmit">
            登录系统
          </el-button>
        </el-form>
      </section>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  position: relative;
  display: grid;
  place-items: center;
  width: 100%;
  min-height: 100vh;
  padding: 32px;
  overflow: hidden;
}

.login-page__backdrop {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.16), transparent 28%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.14), transparent 24%);
  pointer-events: none;
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(1160px, 100%);
  min-height: 680px;
  border: 0;
  border-radius: 28px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.14);
}

.login-card :deep(.el-card__body) {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(360px, 0.9fr);
  min-height: 680px;
  padding: 0;
}

.login-card__brand {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 32px;
  padding: 48px;
  color: #fff;
  background:
    linear-gradient(135deg, rgba(15, 23, 42, 0.96) 0%, rgba(17, 24, 39, 0.94) 50%, rgba(30, 64, 175, 0.92) 100%),
    linear-gradient(135deg, #0f172a 0%, #1e3a8a 100%);
}

.login-card__brand-top {
  display: flex;
  align-items: center;
  gap: 16px;
}

.login-card__logo {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  border-radius: 18px;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: #fff;
  background: linear-gradient(135deg, #4f46e5 0%, #06b6d4 100%);
  box-shadow: 0 16px 32px rgba(59, 130, 246, 0.28);
}

.login-card__brand-top h1 {
  margin: 0;
  font-size: 28px;
  line-height: 1.15;
}

.login-card__brand-top p {
  margin: 6px 0 0;
  color: rgba(255, 255, 255, 0.72);
}

.login-card__brand-body {
  display: grid;
  gap: 22px;
  max-width: 520px;
}

.login-card__brand-body h2 {
  margin: 0;
  font-size: 38px;
  line-height: 1.1;
}

.login-card__brand-body p {
  margin: 0;
  line-height: 1.9;
  color: rgba(255, 255, 255, 0.82);
}

.login-card__highlights {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.login-card__highlights li {
  position: relative;
  padding-left: 22px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.88);
}

.login-card__highlights li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 11px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #93c5fd;
  box-shadow: 0 0 0 6px rgba(59, 130, 246, 0.18);
}

.login-card__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.login-card__stats div {
  padding: 14px 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(12px);
}

.login-card__stats strong {
  display: block;
  font-size: 15px;
  line-height: 1.2;
}

.login-card__stats span {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.72);
}

.login-card__panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 56px 48px;
  background: rgba(255, 255, 255, 0.96);
}

.login-card__panel-header {
  display: grid;
  gap: 8px;
  margin-bottom: 28px;
}

.login-card__panel-header h2 {
  margin: 0;
  font-size: 30px;
}

.login-card__panel-header p {
  margin: 0;
  color: var(--app-muted);
}

.login-form {
  display: grid;
  gap: 6px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.login-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--app-text);
}

.login-form :deep(.el-input__wrapper) {
  min-height: 48px;
  border-radius: 14px;
}

.login-form__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 4px 0 8px;
  font-size: 13px;
  color: var(--app-muted);
}

.login-form__submit {
  width: 100%;
  height: 48px;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 600;
}

@media (max-width: 1024px) {
  .login-page {
    padding: 16px;
  }

  .login-card {
    min-height: auto;
  }

  .login-card :deep(.el-card__body) {
    grid-template-columns: 1fr;
  }

  .login-card__brand,
  .login-card__panel {
    padding: 32px 24px;
  }

  .login-card__brand-body h2 {
    font-size: 30px;
  }

  .login-card__stats {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .login-page {
    padding: 12px;
  }

  .login-card__brand-top h1 {
    font-size: 22px;
  }

  .login-card__panel-header h2 {
    font-size: 24px;
  }

  .login-card__brand-body h2 {
    font-size: 26px;
  }
}
</style>
