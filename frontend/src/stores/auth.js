import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createClient } from '@supabase/supabase-js'

const API_BASE = import.meta.env.VITE_API_URL
const SUPABASE_URL = import.meta.env.VITE_SUPABASE_URL
const SUPABASE_ANON_KEY = import.meta.env.VITE_SUPABASE_ANON_KEY

const hasSupabaseConfig = Boolean(SUPABASE_URL && SUPABASE_ANON_KEY)
const supabase = hasSupabaseConfig ? createClient(SUPABASE_URL, SUPABASE_ANON_KEY) : null

export const useAuthStore = defineStore('auth', () => {
  const session = ref(null)

  const user = computed(() => session.value?.user ?? null)
  const token = computed(() => session.value?.access_token ?? null)
  const isLoggedIn = computed(() => !!session.value)

  function profileFromSession(currentSession, fallbackName = '') {
    const currentUser = currentSession?.user
    const metadata = currentUser?.user_metadata ?? {}
    const displayName = fallbackName || metadata.display_name || currentUser?.email?.split('@')[0] || 'Usuario Dydi'

    return {
      display_name: displayName,
      avatar_url: metadata.avatar_url ?? null,
    }
  }

  async function syncUser(fallbackName = '') {
    if (!session.value?.access_token || !API_BASE) return

    const res = await fetch(`${API_BASE}/api/users/sync`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.value.access_token}`,
      },
      body: JSON.stringify(profileFromSession(session.value, fallbackName)),
    })

    if (!res.ok) {
      throw new Error('No pudimos sincronizar tu perfil. Intenta de nuevo.')
    }
  }

  async function init() {
    if (!supabase) return

    const { data } = await supabase.auth.getSession()
    session.value = data.session
    supabase.auth.onAuthStateChange((_event, s) => {
      session.value = s
    })
  }

  async function login(email, password) {
    if (!supabase) {
      throw new Error('Faltan las variables de Supabase para iniciar sesión.')
    }

    const { data, error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) throw error
    session.value = data.session
    await syncUser()
  }

  async function register(email, password, displayName) {
    if (!supabase) {
      throw new Error('Faltan las variables de Supabase para crear cuentas.')
    }

    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: { data: { display_name: displayName } },
    })
    if (error) throw error
    session.value = data.session
    if (session.value) {
      await syncUser(displayName)
    }
  }

  async function logout() {
    if (!supabase) return

    await supabase.auth.signOut()
    session.value = null
  }

  return { session, user, token, isLoggedIn, init, login, register, logout, syncUser }
})
