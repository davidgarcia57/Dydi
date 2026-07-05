import React, { useMemo } from 'react';
import { View, Text } from 'react-native';
import { useAuth } from '../contexts/AuthContext';
import { useApp } from '../contexts/AppContext';
import TargetGlyph from './ui/TargetGlyph';
import { missedThisWeek, todayStatus } from '../weekStatus';

// Pulso del squad: de un vistazo, quién ya cumplió hoy (anillo salvia), quién
// va pendiente (ámbar), quién dejó ir el día (coral), quién está en riesgo de
// ruleta (diana) y quién anda conectado (punto verde). Espejo del SquadPulse
// del frontend web, adaptado a móvil.
const AVATAR_COLORS = [
  'bg-sage-deep',
  'bg-terracotta',
  'bg-sage',
  'bg-amber',
  'bg-coral',
  'bg-ink-soft',
];

const RING: Record<string, string> = {
  done: 'border-sage-deep',
  pending: 'border-amber',
  missed: 'border-coral',
};

function initials(name = '') {
  return name
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();
}

function avatarBg(name = '') {
  const charCode = name.length > 0 ? name.charCodeAt(0) : 0;
  return AVATAR_COLORS[charCode % AVATAR_COLORS.length];
}

export default function SquadPulse() {
  const { user } = useAuth();
  const { members, onlineMembers, todayCheckins, weekHistory } = useApp();

  const pulse = useMemo(() => {
    const me = {
      user_id: user?.id ?? '',
      display_name: (user?.user_metadata?.display_name as string) || 'Tú',
    };
    const all: { user_id: string; display_name: string }[] = [
      me,
      ...members.filter((m) => m.user_id !== user?.id),
    ];
    return all
      .filter((m) => m.user_id)
      .map((m) => ({
        ...m,
        status: todayStatus(m.user_id, todayCheckins),
        atRisk: missedThisWeek(m.user_id, todayCheckins, weekHistory) > 0,
        online: onlineMembers.has(m.user_id),
        isMe: m.user_id === user?.id,
      }));
  }, [members, user, todayCheckins, weekHistory, onlineMembers]);

  if (!pulse.length) return null;

  return (
    <View className="flex-row flex-wrap" style={{ columnGap: 14, rowGap: 10 }}>
      {pulse.map((m) => (
        <View key={m.user_id} className="items-center w-14">
          <View
            className={`rounded-full border-2 p-0.5 ${m.status ? RING[m.status] : 'border-hairline'}`}
          >
            <View
              className={`w-9 h-9 rounded-full items-center justify-center ${avatarBg(m.display_name)}`}
            >
              <Text className="text-paper text-xs font-bold">{initials(m.display_name)}</Text>
            </View>
            {m.online && (
              <View className="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-sage-deep border-2 border-paper" />
            )}
            {m.atRisk && (
              <View className="absolute -top-1 -right-1.5 w-5 h-5 rounded-full bg-paper border border-hairline items-center justify-center">
                <TargetGlyph size={11} />
              </View>
            )}
          </View>
          <Text className="text-[10px] font-semibold text-ink-soft mt-1" numberOfLines={1}>
            {m.isMe ? 'Tú' : m.display_name.split(' ')[0]}
          </Text>
        </View>
      ))}
    </View>
  );
}
