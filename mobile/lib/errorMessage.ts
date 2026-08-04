// Única fuente de mensajes de error del APK. El cuerpo que responde el backend
// NUNCA debe llegar a la pantalla.
//
// Por qué existe este módulo: los servicios responden cosas como
// {"error":"could not sync user"}, "missing X-User-ID" o "display_name is
// required". Es texto para quien depura, no para quien usa la app, y las
// pantallas lo pintaban tal cual con `e?.error ?? 'fallback'`. En una corrida de
// prueba el usuario vio literalmente {"status":401,"error":"invalid token"}.
//
// Lo que exigen las fuentes:
//   · WCAG 2.2 SC 3.3.1 (nivel A): el error se describe al usuario EN TEXTO —
//     el texto del servidor no describe el error DEL USUARIO.
//   · WCAG 2.2 SC 3.3.3 (nivel AA): si se conoce la corrección, hay que
//     ofrecerla. Todos los errores de Dydi tienen corrección conocida.
//   · NN/g H9: lenguaje llano, sin códigos, indicar el problema con precisión y
//     sugerir una solución; sin culpar al usuario.
//   · Apple HIG: el título dice qué pasó y por qué; prohibido "Error" o un código.
//
// `fallback` es lo que se dice cuando no reconocemos el caso: siempre una frase
// completa y accionable, nunca "Ocurrió un error".

type ApiError = { status?: number; error?: string; message?: string } | null | undefined;

// Casos que el backend distingue por texto y que merecen una frase propia.
// La clave se busca como subcadena en minúsculas del mensaje del servidor.
const BY_SERVER_TEXT: Array<[string, string]> = [
  ['already voted', 'Ya votaste esta propuesta.'],
  ['no longer open', 'Esta propuesta ya se cerró.'],
  ['has expired', 'Esta propuesta ya venció.'],
  ['not a member', 'Ya no eres parte de este grupo.'],
  ['group is full', 'Este grupo ya tiene sus 8 lugares ocupados.'],
  ['connection limit', 'Este grupo alcanzó el límite de conexiones en vivo.'],
  ['invite', 'Ese código de invitación no existe o ya se rotó. Pide uno nuevo al squad.'],
  ['already spun', 'La ruleta de esta persona ya se giró esta semana.'],
  ['not eligible', 'Esta persona no falló esta semana, así que no entra a la ruleta.'],
];

const BY_STATUS: Record<number, string> = {
  400: 'Faltan datos o alguno no tiene el formato esperado. Revísalos e intenta de nuevo.',
  401: 'Tu sesión expiró. Vuelve a iniciar sesión.',
  403: 'No tienes permiso para hacer esto.',
  404: 'Esto ya no existe.',
  409: 'Alguien se te adelantó y esto ya cambió. Vuelve a cargar la pantalla.',
  429: 'Vas muy rápido. Espera unos segundos e intenta de nuevo.',
  500: 'Algo falló de nuestro lado. Intenta de nuevo en un momento.',
  502: 'El servidor está despertando. Intenta de nuevo en unos segundos.',
  503: 'El servidor está despertando. Intenta de nuevo en unos segundos.',
  504: 'El servidor tardó demasiado en responder. Intenta de nuevo.',
};

/**
 * Convierte cualquier error de `api()` en una frase para una persona.
 * @param err   el error que lanzó api()
 * @param fallback frase específica de la acción, para cuando no reconozcamos el caso
 */
export function errorMessage(err: ApiError, fallback: string): string {
  if (!err) return fallback;

  // Fallo de transporte: fetch lanza TypeError, y un timeout aborta.
  const name = (err as { name?: string }).name;
  if (name === 'AbortError') {
    return 'La conexión tardó demasiado. Revisa tu internet e intenta de nuevo.';
  }
  if (err instanceof TypeError) {
    return 'Sin conexión. Revisa tu internet e intenta de nuevo.';
  }

  const raw = `${err.error ?? ''} ${err.message ?? ''}`.toLowerCase();
  for (const [needle, human] of BY_SERVER_TEXT) {
    if (raw.includes(needle)) return human;
  }

  if (err.status && BY_STATUS[err.status]) return BY_STATUS[err.status];

  return fallback;
}
