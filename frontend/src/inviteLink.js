// Texto y enlace de invitación, en un solo lugar.
//
// Antes cada vista armaba el string `uuid:CODIGO` por su cuenta (siete copias
// entre web y móvil), lo que garantizaba que arreglarlo en una y olvidar las
// otras. El enlace importa más de lo que parece: en un producto con cold start,
// volver la invitación un tap en vez de copiar-pegar un código es lo que decide
// si el squad crece.
//
// `window.location.origin` y no una URL configurada: así el enlace siempre apunta
// al mismo despliegue desde el que se comparte, sin variable de entorno que se
// pueda quedar desactualizada.

export function inviteURL(groupID, inviteCode) {
  // Hash routing: el /join va después del #, como el resto de las rutas.
  return `${window.location.origin}/#/join?g=${encodeURIComponent(groupID)}&c=${encodeURIComponent(inviteCode)}`
}

export function inviteMessage(groupName, groupID, inviteCode) {
  return (
    `¡Únete a mi squad "${groupName}" en Dydi!\n\n` +
    `${inviteURL(groupID, inviteCode)}\n\n` +
    `Si te pide el código a mano: ${inviteCode}`
  )
}
