<script setup>
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const mode = ref('login')
const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const form = reactive({
  displayName: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const isRegister = computed(() => mode.value === 'register')
const title = computed(() => (isRegister.value ? 'Crea tu cuenta' : 'Inicia sesión'))
const submitLabel = computed(() => {
  if (loading.value) return isRegister.value ? 'Creando cuenta...' : 'Entrando...'
  return isRegister.value ? 'Crear cuenta' : 'Entrar'
})
const toggleLabel = computed(() => (isRegister.value ? 'Inicia sesión' : 'Crea una cuenta'))

function resetFeedback() {
  errorMessage.value = ''
  successMessage.value = ''
}

function switchMode() {
  mode.value = isRegister.value ? 'login' : 'register'
  form.password = ''
  form.confirmPassword = ''
  resetFeedback()
}

function translateAuthError(error) {
  const msg = error?.message?.toLowerCase?.() ?? ''
  // Correo ya registrado (nuestro marcador o el mensaje nativo de Supabase).
  if (
    msg.includes('email_taken') ||
    msg.includes('already registered') ||
    msg.includes('user already')
  ) {
    return 'Ese correo ya está registrado. Inicia sesión.'
  }
  if (msg.includes('invalid login credentials')) return 'El correo o la contraseña no coinciden.'
  if (msg.includes('email')) return 'Revisa que el correo esté bien escrito.'
  if (msg.includes('password')) return 'La contraseña debe tener al menos 6 caracteres.'
  return error?.message || 'Algo salió mal. Intenta de nuevo.'
}

async function submit() {
  resetFeedback()
  if (isRegister.value && form.password !== form.confirmPassword) {
    errorMessage.value = 'Las contraseñas no coinciden.'
    return
  }
  loading.value = true
  try {
    if (isRegister.value) {
      await auth.register(form.email.trim(), form.password, form.displayName.trim())
      if (!auth.isLoggedIn) {
        successMessage.value = 'Cuenta creada. Revisa tu correo para confirmar el acceso.'
        return
      }
    } else {
      await auth.login(form.email.trim(), form.password)
    }
    router.push('/today')
  } catch (error) {
    errorMessage.value = translateAuthError(error)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-root">
    <!-- ── Hero (solo desktop) ──────────────────────────────────────────── -->
    <aside class="hero-side">
      <span class="logo">DYDI</span>

      <div class="hero-body">
        <span class="badge">Accountability social sin ponerse solemnes</span>

        <h1 class="hero-heading">Cumple tus hábitos o enfréntate a la ruleta del grupo.</h1>

        <p class="hero-sub">
          Arma tu squad, registra tus check-ins diarios y deja que las consecuencias se vuelvan
          parte del juego.
        </p>

        <div class="stats">
          <div class="stat-card">
            <p class="stat-num terracotta">08</p>
            <p class="stat-label">personas por grupo</p>
          </div>
          <div class="stat-card">
            <p class="stat-num sage">24h</p>
            <p class="stat-label">para votar propuestas</p>
          </div>
          <div class="stat-card">
            <p class="stat-num amber">1</p>
            <p class="stat-label">ruleta semanal</p>
          </div>
        </div>
      </div>

      <p class="hero-footer">© 2025 Dydi · UTD Integradora</p>
    </aside>

    <!-- ── Form side ────────────────────────────────────────────────────── -->
    <main class="form-side">
      <!-- Logo solo en mobile -->
      <div class="mobile-logo">
        <span class="logo">DYDI</span>
        <p class="mobile-tagline">Hábitos con consecuencias</p>
      </div>

      <!-- Card -->
      <div class="card">
        <!-- Tab switcher -->
        <div class="tabs" role="tablist">
          <button
            type="button"
            role="tab"
            class="tab"
            :class="{ 'tab--active': !isRegister }"
            @click="mode = 'login'; resetFeedback()"
          >
            Entrar
          </button>
          <button
            type="button"
            role="tab"
            class="tab"
            :class="{ 'tab--active': isRegister }"
            @click="mode = 'register'; resetFeedback()"
          >
            Registro
          </button>
        </div>

        <!-- Header del form -->
        <div class="form-header">
          <h2 class="form-title">{{ title }}</h2>
          <p class="form-sub">
            {{
              isRegister
                ? 'Únete a tu grupo y empieza el reto.'
                : 'Vuelve con tu squad y marca el día.'
            }}
          </p>
        </div>

        <!-- Form -->
        <form class="form" @submit.prevent="submit">
          <div v-if="isRegister" class="field">
            <label for="displayName" class="field-label">Nombre</label>
            <input
              id="displayName"
              v-model="form.displayName"
              type="text"
              autocomplete="name"
              required
              class="field-input"
              placeholder="Tu nombre o apodo"
            />
          </div>

          <div class="field">
            <label for="email" class="field-label">Correo</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              class="field-input"
              placeholder="tu@correo.com"
            />
          </div>

          <div class="field">
            <label for="password" class="field-label">Contraseña</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              minlength="6"
              required
              class="field-input"
              placeholder="Mínimo 6 caracteres"
            />
          </div>

          <div v-if="isRegister" class="field">
            <label for="confirmPassword" class="field-label">Confirmar contraseña</label>
            <input
              id="confirmPassword"
              v-model="form.confirmPassword"
              type="password"
              autocomplete="new-password"
              minlength="6"
              required
              class="field-input"
              placeholder="Repítela una vez"
            />
          </div>

          <p v-if="errorMessage" class="feedback feedback--error" role="alert">
            {{ errorMessage }}
          </p>
          <p v-if="successMessage" class="feedback feedback--success" role="status">
            {{ successMessage }}
          </p>

          <button type="submit" class="submit-btn" :disabled="loading">
            {{ submitLabel }}
          </button>
        </form>

        <!-- Switch mode -->
        <p class="switch-mode">
          {{ isRegister ? '¿Ya tienes cuenta?' : '¿Eres nuevo?' }}
          <button type="button" class="switch-link" @click="switchMode">
            {{ toggleLabel }}
          </button>
        </p>
      </div>
    </main>
  </div>
</template>

<style scoped>
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.login-root {
  display: flex;
  min-height: 100vh;
  background: #f4eee3;
  font-family: 'Hanken Grotesk', system-ui, sans-serif;
  color: #2a251f;
}

/* ── Hero side ──────────────────────────────────────────────────────── */
.hero-side {
  display: none;
}

@media (min-width: 1024px) {
  .hero-side {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    width: 55%;
    padding: 3rem 3.5rem;
    animation: fadeIn 0.6s ease both;
  }
}

.logo {
  font-family: 'Newsreader', Georgia, serif;
  font-size: 1.5rem;
  font-weight: 600;
  color: #c26f4d;
  letter-spacing: 0.05em;
}

.hero-body {
  animation: fadeUp 0.6s 0.1s ease both;
}

.badge {
  display: inline-flex;
  align-items: center;
  background: #dfebe8;
  color: #7ca39d;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 0.4rem 0.9rem;
  border-radius: 999px;
  margin-bottom: 1.75rem;
}

.hero-heading {
  font-family: 'Newsreader', Georgia, serif;
  font-size: clamp(2.2rem, 3.5vw, 3.25rem);
  font-weight: 600;
  line-height: 1.15;
  color: #2a251f;
  margin-bottom: 1.25rem;
  max-width: 22ch;
}

.hero-sub {
  font-size: 1rem;
  line-height: 1.7;
  color: #6f6557;
  max-width: 38ch;
  margin-bottom: 2.5rem;
}

.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  animation: fadeUp 0.6s 0.25s ease both;
}

.stat-card {
  background: #fcf9f3;
  border-radius: 22px;
  padding: 1.1rem;
}

.stat-num {
  font-family: 'Newsreader', Georgia, serif;
  font-size: 2rem;
  font-weight: 600;
  line-height: 1;
}
.stat-num.terracotta {
  color: #c26f4d;
}
.stat-num.sage {
  color: #7ca39d;
}
.stat-num.amber {
  color: #e9c281;
}

.stat-label {
  font-size: 0.75rem;
  color: #6f6557;
  margin-top: 0.35rem;
}

.hero-footer {
  font-size: 0.7rem;
  color: #a89c89;
}

/* ── Form side ──────────────────────────────────────────────────────── */
.form-side {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem 1.25rem;
  min-height: 100vh;
}

@media (min-width: 1024px) {
  .form-side {
    padding: 2.5rem;
    min-height: auto;
  }
}

.mobile-logo {
  text-align: center;
  margin-bottom: 2rem;
  animation: fadeUp 0.5s ease both;
}
@media (min-width: 1024px) {
  .mobile-logo {
    display: none;
  }
}

.mobile-tagline {
  font-size: 0.8rem;
  color: #6f6557;
  margin-top: 0.25rem;
}

/* ── Card ───────────────────────────────────────────────────────────── */
.card {
  width: 100%;
  max-width: 22rem;
  background: #ffffff;
  border-radius: 22px;
  box-shadow: 0 4px 32px 0 rgba(42, 37, 31, 0.1);
  padding: 1.75rem 1.5rem;
  animation: fadeUp 0.5s 0.15s ease both;
}

/* ── Tabs ───────────────────────────────────────────────────────────── */
.tabs {
  display: flex;
  background: #e7decd;
  border-radius: 999px;
  padding: 0.2rem;
  margin-bottom: 1.5rem;
}

.tab {
  flex: 1;
  padding: 0.6rem;
  border-radius: 999px;
  font-size: 0.875rem;
  font-weight: 700;
  color: #6f6557;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab--active {
  background: #ffffff;
  color: #2a251f;
  box-shadow: 0 1px 4px 0 rgba(42, 37, 31, 0.1);
}

/* ── Form header ────────────────────────────────────────────────────── */
.form-header {
  margin-bottom: 1.25rem;
}

.form-title {
  font-family: 'Newsreader', Georgia, serif;
  font-size: 1.5rem;
  font-weight: 600;
  color: #2a251f;
  line-height: 1.2;
}

.form-sub {
  font-size: 0.8125rem;
  color: #6f6557;
  margin-top: 0.25rem;
}

/* ── Fields ─────────────────────────────────────────────────────────── */
.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.field-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #2a251f;
}

.field-input {
  width: 100%;
  background: #fcf9f3;
  border: 1.5px solid #c8bca8;
  border-radius: 12px;
  padding: 0.75rem 1rem;
  font-family: 'Hanken Grotesk', system-ui, sans-serif;
  font-size: 0.9375rem;
  color: #2a251f;
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
  -webkit-appearance: none;
  box-sizing: border-box;
}

.field-input::placeholder {
  color: #a89c89;
}

.field-input:focus {
  border-color: #7ca39d;
  box-shadow: 0 0 0 3px rgba(124, 163, 157, 0.18);
  background: #ffffff;
}

/* ── Feedback ───────────────────────────────────────────────────────── */
.feedback {
  font-size: 0.8125rem;
  font-weight: 500;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  animation: fadeUp 0.3s ease both;
}

.feedback--error {
  background: rgba(237, 164, 143, 0.15);
  border: 1px solid rgba(237, 164, 143, 0.45);
  color: #b85a3d;
}

.feedback--success {
  background: rgba(168, 195, 154, 0.15);
  border: 1px solid rgba(168, 195, 154, 0.45);
  color: #5a8a6f;
}

/* ── Submit ─────────────────────────────────────────────────────────── */
.submit-btn {
  width: 100%;
  background: #7ca39d;
  color: #ffffff;
  border: none;
  border-radius: 999px;
  padding: 0.9rem;
  font-family: 'Hanken Grotesk', system-ui, sans-serif;
  font-size: 0.9375rem;
  font-weight: 700;
  cursor: pointer;
  transition:
    opacity 0.2s ease,
    transform 0.15s ease;
  margin-top: 0.25rem;
}

.submit-btn:hover:not(:disabled) {
  opacity: 0.88;
}
.submit-btn:active:not(:disabled) {
  transform: scale(0.97);
}
.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Switch mode ────────────────────────────────────────────────────── */
.switch-mode {
  text-align: center;
  font-size: 0.8125rem;
  color: #6f6557;
  margin-top: 1rem;
}

.switch-link {
  font-weight: 700;
  color: #7ca39d;
  background: none;
  border: none;
  cursor: pointer;
  margin-left: 0.2rem;
  transition: color 0.2s ease;
  padding: 0;
  font-size: inherit;
  font-family: inherit;
}

.switch-link:hover {
  color: #c26f4d;
}
</style>
