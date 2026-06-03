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
const toggleLabel = computed(() =>
  isRegister.value ? 'Ya tengo cuenta' : 'Crear cuenta nueva'
)

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
  const message = error?.message?.toLowerCase?.() ?? ''

  if (message.includes('invalid login credentials')) {
    return 'El correo o la contraseña no coinciden.'
  }
  if (message.includes('email')) {
    return 'Revisa que el correo esté bien escrito.'
  }
  if (message.includes('password')) {
    return 'La contraseña debe tener al menos 6 caracteres.'
  }

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
  <div class="min-h-screen bg-[#120F0D] text-[#FFF7ED]">
    <div class="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-5 py-6 sm:px-8 lg:px-10">
      <header class="flex items-center justify-between">
        <div>
          <p class="font-display text-sm font-semibold uppercase tracking-[0.18em] text-[#F59E0B]">
            Dydi
          </p>
          <h1 class="font-display text-3xl font-bold leading-tight sm:text-4xl">
            Hábitos con compas.
          </h1>
        </div>
      </header>

      <main class="grid flex-1 items-center gap-8 py-8 lg:grid-cols-[1fr_420px]">
        <section class="max-w-2xl">
          <div class="mb-8 inline-flex rounded-full border border-[#F59E0B]/30 bg-[#F59E0B]/10 px-4 py-2 text-sm font-semibold text-[#FDE68A]">
            Accountability social sin ponerse solemnes
          </div>

          <h2 class="font-display text-4xl font-bold leading-tight sm:text-5xl lg:text-6xl">
            Cumple tus hábitos o enfréntate a la ruleta del grupo.
          </h2>

          <p class="mt-5 max-w-xl text-base leading-7 text-[#E7CFC2] sm:text-lg">
            Arma tu squad, registra tus check-ins diarios y deja que las consecuencias se vuelvan parte del juego.
          </p>

          <div class="mt-8 grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-white/10 bg-[#211916] p-4">
              <p class="text-2xl font-bold text-[#14B8A6]">08</p>
              <p class="mt-1 text-sm text-[#C4A99B]">personas por grupo</p>
            </div>
            <div class="rounded-lg border border-white/10 bg-[#211916] p-4">
              <p class="text-2xl font-bold text-[#22C55E]">24h</p>
              <p class="mt-1 text-sm text-[#C4A99B]">para votar propuestas</p>
            </div>
            <div class="rounded-lg border border-white/10 bg-[#211916] p-4">
              <p class="text-2xl font-bold text-[#F59E0B]">1</p>
              <p class="mt-1 text-sm text-[#C4A99B]">ruleta semanal</p>
            </div>
          </div>
        </section>

        <section class="rounded-lg border border-white/10 bg-[#211916] p-5 shadow-2xl shadow-black/30 sm:p-6">
          <div class="mb-6 flex rounded-lg bg-[#2B211D] p-1" role="tablist" aria-label="Acceso">
            <button
              type="button"
              class="min-h-11 flex-1 rounded-md px-4 text-sm font-semibold transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-[#14B8A6] focus:ring-offset-2 focus:ring-offset-[#211916] cursor-pointer"
              :class="!isRegister ? 'bg-[#F05D4F] text-white' : 'text-[#C4A99B] hover:text-white'"
              :aria-selected="!isRegister"
              role="tab"
              @click="mode = 'login'; resetFeedback()"
            >
              Entrar
            </button>
            <button
              type="button"
              class="min-h-11 flex-1 rounded-md px-4 text-sm font-semibold transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-[#14B8A6] focus:ring-offset-2 focus:ring-offset-[#211916] cursor-pointer"
              :class="isRegister ? 'bg-[#F05D4F] text-white' : 'text-[#C4A99B] hover:text-white'"
              :aria-selected="isRegister"
              role="tab"
              @click="mode = 'register'; resetFeedback()"
            >
              Registro
            </button>
          </div>

          <div class="mb-6">
            <h2 class="font-display text-2xl font-bold">{{ title }}</h2>
            <p class="mt-2 text-sm leading-6 text-[#C4A99B]">
              {{ isRegister ? 'Únete a tu grupo y empieza el reto.' : 'Vuelve con tu squad y marca el día.' }}
            </p>
          </div>

          <form class="space-y-4" @submit.prevent="submit">
            <div v-if="isRegister">
              <label for="displayName" class="mb-2 block text-sm font-semibold text-[#F7E7DA]">
                Nombre
              </label>
              <input
                id="displayName"
                v-model="form.displayName"
                type="text"
                autocomplete="name"
                required
                class="min-h-12 w-full rounded-lg border border-white/10 bg-[#120F0D] px-4 text-base text-white outline-none transition-colors duration-200 placeholder:text-[#8B6F62] focus:border-[#14B8A6] focus:ring-2 focus:ring-[#14B8A6]/30"
                placeholder="Tu nombre o apodo"
              >
            </div>

            <div>
              <label for="email" class="mb-2 block text-sm font-semibold text-[#F7E7DA]">
                Correo
              </label>
              <input
                id="email"
                v-model="form.email"
                type="email"
                autocomplete="email"
                required
                class="min-h-12 w-full rounded-lg border border-white/10 bg-[#120F0D] px-4 text-base text-white outline-none transition-colors duration-200 placeholder:text-[#8B6F62] focus:border-[#14B8A6] focus:ring-2 focus:ring-[#14B8A6]/30"
                placeholder="tu@correo.com"
              >
            </div>

            <div>
              <label for="password" class="mb-2 block text-sm font-semibold text-[#F7E7DA]">
                Contraseña
              </label>
              <input
                id="password"
                v-model="form.password"
                type="password"
                autocomplete="current-password"
                minlength="6"
                required
                class="min-h-12 w-full rounded-lg border border-white/10 bg-[#120F0D] px-4 text-base text-white outline-none transition-colors duration-200 placeholder:text-[#8B6F62] focus:border-[#14B8A6] focus:ring-2 focus:ring-[#14B8A6]/30"
                placeholder="Mínimo 6 caracteres"
              >
            </div>

            <div v-if="isRegister">
              <label for="confirmPassword" class="mb-2 block text-sm font-semibold text-[#F7E7DA]">
                Confirmar contraseña
              </label>
              <input
                id="confirmPassword"
                v-model="form.confirmPassword"
                type="password"
                autocomplete="new-password"
                minlength="6"
                required
                class="min-h-12 w-full rounded-lg border border-white/10 bg-[#120F0D] px-4 text-base text-white outline-none transition-colors duration-200 placeholder:text-[#8B6F62] focus:border-[#14B8A6] focus:ring-2 focus:ring-[#14B8A6]/30"
                placeholder="Repítela una vez"
              >
            </div>

            <p
              v-if="errorMessage"
              class="rounded-lg border border-[#FF4D4D]/40 bg-[#FF4D4D]/10 px-4 py-3 text-sm font-medium text-[#FCA5A5]"
              role="alert"
            >
              {{ errorMessage }}
            </p>

            <p
              v-if="successMessage"
              class="rounded-lg border border-[#22C55E]/40 bg-[#22C55E]/10 px-4 py-3 text-sm font-medium text-[#86EFAC]"
              role="status"
            >
              {{ successMessage }}
            </p>

            <button
              type="submit"
              class="min-h-12 w-full rounded-lg bg-[#F05D4F] px-5 text-base font-bold text-white transition-colors duration-200 hover:bg-[#E34E40] focus:outline-none focus:ring-2 focus:ring-[#14B8A6] focus:ring-offset-2 focus:ring-offset-[#211916] disabled:cursor-not-allowed disabled:opacity-60 cursor-pointer"
              :disabled="loading"
            >
              {{ submitLabel }}
            </button>
          </form>

          <button
            type="button"
            class="mt-5 min-h-11 w-full rounded-lg border border-white/10 px-4 text-sm font-semibold text-[#F7E7DA] transition-colors duration-200 hover:border-[#14B8A6]/50 hover:text-white focus:outline-none focus:ring-2 focus:ring-[#14B8A6] focus:ring-offset-2 focus:ring-offset-[#211916] cursor-pointer"
            @click="switchMode"
          >
            {{ toggleLabel }}
          </button>
        </section>
      </main>
    </div>
  </div>
</template>
