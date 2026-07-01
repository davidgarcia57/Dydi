import { Platform } from 'react-native';
import * as Notifications from 'expo-notifications';

// Notificaciones LOCALES únicamente (sin servidor push ni credenciales FCM):
// recordatorio diario de check-in pendiente y aviso de ruleta recibida por
// WebSocket. En web expo-notifications no está soportado → todo es no-op.
const supported = Platform.OS !== 'web';

const CHANNEL_ID = 'default';
const REMINDER_KIND = 'checkin-reminder';
// Hora local del recordatorio diario.
const REMINDER_HOUR = 20;
// Días agendados por adelantado: cubre días sin abrir la app (una notificación
// repetitiva no permite saltarse "hoy" cuando ya se cumplió; por eso se agendan
// one-shots y cada sync los re-crea).
const REMINDER_DAYS_AHEAD = 7;

let handlerReady = false;

function setupHandler() {
  if (handlerReady) return;
  handlerReady = true;
  Notifications.setNotificationHandler({
    // En Android, shouldPlaySound:false suprime el banner heads-up.
    handleNotification: async () => ({
      shouldShowBanner: true,
      shouldShowList: true,
      shouldPlaySound: true,
      shouldSetBadge: false,
    }),
  });
}

async function hasPermission(): Promise<boolean> {
  const current = await Notifications.getPermissionsAsync();
  return current.granted;
}

// Configura handler + canal y pide permiso una vez que hay sesión (el SO solo
// muestra el prompt si el usuario aún no ha decidido; Android 13+ lo exige).
export async function initNotifications(): Promise<boolean> {
  if (!supported) return false;
  setupHandler();
  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync(CHANNEL_ID, {
      name: 'Recordatorios y ruleta',
      importance: Notifications.AndroidImportance.HIGH,
    });
  }
  const current = await Notifications.getPermissionsAsync();
  if (current.granted) return true;
  if (!current.canAskAgain) return false;
  const asked = await Notifications.requestPermissionsAsync();
  return asked.granted;
}

// Re-agenda los recordatorios según el estado actual: cancela los nuestros y
// crea uno por día a las REMINDER_HOUR. Hoy solo si sigue pendiente; los días
// futuros siempre (un día nuevo arranca con todo pendiente y quizá no se abra
// la app). Con hasHabits=false solo cancela (sin hábitos, o al salir del grupo).
export async function syncCheckinReminders(opts: {
  hasHabits: boolean;
  hasPendingToday: boolean;
}): Promise<void> {
  if (!supported || !(await hasPermission())) return;

  const scheduled = await Notifications.getAllScheduledNotificationsAsync();
  await Promise.all(
    scheduled
      .filter((n) => n.content.data?.kind === REMINDER_KIND)
      .map((n) => Notifications.cancelScheduledNotificationAsync(n.identifier))
  );
  if (!opts.hasHabits) return;

  const now = new Date();
  for (let day = 0; day < REMINDER_DAYS_AHEAD; day++) {
    if (day === 0 && !opts.hasPendingToday) continue;
    const fireAt = new Date(
      now.getFullYear(),
      now.getMonth(),
      now.getDate() + day,
      REMINDER_HOUR,
      0,
      0
    );
    if (fireAt <= now) continue;
    await Notifications.scheduleNotificationAsync({
      content: {
        title: 'Check-in pendiente',
        body: 'Aún no marcas tus hábitos de hoy. Hazlo antes de que cierre el día 🎯',
        data: { kind: REMINDER_KIND },
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.DATE,
        date: fireAt,
        channelId: CHANNEL_ID,
      },
    });
  }
}

// Aviso inmediato cuando llega por WebSocket una ruleta contra este usuario.
export async function notifyRouletteOnMe(groupName?: string): Promise<void> {
  if (!supported || !(await hasPermission())) return;
  await Notifications.scheduleNotificationAsync({
    content: {
      title: '¡Abrieron la ruleta en tu contra! 🎰',
      body: groupName
        ? `Tu squad "${groupName}" abrió la ruleta de penitencias contra ti. Entra a ver las sugerencias.`
        : 'Tu squad abrió la ruleta de penitencias contra ti. Entra a ver las sugerencias.',
      data: { kind: 'roulette-alert' },
    },
    trigger: { channelId: CHANNEL_ID },
  });
}
