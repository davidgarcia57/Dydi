import { Platform } from 'react-native';
import * as Notifications from 'expo-notifications';

// Notificaciones LOCALES únicamente (recordatorio de check-in + ruleta abierta
// vista por WS en foreground). Sin FCM ni servidor push: no requiere
// credenciales extra en el build del APK.

const REMINDER_ID = 'daily-checkin-reminder';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: false,
    shouldSetBadge: false,
  }),
});

// Recordatorio diario a las 20:00 locales. Idempotente: mismo identifier,
// se cancela y re-agenda en cada arranque con grupo activo.
export async function setupCheckinReminder() {
  if (Platform.OS === 'web') return;
  try {
    const { status } = await Notifications.requestPermissionsAsync();
    if (status !== 'granted') return;
    if (Platform.OS === 'android') {
      await Notifications.setNotificationChannelAsync('default', {
        name: 'Recordatorios',
        importance: Notifications.AndroidImportance.DEFAULT,
      });
    }
    await Notifications.cancelScheduledNotificationAsync(REMINDER_ID).catch(() => {});
    await Notifications.scheduleNotificationAsync({
      identifier: REMINDER_ID,
      content: {
        title: 'Dydi',
        body: '¿Ya hiciste tu check-in de hoy? Tu racha está en juego.',
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.DAILY,
        hour: 20,
        minute: 0,
      },
    });
  } catch (err) {
    console.warn('No se pudo agendar el recordatorio:', err);
  }
}

// Aviso inmediato cuando llega roulette_start por WebSocket (app abierta).
export async function notifyRouletteOpened(debtorName: string, isMe: boolean) {
  if (Platform.OS === 'web') return;
  try {
    await Notifications.scheduleNotificationAsync({
      content: {
        title: isMe ? '¡Estás en la ruleta!' : `Ruleta abierta para ${debtorName}`,
        body: isMe
          ? 'Tu squad está escribiendo tu penitencia…'
          : 'Entra a proponer una penitencia antes de que cierre.',
      },
      trigger: null,
    });
  } catch {
    // best-effort: sin permiso simplemente no avisa
  }
}
