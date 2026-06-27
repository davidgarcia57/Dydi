<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { useHabitsStore } from '@/stores/habits'
import PageContainer from '@/components/ui/PageContainer.vue'

const router = useRouter()
const auth = useAuthStore()
const group = useGroupStore()
const habits = useHabitsStore()

const loggingOut = ref(false)
const leavingGroup = ref(false)
const confirmLeave = ref(false)
const loadError = ref(false)
const copiedInvite = ref(false)
const profileSaving = ref(false)
const passwordSaving = ref(false)
const feedback = ref({ type: '', message: '' })

const profileForm = reactive({
  displayName: '',
})

const passwordForm = reactive({
  password: '',
  confirmPassword: '',
})

const displayName = computed(
  () => auth.user?.user_metadata?.display_name ?? auth.user?.email?.split('@')[0] ?? 'Tu cuenta'
)

const email = computed(() => auth.user?.email ?? '')
const myStreak = computed(() => habits.streaks[auth.user?.id] ?? 0)
const inviteCode = computed(() =>
  group.group?.id && group.group?.invite_code ? group.group.id + ':' + group.group.invite_code : ''
)
const isProfileDirty = computed(() => profileForm.displayName.trim() !== displayName.value)

const COLORS = ['bg-sage-deep', 'bg-terracotta', 'bg-sage', 'bg-amber', 'bg-coral']
const initials = (name = '') =>
  name
    .trim()
    .split(/\s+/)
    .map((word) => word[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
const avatarBg = (name = '') => COLORS[(name?.charCodeAt(0) ?? 0) % COLORS.length]

function setFeedback(type, message) {
  feedback.value = { type, message }
}

function resetFeedback() {
  feedback.value = { type: '', message: '' }
}

function syncProfileForm() {
  profileForm.displayName = displayName.value
}

async function saveProfile() {
  const nextName = profileForm.displayName.trim()
  if (!nextName || profileSaving.value) return

  profileSaving.value = true
  resetFeedback()
  try {
    await auth.updateProfile(nextName)
    await api('/api/users/sync', {
      method: 'POST',
      body: JSON.stringify({ display_name: nextName, avatar_url: null }),
    })
    syncProfileForm()
    setFeedback('success', 'Tu nombre se actualizó.')
  } catch (error) {
    setFeedback('error', error?.error ?? error?.message ?? 'No pudimos actualizar tu perfil.')
  } finally {
    profileSaving.value = false
  }
}

async function changePassword() {
  resetFeedback()
  if (passwordForm.password.length < 6) {
    setFeedback('error', 'La contraseña debe tener al menos 6 caracteres.')
    return
  }
  if (passwordForm.password !== passwordForm.confirmPassword) {
    setFeedback('error', 'Las contraseñas no coinciden.')
    return
  }

  passwordSaving.value = true
  try {
    await auth.changePassword(passwordForm.password)
    passwordForm.password = ''
    passwordForm.confirmPassword = ''
    setFeedback('success', 'Tu contraseña se actualizó.')
  } catch (error) {
    setFeedback('error', error?.message ?? 'No pudimos cambiar tu contraseña.')
  } finally {
    passwordSaving.value = false
  }
}

async function copyInviteCode() {
  if (!inviteCode.value) return
  await navigator.clipboard?.writeText(inviteCode.value)
  copiedInvite.value = true
  setTimeout(() => {
    copiedInvite.value = false
  }, 2000)
}

async function shareInvite() {
  if (!inviteCode.value) return
  const text = 'Únete a mi squad "' + group.group.name + '" en Dydi. Código: ' + inviteCode.value
  if (navigator.share) {
    try {
      await navigator.share({ title: 'Únete a Dydi', text })
      return
    } catch {}
  }
  await copyInviteCode()
}

async function handleLogout() {
  loggingOut.value = true
  await auth.logout()
  group.reset()
  router.replace('/login')
}

async function handleLeaveGroup() {
  leavingGroup.value = true
  resetFeedback()
  try {
    await group.leaveGroup()
    router.replace('/onboarding')
  } catch (error) {
    confirmLeave.value = false
    setFeedback('error', error?.error ?? error?.message ?? 'No pudimos sacarte del grupo.')
  } finally {
    leavingGroup.value = false
  }
}

async function load() {
  loadError.value = false
  syncProfileForm()
  try {
    await group.autoLoad()
    if (group.group?.id && auth.user?.id) {
      await habits.loadStreaks(auth.user.id)
    }
  } catch (_) {
    loadError.value = true
  }
}

onMounted(load)
</script>

<template>
  <PageContainer>
    <header class="mb-6 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-eyebrow mb-2">CUENTA</p>
        <h1 class="font-serif text-3xl font-semibold text-ink leading-tight">Tu espacio</h1>
        <p class="text-sm text-ink-soft mt-1">Perfil, seguridad y grupo en un solo lugar.</p>
      </div>
      <button
        :disabled="loggingOut"
        class="rounded-pill border border-hairline bg-paper px-4 py-2 text-sm font-bold text-ink-soft transition-opacity active:opacity-70 disabled:opacity-40"
        @click="handleLogout"
      >
        {{ loggingOut ? 'Cerrando...' : 'Cerrar sesión' }}
      </button>
    </header>

    <div
      v-if="loadError"
      class="rounded-card bg-coral-soft/40 border border-coral/40 p-4 mb-4 flex flex-wrap items-center justify-between gap-3"
    >
      <p class="text-sm font-medium text-coral-deep">
        No pudimos cargar todos tus datos. Puedes editar tu cuenta y reintentar lo del grupo.
      </p>
      <button
        class="rounded-pill bg-coral text-paper px-4 py-2 text-sm font-bold active:opacity-80 transition-opacity"
        @click="load"
      >
        Reintentar
      </button>
    </div>

    <p
      v-if="feedback.message"
      class="mb-4 rounded-xl px-4 py-3 text-sm font-semibold"
      :class="
        feedback.type === 'success'
          ? 'bg-sage-soft text-sage-deep border border-sage/40'
          : 'bg-coral-soft text-coral-deep border border-coral/40'
      "
      role="status"
    >
      {{ feedback.message }}
    </p>

    <section
      class="rounded-card shadow-card bg-paper p-5 mb-5 flex flex-col gap-4 sm:flex-row sm:items-center"
    >
      <div
        class="w-16 h-16 rounded-full flex items-center justify-center text-paper text-xl font-bold flex-shrink-0"
        :class="avatarBg(displayName)"
      >
        {{ initials(displayName) }}
      </div>
      <div class="flex-1 min-w-0">
        <h2 class="font-serif text-2xl font-semibold text-ink leading-tight truncate">
          {{ displayName }}
        </h2>
        <p class="text-sm text-ink-soft truncate">{{ email }}</p>
        <p class="text-xs text-ink-faint mt-1">{{ group.group?.name ?? 'Sin grupo activo' }}</p>
      </div>
      <div class="rounded-card bg-surface px-5 py-3 text-center sm:min-w-28">
        <p class="font-serif text-4xl font-semibold text-terracotta leading-none">{{ myStreak }}</p>
        <p class="text-eyebrow text-terracotta mt-1">RACHA</p>
      </div>
    </section>

    <div class="grid grid-cols-1 xl:grid-cols-[minmax(0,1fr)_22rem] gap-5 items-start">
      <div class="space-y-5">
        <section class="rounded-card shadow-flat bg-paper p-5">
          <div class="mb-4">
            <h2 class="font-serif text-xl font-semibold text-ink">Perfil</h2>
            <p class="text-sm text-ink-soft mt-1">
              Este nombre aparece en tu squad y en tus check-ins.
            </p>
          </div>

          <form
            class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end"
            @submit.prevent="saveProfile"
          >
            <label class="block">
              <span class="text-xs font-bold text-ink-soft">Nombre o apodo</span>
              <input
                v-model="profileForm.displayName"
                type="text"
                maxlength="40"
                autocomplete="name"
                class="mt-1 w-full rounded-xl border border-hairline bg-surface px-4 py-3 text-sm text-ink outline-none transition-colors focus:border-sage-deep focus:bg-paper focus:ring-[3px] focus:ring-sage-deep/20"
              />
            </label>
            <button
              :disabled="!profileForm.displayName.trim() || !isProfileDirty || profileSaving"
              class="rounded-pill bg-sage-deep text-paper px-5 py-3 text-sm font-bold transition-opacity active:opacity-80 disabled:opacity-40"
            >
              {{ profileSaving ? 'Guardando...' : 'Guardar' }}
            </button>
          </form>
        </section>

        <section class="rounded-card shadow-flat bg-paper p-5">
          <div class="mb-4">
            <h2 class="font-serif text-xl font-semibold text-ink">Seguridad</h2>
            <p class="text-sm text-ink-soft mt-1">Cambia tu contraseña sin salir de la app.</p>
          </div>

          <form class="grid gap-3 sm:grid-cols-2" @submit.prevent="changePassword">
            <label class="block">
              <span class="text-xs font-bold text-ink-soft">Nueva contraseña</span>
              <input
                v-model="passwordForm.password"
                type="password"
                minlength="6"
                autocomplete="new-password"
                class="mt-1 w-full rounded-xl border border-hairline bg-surface px-4 py-3 text-sm text-ink outline-none transition-colors focus:border-sage-deep focus:bg-paper focus:ring-[3px] focus:ring-sage-deep/20"
              />
            </label>
            <label class="block">
              <span class="text-xs font-bold text-ink-soft">Confirmar</span>
              <input
                v-model="passwordForm.confirmPassword"
                type="password"
                minlength="6"
                autocomplete="new-password"
                class="mt-1 w-full rounded-xl border border-hairline bg-surface px-4 py-3 text-sm text-ink outline-none transition-colors focus:border-sage-deep focus:bg-paper focus:ring-[3px] focus:ring-sage-deep/20"
              />
            </label>
            <button
              :disabled="!passwordForm.password || !passwordForm.confirmPassword || passwordSaving"
              class="sm:col-span-2 rounded-pill bg-ink text-paper px-5 py-3 text-sm font-bold transition-opacity active:opacity-80 disabled:opacity-40"
            >
              {{ passwordSaving ? 'Actualizando...' : 'Cambiar contraseña' }}
            </button>
          </form>
        </section>

        <section class="rounded-card shadow-flat bg-paper p-5">
          <div class="mb-4">
            <h2 class="font-serif text-xl font-semibold text-ink">Zona delicada</h2>
            <p class="text-sm text-ink-soft mt-1">
              Borrar cuenta requiere un endpoint seguro con service role; no debe hacerse desde el
              cliente.
            </p>
          </div>
          <button
            disabled
            class="w-full rounded-pill border border-coral/40 bg-coral-soft/30 text-coral-deep py-3 text-sm font-bold opacity-70"
          >
            Borrar cuenta pendiente de backend seguro
          </button>
        </section>
      </div>

      <aside class="space-y-5">
        <section class="rounded-card shadow-card bg-paper p-5">
          <div class="mb-4">
            <h2 class="font-serif text-xl font-semibold text-ink">Grupo</h2>
            <p class="text-sm text-ink-soft mt-1">Invitación y salida del squad.</p>
          </div>

          <div v-if="group.group" class="space-y-4">
            <div class="rounded-card bg-surface p-4">
              <p class="text-eyebrow mb-1">SQUAD ACTUAL</p>
              <p class="font-semibold text-ink">{{ group.group.name }}</p>
              <p class="font-mono text-xs text-ink-soft break-all mt-2">{{ inviteCode }}</p>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <button
                class="rounded-pill border border-hairline bg-surface text-ink-soft py-2.5 text-sm font-bold transition-colors active:opacity-80"
                :class="{ 'bg-sage-soft text-sage-deep border-sage/40': copiedInvite }"
                @click="copyInviteCode"
              >
                {{ copiedInvite ? 'Copiado' : 'Copiar' }}
              </button>
              <button
                class="rounded-pill bg-terracotta text-paper py-2.5 text-sm font-bold active:opacity-80 transition-opacity"
                @click="shareInvite"
              >
                Compartir
              </button>
            </div>

            <div v-if="!confirmLeave">
              <button
                class="w-full rounded-pill border border-hairline text-ink-soft py-3 font-semibold text-sm active:opacity-70 transition-opacity"
                @click="confirmLeave = true"
              >
                Salir del grupo
              </button>
            </div>
            <div v-else class="rounded-card border border-coral/40 bg-coral/5 p-4">
              <p class="text-sm font-semibold text-ink mb-1">
                ¿Seguro que quieres salir de
                <span class="text-coral-deep">{{ group.group.name }}</span
                >?
              </p>
              <p class="text-xs text-ink-soft mb-4">Perderás tus hábitos y rachas en este grupo.</p>
              <div class="flex gap-2">
                <button
                  :disabled="leavingGroup"
                  class="flex-1 rounded-pill bg-coral text-paper py-2.5 font-bold text-sm disabled:opacity-40 active:opacity-80 transition-opacity"
                  @click="handleLeaveGroup"
                >
                  {{ leavingGroup ? 'Saliendo...' : 'Sí, salir' }}
                </button>
                <button
                  class="flex-1 rounded-pill border border-hairline text-ink-soft py-2.5 font-semibold text-sm"
                  @click="confirmLeave = false"
                >
                  Cancelar
                </button>
              </div>
            </div>
          </div>

          <div v-else class="rounded-card bg-surface p-4 text-sm text-ink-soft">
            Todavía no tienes un grupo activo.
          </div>
        </section>
      </aside>
    </div>
  </PageContainer>
</template>
