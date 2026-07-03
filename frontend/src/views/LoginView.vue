<script setup>
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import BaseButton from '@/components/ui/BaseButton.vue'
import BrandWordmark from '@/components/ui/BrandWordmark.vue'

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

const showPassword = ref(false)

const strengthCriteria = computed(() => {
  const p = form.password
  return {
    length: p.length >= 6,
    uppercase: /[A-Z]/.test(p),
    lowercase: /[a-z]/.test(p),
    number: /[0-9]/.test(p),
  }
})

const isPasswordStrong = computed(() => {
  return (
    strengthCriteria.value.length &&
    strengthCriteria.value.uppercase &&
    strengthCriteria.value.lowercase &&
    strengthCriteria.value.number
  )
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

function setMode(next) {
  mode.value = next
  resetFeedback()
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
  if (isRegister.value) {
    if (!isPasswordStrong.value) {
      errorMessage.value = 'La contraseña no cumple con los requisitos de seguridad.'
      return
    }
    if (form.password !== form.confirmPassword) {
      errorMessage.value = 'Las contraseñas no coinciden.'
      return
    }
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

const fieldInput =
  'w-full bg-surface border border-hairline rounded-xl px-4 py-3 text-[0.9375rem] text-ink ' +
  'placeholder-ink-faint outline-none transition-colors ' +
  'focus:border-sage-deep focus:bg-paper focus:ring-[3px] focus:ring-sage-deep/20'
</script>

<template>
  <div class="flex min-h-screen bg-cream text-ink">
    <!-- ── Hero (solo desktop) ──────────────────────────────────────────── -->
    <aside
      class="hidden lg:flex lg:w-[55%] lg:flex-col lg:justify-between lg:px-14 lg:py-12 animate-fade-in"
    >
      <BrandWordmark size="lg" />

      <div class="animate-fade-up [animation-delay:100ms]">
        <span
          class="inline-flex items-center bg-wash text-sage-deep text-[11px] font-bold uppercase tracking-eyebrow px-3.5 py-1.5 rounded-pill mb-7"
        >
          Accountability social sin ponerse solemnes
        </span>

        <h1
          class="font-serif font-semibold text-ink leading-[1.15] mb-5 max-w-[22ch] text-[clamp(2.2rem,3.5vw,3.25rem)]"
        >
          Cumple tus hábitos o enfréntate a la ruleta del grupo.
        </h1>

        <p class="text-base leading-[1.7] text-ink-soft max-w-[38ch] mb-10">
          Arma tu squad, registra tus check-ins diarios y deja que las consecuencias se vuelvan
          parte del juego.
        </p>

        <div class="grid grid-cols-3 gap-3 animate-fade-up [animation-delay:250ms]">
          <div class="bg-surface rounded-card p-4">
            <p class="font-serif text-3xl font-semibold leading-none text-terracotta">08</p>
            <p class="text-xs text-ink-soft mt-1.5">personas por grupo</p>
          </div>
          <div class="bg-surface rounded-card p-4">
            <p class="font-serif text-3xl font-semibold leading-none text-sage-deep">24h</p>
            <p class="text-xs text-ink-soft mt-1.5">para votar propuestas</p>
          </div>
          <div class="bg-surface rounded-card p-4">
            <p class="font-serif text-3xl font-semibold leading-none text-amber-deep">1</p>
            <p class="text-xs text-ink-soft mt-1.5">ruleta semanal</p>
          </div>
        </div>
      </div>

      <p class="text-[0.7rem] text-ink-faint">© 2026 DYDI · UTD Integradora</p>
    </aside>

    <!-- ── Form side ────────────────────────────────────────────────────── -->
    <main
      class="flex-1 flex flex-col items-center justify-center px-5 py-6 min-h-screen lg:p-10 lg:min-h-0"
    >
      <!-- Logo solo en mobile -->
      <div class="text-center mb-8 animate-fade-up lg:hidden">
        <BrandWordmark size="lg" />
        <p class="text-sm text-ink-soft mt-1">Hábitos con consecuencias</p>
      </div>

      <!-- Card -->
      <div
        class="w-full max-w-[22rem] bg-paper rounded-card shadow-card px-6 py-7 animate-fade-up [animation-delay:150ms]"
      >
        <!-- Tab switcher -->
        <div class="flex bg-hairline rounded-pill p-1 mb-6" role="tablist">
          <button
            type="button"
            role="tab"
            :aria-selected="!isRegister"
            class="flex-1 py-2.5 rounded-pill text-sm font-bold transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage-deep/50"
            :class="!isRegister ? 'bg-paper text-ink shadow-flat' : 'text-ink-soft'"
            @click="setMode('login')"
          >
            Entrar
          </button>
          <button
            type="button"
            role="tab"
            :aria-selected="isRegister"
            class="flex-1 py-2.5 rounded-pill text-sm font-bold transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage-deep/50"
            :class="isRegister ? 'bg-paper text-ink shadow-flat' : 'text-ink-soft'"
            @click="setMode('register')"
          >
            Registro
          </button>
        </div>

        <!-- Header del form -->
        <div class="mb-5">
          <h2 class="font-serif text-2xl font-semibold text-ink leading-tight">{{ title }}</h2>
          <p class="text-[0.8125rem] text-ink-soft mt-1">
            {{
              isRegister
                ? 'Únete a tu grupo y empieza el reto.'
                : 'Vuelve con tu squad y marca el día.'
            }}
          </p>
        </div>

        <!-- Form -->
        <form class="flex flex-col gap-4" @submit.prevent="submit">
          <div v-if="isRegister" class="flex flex-col gap-1.5">
            <label for="displayName" class="text-[0.8125rem] font-semibold text-ink">Nombre</label>
            <input
              id="displayName"
              v-model="form.displayName"
              type="text"
              autocomplete="name"
              required
              :class="fieldInput"
              placeholder="Tu nombre o apodo"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label for="email" class="text-[0.8125rem] font-semibold text-ink">Correo</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              :class="fieldInput"
              placeholder="tu@correo.com"
            />
          </div>

          <div class="flex flex-col gap-1.5">
            <label for="password" class="text-[0.8125rem] font-semibold text-ink">Contraseña</label>
            <div class="relative">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                minlength="6"
                required
                :class="fieldInput + ' pr-10'"
                placeholder="Mínimo 6 caracteres"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-ink-soft hover:text-sage-deep transition-colors focus:outline-none"
                :aria-label="showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'"
                @click="showPassword = !showPassword"
              >
                <!-- SVG Ocultar -->
                <svg
                  v-if="showPassword"
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                  <path
                    d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"
                  />
                  <path
                    d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"
                  />
                  <line x1="2" x2="22" y1="2" y2="22" />
                </svg>
                <!-- SVG Mostrar -->
                <svg
                  v-else
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              </button>
            </div>

            <!-- Indicadores de seguridad -->
            <ul
              v-if="isRegister && form.password"
              class="mt-1 flex flex-col gap-1.5 p-3 bg-surface rounded-xl border border-hairline animate-fade-in"
            >
              <li
                class="flex items-center gap-2 text-[0.75rem]"
                :class="strengthCriteria.length ? 'text-sage-deep font-medium' : 'text-ink-soft'"
              >
                <svg
                  v-if="strengthCriteria.length"
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <div v-else class="w-3.5 h-3.5 border border-current rounded-full opacity-40"></div>
                Mínimo 6 caracteres
              </li>
              <li
                class="flex items-center gap-2 text-[0.75rem]"
                :class="strengthCriteria.uppercase ? 'text-sage-deep font-medium' : 'text-ink-soft'"
              >
                <svg
                  v-if="strengthCriteria.uppercase"
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <div v-else class="w-3.5 h-3.5 border border-current rounded-full opacity-40"></div>
                Al menos 1 mayúscula
              </li>
              <li
                class="flex items-center gap-2 text-[0.75rem]"
                :class="strengthCriteria.lowercase ? 'text-sage-deep font-medium' : 'text-ink-soft'"
              >
                <svg
                  v-if="strengthCriteria.lowercase"
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <div v-else class="w-3.5 h-3.5 border border-current rounded-full opacity-40"></div>
                Al menos 1 minúscula
              </li>
              <li
                class="flex items-center gap-2 text-[0.75rem]"
                :class="strengthCriteria.number ? 'text-sage-deep font-medium' : 'text-ink-soft'"
              >
                <svg
                  v-if="strengthCriteria.number"
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <div v-else class="w-3.5 h-3.5 border border-current rounded-full opacity-40"></div>
                Al menos 1 número
              </li>
            </ul>
          </div>

          <div v-if="isRegister" class="flex flex-col gap-1.5">
            <label for="confirmPassword" class="text-[0.8125rem] font-semibold text-ink">
              Confirmar contraseña
            </label>
            <input
              id="confirmPassword"
              v-model="form.confirmPassword"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="new-password"
              minlength="6"
              required
              :class="fieldInput"
              placeholder="Repítela una vez"
            />
          </div>

          <p
            v-if="errorMessage"
            class="text-[0.8125rem] font-medium px-4 py-3 rounded-xl bg-coral-soft border border-coral/40 text-coral-deep animate-fade-up"
            role="alert"
          >
            {{ errorMessage }}
          </p>
          <p
            v-if="successMessage"
            class="text-[0.8125rem] font-medium px-4 py-3 rounded-xl bg-sage-soft border border-sage/40 text-sage-deep animate-fade-up"
            role="status"
          >
            {{ successMessage }}
          </p>

          <BaseButton type="submit" block size="lg" :loading="loading" class="mt-1">
            {{ submitLabel }}
          </BaseButton>
        </form>

        <!-- Switch mode -->
        <p class="text-center text-[0.8125rem] text-ink-soft mt-4">
          {{ isRegister ? '¿Ya tienes cuenta?' : '¿Eres nuevo?' }}
          <button
            type="button"
            class="font-bold text-sage-deep hover:text-terracotta transition-colors ml-0.5 focus-visible:outline-none focus-visible:underline"
            @click="switchMode"
          >
            {{ toggleLabel }}
          </button>
        </p>
      </div>
    </main>
  </div>
</template>
