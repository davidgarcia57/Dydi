# CLAUDE.md — frontend (Dydi)

## Purpose
Vue 3 SPA deployed on Vercel. Mobile-first, responsive up to desktop (1280px+).
Communicates with the backend exclusively through `api-gateway`.
Real-time updates via WebSocket connection to `realtime-service` (through gateway).

---

## CRITICAL: Vue 3 Composition API only
**Always use `<script setup>` syntax.** Never use Options API or the old
`defineComponent()` pattern.

```vue
<!-- CORRECT -->
<script setup>
import { ref, computed } from 'vue'
const count = ref(0)
</script>

<!-- WRONG — do not use -->
<script>
export default {
  data() { return { count: 0 } }
}
</script>
```

---

## Stack

| Tool | Purpose | Notes |
|---|---|---|
| Vue 3 | UI framework | Composition API + `<script setup>` only |
| Vite | Build tool | Do not switch to other bundlers |
| Pinia | State management | One store per domain (auth, group, habits, penalties) |
| Tailwind CSS | Styling | Utility classes only, no custom CSS unless unavoidable |
| @vueuse/core | Utilities | Use `useWebSocket` for WebSocket, `useStorage` for local persistence |
| Vue Router | Routing | Hash mode for Vercel SPA compatibility |
| Vercel | Hosting | Auto-deploy from main branch |

---

## Design System (Dydi Identity)

**Paleta calida / clara. No dark mode. No cambiar sin consultar.**

### Colores — tokens en Tailwind y CSS variables
| Tailwind class | CSS var | Hex | Uso |
|---|---|---|---|
| `bg-cream` | `--color-cream` | `#F4EEE3` | Fondo principal de la app |
| `bg-surface` | `--color-surface` | `#FCF9F3` | Cards, modales, superficies |
| `bg-paper` | `--color-paper` | `#FFFFFF` | Input, superficies muy limpias |
| `border-hairline` | `--color-hairline` | `#E7DECD` | Bordes y divisores |
| `text-ink` | `--color-ink` | `#2A251F` | Texto principal |
| `text-ink-soft` | `--color-ink-soft` | `#6F6557` | Texto secundario |
| `text-ink-faint` | `--color-ink-faint` | `#A89C89` | Placeholders, deshabilitados |
| `bg-sage` / `text-sage` | `--color-sage` | `#A8C39A` | Estado: cumplio |
| `bg-amber` / `text-amber` | `--color-amber` | `#E9C281` | Estado: pendiente |
| `bg-coral` / `text-coral` | `--color-coral` | `#EDA48F` | Estado: fallo |
| `bg-sage-deep` / `text-sage-deep` | `--color-sage-deep` | `#7CA39D` | **CTA primario** (Hacer check-in) |
| `bg-terracotta` / `text-terracotta` | `--color-terracotta` | `#C26F4D` | **CTA secundario** (Girar ruleta), marca |
| `bg-wash` | `--color-wash` | `#DFEBE8` | Fondos de acento suave |

### Tipografia
| Fuente | Clase Tailwind | Uso |
|---|---|---|
| **Newsreader** (serif) | `font-serif` o `font-display` | Titulos display, numeros hero (racha, countdown, score) |
| **Hanken Grotesk** (sans) | `font-sans` (default) | Todo el UI: botones, etiquetas, cuerpo |

Pesos Hanken Grotesk:
- `font-bold` (700) — botones, nombres, labels clave
- `font-semibold` (600) — etiquetas y habitos
- `font-medium` (500) — cuerpo, descripciones

Eyebrow (clase utilitaria `.text-eyebrow`): `HANKEN GROTESK 700 · 11px · 0.1em tracking · UPPERCASE · color ink-soft`

### Cards y superficies
```
Card elevada -> rounded-card shadow-card bg-surface
Card plana   -> rounded-card shadow-flat bg-surface
Pill / tag   -> rounded-pill px-3 py-1 text-sm font-semibold
```

### Botones
```
Primario   -> bg-sage-deep text-paper rounded-pill px-6 py-3 font-bold
Secundario -> bg-terracotta text-paper rounded-pill px-6 py-3 font-bold
Ghost      -> border border-ink/20 text-ink rounded-pill px-6 py-3 font-bold bg-transparent
```

### Regla de oro del diseno Dydi
**Numero grande en Newsreader + descripcion pequena en Hanken Grotesk tenue = el look Dydi.**
Ejemplo: racha `13` en `font-serif text-5xl text-terracotta` + `dias de racha` en `text-eyebrow`.

---

## Routes & Views

| Route | View | Description |
|---|---|---|
| `/login` | `LoginView.vue` | Auth screen |
| `/today` | `TodayView.vue` | Countdown, check-in propio, resumen del squad |
| `/checkin` | `CheckinView.vue` | Flujo de check-in: inicial → exito (racha +1) → fallo |
| `/squad` | `SquadView.vue` | Lista de miembros con tiras de 7 dias |
| `/ruleta` | `RuletaView.vue` | Pool de penitencias, girar ruleta, resultado |
| `/tu` | `TuView.vue` | Perfil propio, historial, rachas |

The router uses **hash mode** (`createWebHashHistory`) for Vercel SPA compatibility.
After login, redirect to `/today`. Unauthenticated users redirect to `/login`.

---

## Folder Structure

```
src/
+-- components/
|   +-- ui/              <- reusable primitives (Button, Card, Badge, Avatar)
|   +-- squad/           <- SquadView components
|   +-- today/           <- TodayView and check-in components
|   +-- ruleta/          <- RuletaView components (roulette, debt list)
|   +-- tu/              <- TuView components (profile, streak history)
+-- composables/
|   +-- useGroupSocket.js   <- WebSocket connection + event handling
|   +-- useAuth.js          <- Auth state + Supabase Auth client
|   +-- useFormatters.js    <- date, streak, percentage formatters
+-- stores/
|   +-- auth.js          <- user session
|   +-- group.js         <- active group + members
|   +-- habits.js        <- today's habits + checkins
|   +-- penalties.js     <- debts + roulette state
+-- views/
|   +-- LoginView.vue
|   +-- TodayView.vue
|   +-- CheckinView.vue
|   +-- SquadView.vue
|   +-- RuletaView.vue
|   +-- TuView.vue
+-- router/
|   +-- index.js
+-- App.vue
+-- main.js
```

---

## Responsive Layout Rules

| Element | Mobile (< 768px) | Desktop (>= 1280px) |
|---|---|---|
| Navigation | Bottom tab bar (4 tabs) | Left sidebar (collapsible) |
| Squad grid | 2 columns | 4 columns |
| Roulette countdown | Top banner, full width | Fixed right panel |
| Font sizes display | text-4xl | text-6xl |

**Bottom nav tabs (mobile):** Hoy · Squad · Ruleta · Tu

---

## WebSocket (useGroupSocket composable)

The WebSocket connects to `realtime-service` via the gateway.
URL format: `wss://api-gateway.onrender.com/ws/:groupID`

Handle these event types:
```js
const eventHandlers = {
  'checkin':          (payload) => habitsStore.updateCheckin(payload),
  'streak_update':    (payload) => habitsStore.updateStreak(payload),
  'member_online':    (payload) => groupStore.setMemberOnline(payload.userID),
  'member_offline':   (payload) => groupStore.setMemberOffline(payload.userID),
  'roulette_start':   (payload) => penaltiesStore.startRoulette(payload),
  'roulette_result':  (payload) => penaltiesStore.setRouletteResult(payload),
  'debt_created':     (payload) => penaltiesStore.addDebt(payload),
}
```

Reconnection: `useWebSocket` from @vueuse/core handles auto-reconnect.
Set `autoReconnect: { retries: 10, delay: 3000 }`.

---

## API Calls Convention

All HTTP calls go through a single `api.js` utility, never raw `fetch` in components:
```js
// src/api.js
const BASE = import.meta.env.VITE_API_URL  // https://api-gateway.onrender.com

export async function api(path, options = {}) {
  const token = useAuthStore().token
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options.headers,
    }
  })
  if (!res.ok) throw await res.json()
  return res.json()
}
```

---

## Environment Variables
```
VITE_API_URL=https://api-gateway.onrender.com
VITE_WS_URL=wss://api-gateway.onrender.com
VITE_SUPABASE_URL=https://xxxx.supabase.co
VITE_SUPABASE_ANON_KEY=...
```

Never use `process.env` in Vue/Vite — always `import.meta.env`.

---

## Vercel Deployment Notes
- Vercel auto-detects Vite projects, no special config needed
- Add a `vercel.json` with SPA rewrite rule:
```json
{
  "rewrites": [{ "source": "/(.*)", "destination": "/index.html" }]
}
```
- Set all `VITE_*` env vars in Vercel dashboard under Project Settings → Environment Variables

---

## Component Rules
- Props must have explicit types defined with `defineProps`
- Emits must be declared with `defineEmits`
- No business logic in components — delegate to stores and composables
- Loading and error states must always be handled, never leave the UI hanging
- The LIVE dot must pulse using a CSS animation, not blink
