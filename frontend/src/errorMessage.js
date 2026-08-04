// Única fuente de mensajes de error de la web. Espejo de mobile/lib/errorMessage.ts:
// si cambias uno, cambia el otro — el usuario puede ver ambos clientes.
//
// Por qué existe: los servicios responden cosas como {"error":"could not sync
// user"} o "display_name is required", y las vistas lo pintaban tal cual con
// `e?.error ?? 'fallback'`. Es texto para quien depura, no para quien usa la app.
//
// Lo que exigen las fuentes (ver docs/ux-guidelines.md):
//   · WCAG 2.2 SC 3.3.1 (A): el error se describe al usuario EN TEXTO.
//   · WCAG 2.2 SC 3.3.3 (AA): si se conoce la corrección, hay que ofrecerla.
//   · NN/g H9: lenguaje llano, sin códigos, con solución y sin culpar al usuario.
//   · Apple HIG: el título dice qué pasó; prohibido "Error" o un código.

// Casos que el backend distingue por texto y merecen una frase propia.
// La clave se busca como subcadena, en minúsculas.
const BY_SERVER_TEXT = [
  ['already voted', 'Ya votaste esta propuesta.'],
  ['no longer open', 'Esta propuesta ya se cerró.'],
  ['has expired', 'Esta propuesta ya venció.'],
  ['not a member', 'Ya no eres parte de este grupo.'],
  ['group is full', 'Este grupo ya tiene sus 8 lugares ocupados.'],
  ['connection limit', 'Este grupo alcanzó el límite de conexiones en vivo.'],
  ['invite', 'Ese código de invitación no existe o ya se rotó. Pide uno nuevo al squad.'],
  ['already spun', 'La ruleta de esta persona ya se giró esta semana.'],
  ['not eligible', 'Esta persona no falló esta semana, así que no entra a la ruleta.'],
]

const BY_STATUS = {
  400: 'Faltan datos o alguno no tiene el formato esperado. Revísalos e intenta de nuevo.',
  401: 'Tu sesión expiró. Vuelve a iniciar sesión.',
  403: 'No tienes permiso para hacer esto.',
  404: 'Esto ya no existe.',
  409: 'Alguien se te adelantó y esto ya cambió. Recarga la página.',
  429: 'Vas muy rápido. Espera unos segundos e intenta de nuevo.',
  500: 'Algo falló de nuestro lado. Intenta de nuevo en un momento.',
  502: 'El servidor está despertando. Intenta de nuevo en unos segundos.',
  503: 'El servidor está despertando. Intenta de nuevo en unos segundos.',
  504: 'El servidor tardó demasiado en responder. Intenta de nuevo.',
}

/**
 * Convierte cualquier error de `api()` en una frase para una persona.
 * @param {*} err error lanzado por api()
 * @param {string} fallback frase específica de la acción, para cuando no reconozcamos el caso
 * @returns {string}
 */
export function errorMessage(err, fallback) {
  if (!err) return fallback

  if (err.name === 'AbortError') {
    return 'La conexión tardó demasiado. Revisa tu internet e intenta de nuevo.'
  }
  if (err instanceof TypeError) {
    return 'Sin conexión. Revisa tu internet e intenta de nuevo.'
  }

  const raw = `${err.error ?? ''} ${err.message ?? ''}`.toLowerCase()
  for (const [needle, human] of BY_SERVER_TEXT) {
    if (raw.includes(needle)) return human
  }

  if (err.status && BY_STATUS[err.status]) return BY_STATUS[err.status]

  return fallback
}
