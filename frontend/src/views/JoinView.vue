<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupStore } from '@/stores/group'
import { showToast } from '@/composables/useToast'
import { savePendingInvite, clearPendingInvite } from '@/pendingInvite'
import BrandWordmark from '@/components/ui/BrandWordmark.vue'

// Invitación de un tap. La ruta es pública porque el caso que importa es que te
// inviten SIN tener cuenta: aquí se guarda la invitación, se manda al login, y el
// guard del router regresa a /join en cuanto haya sesión.
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const group = useGroupStore()

const error = ref('')

onMounted(async () => {
  const g = route.query.g
  const c = route.query.c

  if (!g || !c) {
    error.value = 'Este enlace de invitación está incompleto.'
    return
  }

  // Se guarda antes de cualquier redirección: si hay que pasar por el login, al
  // volver ya no está en la URL.
  savePendingInvite(g, c)

  if (!auth.isLoggedIn) {
    router.replace('/login')
    return
  }

  try {
    await group.joinGroup(g, c)
    clearPendingInvite()
    showToast('¡Ya estás dentro del squad!')
    router.replace('/today')
  } catch (e) {
    clearPendingInvite()
    // 409 = ya eres miembro. No es un error para el usuario: ya llegó donde iba.
    if (e?.status === 409) {
      router.replace('/today')
      return
    }
    error.value = e?.error ?? 'No pudimos usar esta invitación. Pídele al squad un código nuevo.'
  }
})
</script>

<template>
  <div class="min-h-screen bg-cream flex flex-col items-center justify-center px-6 text-center">
    <BrandWordmark size="md" class="mb-8" />

    <template v-if="error">
      <h1 class="font-serif text-2xl font-semibold text-ink mb-2">Invitación no válida</h1>
      <p class="text-sm text-ink-soft max-w-sm mb-6">{{ error }}</p>
      <RouterLink
        to="/today"
        class="rounded-pill bg-terracotta text-paper px-6 py-3 text-sm font-bold hover:opacity-90 transition-opacity"
      >
        Ir a Dydi
      </RouterLink>
    </template>

    <template v-else>
      <div
        class="w-12 h-12 rounded-full border-4 border-hairline border-t-sage-deep animate-spin mb-6"
      />
      <h1 class="font-serif text-2xl font-semibold text-ink mb-2">Entrando al squad…</h1>
      <p class="text-sm text-ink-soft">Un momento.</p>
    </template>
  </div>
</template>
