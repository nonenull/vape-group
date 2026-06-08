<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)

const redirectTarget = computed(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' && redirect ? redirect : '/'
})

async function submitLogin() {
  errorMessage.value = ''
  submitting.value = true
  try {
    await store.login(username.value, password.value)
    await store.bootstrap()
    await router.replace(redirectTarget.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败，请稍后重试。'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="login-shell">
    <div class="login-card">
      <div class="login-copy">
        <p class="login-kicker">Vape Group Admin</p>
        <h1>后台登录</h1>
        <p>请输入管理员账号与密码后进入后台。</p>
      </div>

      <form class="login-form" @submit.prevent="submitLogin">
        <label>
          <span>用户名</span>
          <input v-model.trim="username" class="admin-input" type="text" autocomplete="username" required>
        </label>

        <label>
          <span>密码</span>
          <input v-model="password" class="admin-input" type="password" autocomplete="current-password" required>
        </label>

        <p v-if="errorMessage" class="login-error">{{ errorMessage }}</p>

        <button class="admin-primary login-submit" type="submit" :disabled="submitting">
          {{ submitting ? '登录中...' : '登录后台' }}
        </button>
      </form>
    </div>
  </section>
</template>

<style scoped>
.login-shell {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background:
    radial-gradient(circle at top left, rgba(34, 113, 177, 0.2), transparent 36%),
    linear-gradient(180deg, #f3f6f8 0%, #e9eef3 100%);
}

.login-card {
  width: min(100%, 420px);
  display: grid;
  gap: 1.25rem;
  padding: 1.5rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(34, 113, 177, 0.14);
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.12);
}

.login-copy {
  display: grid;
  gap: 0.35rem;
}

.login-kicker {
  color: var(--wp-blue);
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.login-copy p:last-child {
  color: var(--wp-text-muted);
}

.login-form {
  display: grid;
  gap: 0.9rem;
}

.login-form label {
  display: grid;
  gap: 0.45rem;
}

.login-form span {
  font-weight: 600;
}

.login-error {
  color: var(--wp-red);
  font-size: 0.92rem;
}

.login-submit {
  width: 100%;
}
</style>
