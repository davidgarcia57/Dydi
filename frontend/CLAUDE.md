# CLAUDE.md — frontend (Dydi)

## Purpose
Vue 3 SPA deployed on Vercel. Mobile-first, responsive up to desktop (1280px+).
Communicates with the backend exclusively through `api-gateway`.
Real-time updates via WebSocket connection to `realtime-service` (through gateway).

---

## CRITICAL: Vue 3 Composition API only
**Always use `<script setup>` syntax.** Never use Options API or the old
`defineComponent()` pattern. If you see Options API in existing code, ask
before refactoring it.

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
| Vite | Build tool | Do not switch to webpack or other bundlers |
| Pinia | State management | One store per domain (auth, group, habits, penalties) |
| Tailwind CSS | Styling | Utility classes only, no custom CSS unless unavoidable |
| @vueuse/core | Utilities | Use `useWebSocket` for WebSocket, `useStorage` for local persistence |
| Vue Router | Routing | Hash mode for Vercel SPA compatibility |
| Vercel | Hosting | Auto-deploy from main branch |

---

## Design System (Dydi Identity)

### Colors — use these CSS variables, never hardcode hex in components
```css
--color-bg:       #0D0D0D;   /* page background */
--color-surface:  #1A1A1A;   /* cards, modals */
--color-surface2: #242424;   /* nested surfaces */
--color-primary:  #7C5CFC;   /* violet — primary actions */
--color-live:     #22D3EE;   /* cyan — realtime/live indicators */
--color-success:  #22C55E;   /* green — streaks, completed */
--color-danger:   #FF4D4D;   /* red — debts, missed habits */
--color-warning:  #F97316;   /* orange — mid progress */
--color-text:     #F5F5F5;   /* primary text */
--color-muted:    #94A3B8;   /* secondary text */
```

### Typography
- Display / big numbers (streaks, percentages, countdown): `font-bold text-4xl+`, font: DM Sans
- UI / body: Inter
- Key rule: **large bold numbers + small muted descriptions** = the Dydi look

### Tailwind classes for Dydi surfaces
```
bg-[#0D0D0D]     ← page background
bg-[#1A1A1A]     ← card background
bg-[#242424]     ← nested card
rounded-xl        ← standard card radius
border border-white/5  ← subtle card border
```

---

## Folder Structure

```
src/
├── components/
│   ├── ui/              ← reusable primitives (Button, Card, Badge, Avatar)
│   ├── squad/           ← Squad tab components
│   ├── today/           ← Today/check-in components
│   ├── roulette/        ← Saturday roulette components
│   └── shame/           ← Wall of Shame components
├── composables/
│   ├── useGroupSocket.js   ← WebSocket connection + event handling
│   ├── useAuth.js          ← Auth state + Supabase Auth client
│   └── useFormatters.js    ← date, streak, percentage formatters
├── stores/
│   ├── auth.js          ← user session
│   ├── group.js         ← active group + members
│   ├── habits.js        ← today's habits + checkins
│   └── penalties.js     ← debts + roulette state
├── views/
│   ├── SquadView.vue
│   ├── TodayView.vue
│   ├── TrialView.vue    ← Roulette screen
│   └── ShameView.vue
├── router/
│   └── index.js
├── App.vue
└── main.js
```

---

## Responsive Layout Rules

| Element | Mobile (< 768px) | Desktop (≥ 1280px) |
|---|---|---|
| Navigation | Bottom tab bar | Left sidebar (collapsible) |
| Squad grid | 2 columns | 4 columns |
| Roulette countdown | Top banner, full width | Fixed right panel |
| Live activity feed | Inline below squad grid | Right column, always visible |
| Font sizes display | text-4xl | text-6xl |

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
- The LIVE cyan dot (`● LIVE`) must pulse using a CSS animation, not blink
