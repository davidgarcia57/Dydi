import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createClient } from '@supabase/supabase-js'

const SUPABASE_URL = import.meta.env.VITE_SUPABASE_URL
const SUPABASE_ANON_KEY = import.meta.env.VITE_SUPABASE_ANON_KEY

const hasSupabaseConfig = Boolean(SUPABASE_URL && SUPABASE_ANON_KEY)
const supabase = hasSupabaseConfig ? createClient(SUPABASE_URL, SUPABASE_ANON_KEY) : null

export const useAuthStore = defineStore('auth', () => {
  const session = ref(null)

  const user = computed(() => session.value?.user ?? null)
  const token = computed(() => session.value?.access_token ?? null)
  const isLoggedIn = computed(() => !!session.value)

  async function init() {
    if (!supabase) return
    const { data } = await supabase.auth.getSession()
    session.value = data.session
    supabase.auth.onAuthStateChange((_event, s) => {
      session.value = s
    })
  }

  async function login(email, password) {
    if (!supabase) throw new Error('Faltan las variables de Supabase para iniciar sesión.')
    const { data, error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) throw error
    session.value = data.session
  }

  async function register(email, password, displayName) {
    if (!supabase) throw new Error('Faltan las variables de Supabase para crear cuentas.')
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: { data: { display_name: displayName } },
    })
    if (error) throw error

    // Anti-enumeración de Supabase: cuando el correo YA está registrado, signUp
    // no devuelve error sino un usuario "fantasma" con identities vacío.
    // Lo detectamos para no permitir un registro duplicado.
    if (data.user && Array.isArray(data.user.identities) && data.user.identities.length === 0) {
      throw new Error('EMAIL_TAKEN')
    }

    session.value = data.session
  }

  async function logout() {
    if (!supabase) return
    await supabase.auth.signOut()
    session.value = null
  }

  return { session, user, token, isLoggedIn, init, login, register, logout }
})
