<script setup>
import { computed } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import ServerWakeup from '@/components/ui/ServerWakeup.vue'
import ToastHost from '@/components/ui/ToastHost.vue'
import BrandWordmark from '@/components/ui/BrandWordmark.vue'
import GroupSwitcher from '@/components/GroupSwitcher.vue'

const route = useRoute()
const isPublic = computed(
  () => route.meta.public || route.meta.checkinFlow || route.meta.onboarding
)

const tabs = [
  {
    path: '/today',
    label: 'Hoy',
    icon: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8"
      d="m2.25 12 8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125
      c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25
      c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504
      1.125-1.125V9.75"/>`,
  },
  {
    path: '/squad',
    label: 'Squad',
    icon: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8"
      d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94
      3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21
      c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971
      5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995
      0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197
      a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25
      2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0
      2.25 2.25 0 0 1 4.5 0Z"/>`,
  },
  {
    path: '/propuestas',
    label: 'Votar',
    icon: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8"
      d="M10.05 4.575a1.575 1.575 0 1 0-3.15 0v3m3.15-3v-1.5a1.575 1.575 0 0 1 3.15 0v1.5
      m-3.15 0 .075 5.925m3.075.75V4.575m0 0a1.575 1.575 0 0 1 3.15 0V15M6.9 7.575a1.575
      1.575 0 1 0-3.15 0v8.175a6.75 6.75 0 0 0 13.5 0v-5.1m-6.45-8.4"/>`,
  },
  {
    path: '/ruleta',
    label: 'Ruleta',
    icon: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8"
      d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591
      M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636
      5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z"/>`,
  },
  {
    path: '/tu',
    label: 'Cuenta',
    icon: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8"
      d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501
      20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12
      21.75c-2.676 0-5.216-.584-7.499-1.632Z"/>`,
  },
]

function isActive(path) {
  return route.path === path || route.path.startsWith(path + '/')
}

// Móvil: deja espacio para el bottom-nav. Desktop: deja espacio para el sidebar.
const mainClass = computed(() => (isPublic.value ? '' : 'pb-20 lg:pb-0 lg:pl-64'))
</script>

<template>
  <div class="min-h-screen bg-cream">
    <ServerWakeup />
    <ToastHost />

    <!-- ── Sidebar (escritorio) ──────────────────────────────────────────── -->
    <aside
      v-if="!isPublic"
      class="hidden lg:flex lg:flex-col lg:fixed lg:inset-y-0 lg:left-0 lg:w-64 bg-paper border-r border-hairline z-40"
    >
      <div class="px-6 py-6">
        <RouterLink to="/today" aria-label="Dydi — inicio">
          <BrandWordmark size="md" />
        </RouterLink>
      </div>

      <div class="px-4 pb-4">
        <GroupSwitcher />
      </div>

      <nav class="flex-1 px-3 space-y-1">
        <RouterLink
          v-for="tab in tabs"
          :key="tab.path"
          :to="tab.path"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage-deep/50"
          :class="isActive(tab.path) ? 'bg-wash text-sage-deep' : 'text-ink-soft hover:bg-cream-2'"
        >
          <!-- eslint-disable vue/no-v-html -- íconos SVG estáticos y de confianza (sin datos de usuario) -->
          <svg
            class="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
            v-html="tab.icon"
          />
          <!-- eslint-enable vue/no-v-html -->
          {{ tab.label }}
        </RouterLink>
      </nav>

      <p class="px-6 py-4 text-[0.7rem] text-ink-faint">© 2026 DYDI · UTD</p>
    </aside>

    <!-- ── Contenido ─────────────────────────────────────────────────────── -->
    <main :class="mainClass">
      <RouterView />
    </main>

    <!-- ── Bottom nav (móvil) ────────────────────────────────────────────── -->
    <nav
      v-if="!isPublic"
      class="lg:hidden fixed bottom-0 inset-x-0 bg-paper border-t border-hairline z-50 safe-area-bottom"
    >
      <div class="flex items-center max-w-md mx-auto">
        <RouterLink
          v-for="tab in tabs"
          :key="tab.path"
          :to="tab.path"
          class="flex-1 flex flex-col items-center py-2.5 gap-0.5 transition-colors"
          :class="isActive(tab.path) ? 'text-sage-deep' : 'text-ink-faint'"
        >
          <!-- eslint-disable vue/no-v-html -- íconos SVG estáticos y de confianza (sin datos de usuario) -->
          <svg
            class="w-6 h-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
            v-html="tab.icon"
          />
          <!-- eslint-enable vue/no-v-html -->
          <span class="text-[10px] font-semibold tracking-wide">{{ tab.label }}</span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>
