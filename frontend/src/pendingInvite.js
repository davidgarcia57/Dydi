// Invitación en espera de que haya sesión.
//
// El caso que importa del enlace de invitación es que te inviten SIN tener
// cuenta: se hace clic, no hay sesión, y el guard del router manda al login
// perdiendo el destino. Aquí se guarda la invitación para recuperarla en cuanto
// haya sesión.
//
// sessionStorage y no localStorage: una invitación pendiente no debe sobrevivir a
// cerrar el navegador — si alguien vuelve tres días después ya no tiene sentido
// meterlo a un squad que ni recuerda.
const KEY = 'dydi.pendingInvite'

export function savePendingInvite(g, c) {
  try {
    sessionStorage.setItem(KEY, JSON.stringify({ g, c }))
  } catch {
    // Modo privado o storage lleno: el enlace sigue funcionando para quien ya
    // tiene sesión, solo se pierde el rescate post-login.
  }
}

export function readPendingInvite() {
  try {
    const raw = sessionStorage.getItem(KEY)
    if (!raw) return null
    const { g, c } = JSON.parse(raw)
    return g && c ? { g, c } : null
  } catch {
    return null
  }
}

export function clearPendingInvite() {
  try {
    sessionStorage.removeItem(KEY)
  } catch {
    /* nada que limpiar */
  }
}
