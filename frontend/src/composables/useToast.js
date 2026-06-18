import { reactive } from 'vue'

// Toast minimalista de un solo mensaje (suficiente para feedback efímero como
// "código copiado"). Reemplaza el alert() nativo, que rompía la estética.
const state = reactive({ message: '', visible: false })
let timer = null

export function showToast(message, duration = 2600) {
  state.message = message
  state.visible = true
  if (timer) clearTimeout(timer)
  timer = setTimeout(() => {
    state.visible = false
  }, duration)
}

export function useToast() {
  return { toast: state, showToast }
}
