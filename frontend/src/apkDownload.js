// Descarga del APK.
//
// Permalink de GitHub al asset de la ÚLTIMA release: `releases/latest/download/`
// devuelve un 302 al tag más nuevo, así que este enlace NO hay que tocarlo cuando
// se saca versión — apunta solo. Verificado: 302 → v1.1.5 → 200 con
// content-type application/vnd.android.package-archive.
//
// Se enlaza a GitHub en vez de servir el archivo desde la web por dos razones: el
// APK pesa ~99 MB (meterlo al despliegue lo infla y obliga a redesplegar la web en
// cada release), y GitHub ya paga ese ancho de banda.
//
// ⚠️ Depende de que el asset se siga llamando `app-release.apk`. Lo publica con
// ese nombre `.github/workflows/build-apk.yml`; si ahí se renombra, este enlace
// se rompe en silencio (GitHub responde 404).
export const APK_URL =
  'https://github.com/davidgarcia57/Dydi/releases/latest/download/app-release.apk'

// Página de releases, para quien quiera ver notas o una versión anterior.
export const RELEASES_URL = 'https://github.com/davidgarcia57/Dydi/releases'
