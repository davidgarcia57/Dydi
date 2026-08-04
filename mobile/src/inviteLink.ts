// Texto y enlace de invitación, en un solo lugar. Espejo del inviteLink.js de la
// web.
//
// El enlace apunta a la web porque es donde vive la ruta /join: quien recibe la
// invitación probablemente no tiene la app instalada, así que mandarlo a un deep
// link `dydi://` no llevaría a ningún lado. Desde el navegador puede registrarse y
// entrar al squad, y luego bajar el APK si quiere.
const WEB_URL = process.env.EXPO_PUBLIC_WEB_URL || 'https://dydi-xi.vercel.app';

export function inviteURL(groupID: string, inviteCode: string): string {
  // Hash routing en la web: el /join va después del #.
  return `${WEB_URL}/#/join?g=${encodeURIComponent(groupID)}&c=${encodeURIComponent(inviteCode)}`;
}

export function inviteMessage(groupName: string, groupID: string, inviteCode: string): string {
  return (
    `¡Únete a mi squad "${groupName}" en Dydi!\n\n` +
    `${inviteURL(groupID, inviteCode)}\n\n` +
    `Si te pide el código a mano: ${inviteCode}`
  );
}
