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
      <div class="login-card__brand">
        <div class="login-card__logo">G</div>
        <div>
          <h1>{{ appTitle }}</h1>
          <p>Phase 11 Auth · JWT 登录与会话管理</p>
        </div>
      </div>

      <div class="login-card__body">
        <div class="login-card__hero">
          <h2>欢迎回来</h2>
          <p>请输入后端创建的账号登录系统。开发环境默认账号为 <strong>admin / admin123</strong>。</p>
          <el-tag effect="plain" round type="info">API 基址：{{ apiBaseUrl }}</el-tag>
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

          <el-button class="login-form__submit" type="primary" :loading="loading" @click="onSubmit">
            登录系统
          </el-button>
        </el-form>
      </div>
    </el-card>
  </div>
</template>
