<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useGroupStore } from '@/stores/group'
import BrandWordmark from '@/components/ui/BrandWordmark.vue'

const router = useRouter()
const group = useGroupStore()

// 'home' | 'create' | 'join' | 'created'
const step = ref('home')
const groupName = ref('')
const joinCode = ref('')
const loading = ref(false)
const errMsg = ref('')
const createdGroup = ref(null)

// Volver a la pantalla inicial limpiando el error. En método (no inline) porque
// un handler con dos sentencias se rompe al formatear con Prettier.
function goHome() {
  step.value = 'home'
  errMsg.value = ''
}

async function submitCreate() {
  if (!groupName.value.trim() || loading.value) return
  loading.value = true
  errMsg.value = ''
  try {
    createdGroup.value = await group.createGroup(groupName.value.trim())
    step.value = 'created'
  } catch (e) {
    errMsg.value = e?.error ?? e?.message ?? 'No se pudo crear el grupo.'
  } finally {
    loading.value = false
  }
}

async function submitJoin() {
  if (!joinCode.value.trim() || loading.value) return
  errMsg.value = ''

  // Expected format: "{groupID}:{inviteCode}"
  const parts = joinCode.value.trim().split(':')
  if (parts.length !== 2 || !parts[0] || !parts[1]) {
    errMsg.value = 'Formato inválido. Debe ser el código completo que te compartieron.'
    return
  }

  loading.value = true
  try {
    await group.joinGroup(parts[0], parts[1])
    router.replace('/today')
  } catch (e) {
    errMsg.value = e?.error ?? e?.message ?? 'Código inválido o grupo no encontrado.'
  } finally {
    loading.value = false
  }
}

function copyInviteCode() {
  if (!createdGroup.value) return
  const code = `${createdGroup.value.id}:${createdGroup.value.invite_code}`
  navigator.clipboard?.writeText(code)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}

const copied = ref(false)
</script>

<template>
  <div class="min-h-screen bg-cream flex flex-col items-center justify-center px-6 py-12">
    <!-- ── Home: elige acción ─────────────────────────────────────────────── -->
    <template v-if="step === 'home'">
      <div class="w-full max-w-sm text-center">
        <BrandWordmark size="xl" class="mb-4" />
        <h1 class="font-serif text-2xl font-semibold text-ink mb-2">Bienvenido</h1>
        <p class="text-sm text-ink-soft mb-10">Únete a tu squad o crea uno nuevo para empezar.</p>

        <div class="space-y-3">
          <button
            class="w-full rounded-pill bg-sage-deep text-paper py-4 font-bold text-sm active:opacity-80 transition-opacity"
            @click="step = 'create'"
          >
            Crear grupo →
          </button>
          <button
            class="w-full rounded-pill border border-hairline bg-paper text-ink py-4 font-bold text-sm active:opacity-80 transition-opacity"
            @click="step = 'join'"
          >
            Unirme con código
          </button>
        </div>
      </div>
    </template>

    <!-- ── Crear grupo ─────────────────────────────────────────────────────── -->
    <template v-else-if="step === 'create'">
      <div class="w-full max-w-sm">
        <button class="flex items-center gap-1 text-sm text-ink-soft mb-8" @click="goHome">
          ← Volver
        </button>

        <p class="text-eyebrow mb-2">NUEVO GRUPO</p>
        <h1 class="font-serif text-2xl font-semibold text-ink mb-8 leading-snug">
          ¿Cómo se llama tu squad?
        </h1>

        <label class="block mb-6">
          <input
            v-model="groupName"
            type="text"
            maxlength="40"
            placeholder="Los Incumplidos, El Squad, …"
            class="w-full rounded-[14px] border border-hairline bg-paper px-4 py-3.5 text-sm text-ink placeholder-ink-faint focus:outline-none focus:border-sage-deep transition-colors"
            @keyup.enter="submitCreate"
          />
        </label>

        <p v-if="errMsg" class="text-sm text-coral mb-4 font-medium">{{ errMsg }}</p>

        <button
          :disabled="!groupName.trim() || loading"
          class="w-full rounded-pill bg-sage-deep text-paper py-4 font-bold text-sm disabled:opacity-40 transition-opacity active:opacity-80"
          @click="submitCreate"
        >
          <span v-if="loading" class="flex items-center justify-center gap-2">
            <span
              class="w-4 h-4 rounded-full border-2 border-paper border-t-transparent animate-spin"
            />
            Creando…
          </span>
          <span v-else>Crear grupo →</span>
        </button>
      </div>
    </template>

    <!-- ── Grupo creado: muestra código de invitación ─────────────────────── -->
    <template v-else-if="step === 'created'">
      <div class="w-full max-w-sm text-center">
        <div
          class="w-16 h-16 rounded-full bg-sage/20 flex items-center justify-center mx-auto mb-6"
        >
          <svg class="w-8 h-8 text-sage-deep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>

        <p class="text-eyebrow text-sage-deep mb-1">GRUPO CREADO</p>
        <h1 class="font-serif text-2xl font-semibold text-ink mb-2">
          {{ createdGroup?.name }}
        </h1>
        <p class="text-sm text-ink-soft mb-8">
          Comparte este código con tu squad para que se unan.
        </p>

        <!-- Invite code box -->
        <div class="rounded-card bg-paper shadow-card px-5 py-4 mb-3 text-left">
          <p class="text-eyebrow mb-1">CÓDIGO DE INVITACIÓN</p>
          <p class="font-mono text-xs text-ink break-all leading-relaxed">
            {{ createdGroup?.id }}:{{ createdGroup?.invite_code }}
          </p>
        </div>

        <button
          class="w-full rounded-pill border border-hairline bg-surface text-ink-soft py-3 font-semibold text-sm mb-6 transition-colors"
          :class="{ 'bg-sage/10 text-sage-deep border-sage/30': copied }"
          @click="copyInviteCode"
        >
          {{ copied ? '¡Copiado! ✓' : 'Copiar código' }}
        </button>

        <button
          class="w-full rounded-pill bg-sage-deep text-paper py-4 font-bold text-sm active:opacity-80 transition-opacity"
          @click="router.replace('/today')"
        >
          Ir a Hoy →
        </button>
      </div>
    </template>

    <!-- ── Unirse con código ───────────────────────────────────────────────── -->
    <template v-else-if="step === 'join'">
      <div class="w-full max-w-sm">
        <button class="flex items-center gap-1 text-sm text-ink-soft mb-8" @click="goHome">
          ← Volver
        </button>

        <p class="text-eyebrow mb-2">UNIRSE A UN GRUPO</p>
        <h1 class="font-serif text-2xl font-semibold text-ink mb-2 leading-snug">
          Pega el código que te compartieron
        </h1>
        <p class="text-xs text-ink-soft mb-8">
          El código completo tiene el formato<br />
          <span class="font-mono">id-del-grupo:código-acceso</span>
        </p>

        <label class="block mb-6">
          <textarea
            v-model="joinCode"
            rows="3"
            placeholder="Pega aquí el código completo…"
            class="w-full rounded-[14px] border border-hairline bg-paper px-4 py-3.5 text-sm text-ink placeholder-ink-faint font-mono resize-none focus:outline-none focus:border-sage-deep transition-colors"
          />
        </label>

        <p v-if="errMsg" class="text-sm text-coral mb-4 font-medium">{{ errMsg }}</p>

        <button
          :disabled="!joinCode.trim() || loading"
          class="w-full rounded-pill bg-sage-deep text-paper py-4 font-bold text-sm disabled:opacity-40 transition-opacity active:opacity-80"
          @click="submitJoin"
        >
          <span v-if="loading" class="flex items-center justify-center gap-2">
            <span
              class="w-4 h-4 rounded-full border-2 border-paper border-t-transparent animate-spin"
            />
            Uniéndome…
          </span>
          <span v-else>Unirme al squad →</span>
        </button>
      </div>
    </template>
  </div>
</template>
