import React, { useEffect, useState, useMemo } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Share,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp, type Checkin } from '../../src/contexts/AppContext';
import BrandWordmark from '../../src/components/ui/BrandWordmark';
import SquadPulse from '../../src/components/SquadPulse';
import TargetGlyph from '../../src/components/ui/TargetGlyph';
import { missedThisWeek } from '../../src/weekStatus';

const AVATAR_COLORS = [
  'bg-sage-deep',
  'bg-terracotta',
  'bg-sage',
  'bg-amber',
  'bg-coral',
  'bg-ink-soft',
];

function getInitials(name = '') {
  return name
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();
}

function getAvatarBg(name = '') {
  const charCode = name.length > 0 ? name.charCodeAt(0) : 0;
  return AVATAR_COLORS[charCode % AVATAR_COLORS.length];
}

const DAY_LABELS = ['L', 'M', 'M', 'J', 'V', 'S', 'D'];

const STATUS_STYLE: Record<string, { strip: string; icon: string; iconColor: string }> = {
  done: { strip: 'bg-sage', icon: '✓', iconColor: 'text-sage-deep' },
  pending: { strip: 'bg-amber', icon: '', iconColor: '' },
  missed: { strip: 'bg-coral', icon: '✗', iconColor: 'text-coral-deep' },
  future: { strip: 'border border-dashed border-hairline bg-transparent', icon: '', iconColor: '' },
};

const STATUS_PILL: Record<string, { cls: string; label: string }> = {
  done: { cls: 'bg-sage-soft text-sage-deep', label: '✓ hoy' },
  pending: { cls: 'bg-amber-soft text-amber-deep', label: 'pendiente' },
  missed: { cls: 'bg-coral-soft text-coral-deep', label: '✗ hoy' },
};

export default function TodayScreen() {
  const router = useRouter();
  const { user } = useAuth();
  const {
    group,
    members,
    onlineMembers,
    todayCheckins,
    streaks,
    weekHistory,
    autoLoad,
    loadAllData,
    wsConnected,
  } = useApp();

  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [now, setNow] = useState(new Date());

  useEffect(() => {
    async function checkGroup() {
      try {
        const hasGroup = await autoLoad();
        if (!hasGroup) {
          router.replace('/onboarding');
        }
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    checkGroup();
  }, []);

  // Countdown and cycle calculation
  useEffect(() => {
    const timer = setInterval(() => {
      setNow(new Date());
    }, 30000);
    return () => clearInterval(timer);
  }, []);

  // La ruleta despierta el sábado 00:00; sábado y domingo son su fin de semana.
  const isRouletteWeekend = useMemo(() => {
    const dow = now.getDay();
    return dow === 6 || dow === 0;
  }, [now]);

  const rouletteStart = useMemo(() => {
    const d = new Date(now);
    const daysLeft = (6 - d.getDay() + 7) % 7 || 7;
    d.setDate(d.getDate() + daysLeft);
    d.setHours(0, 0, 0, 0);
    return d;
  }, [now]);

  const countdown = useMemo(() => {
    const diff = rouletteStart.getTime() - now.getTime();
    if (diff <= 0) return { days: '00', hours: '00', mins: '00' };
    const pad = (n: number) => String(n).padStart(2, '0');
    return {
      days: pad(Math.floor(diff / 86400000)),
      hours: pad(Math.floor((diff % 86400000) / 3600000)),
      mins: pad(Math.floor((diff % 3600000) / 60000)),
    };
  }, [rouletteStart, now]);

  const weekNumber = useMemo(() => {
    const d = new Date();
    const jan1 = new Date(d.getFullYear(), 0, 1);
    return Math.ceil(((d.getTime() - jan1.getTime()) / 86400000 + jan1.getDay() + 1) / 7);
  }, []);

  // Today checkins and status
  const myCheckins = useMemo(() => {
    return todayCheckins.filter((c) => c.user_id === user?.id);
  }, [todayCheckins, user?.id]);

  const hasPending = useMemo(() => myCheckins.some((c) => c.status === 'pending'), [myCheckins]);
  const allDone = useMemo(() => myCheckins.length > 0 && myCheckins.every((c) => c.status === 'done'), [myCheckins]);
  const anyMissed = useMemo(() => myCheckins.some((c) => c.status === 'missed') && !hasPending, [myCheckins, hasPending]);

  const myStreak = useMemo(() => {
    return user?.id ? streaks[user.id] ?? 0 : 0;
  }, [streaks, user?.id]);

  // Riesgo de ruleta: fallos L–V acumulados esta semana (días sin check-in).
  const myMissed = useMemo(
    () => (user?.id ? missedThisWeek(user.id, todayCheckins, weekHistory) : 0),
    [user?.id, todayCheckins, weekHistory]
  );

  const riskBanner = useMemo(() => {
    if (!myMissed) return null;
    if (isRouletteWeekend) {
      return {
        text: `Estás en el bote con ${myMissed} ${myMissed === 1 ? 'fallo' : 'fallos'} esta semana. La ruleta te espera.`,
        cta: true,
      };
    }
    if (myMissed === 1) {
      return { text: 'Llevas 1 fallo esta semana — el sábado entras al bote.', cta: false };
    }
    return { text: `${myMissed} fallos esta semana. La ruleta ya te tiene en la lista.`, cta: false };
  }, [myMissed, isRouletteWeekend]);

  // Squad Stats
  const stats = useMemo(() => {
    const byUser: Record<string, string[]> = {};
    for (const c of todayCheckins) {
      if (!byUser[c.user_id]) byUser[c.user_id] = [];
      byUser[c.user_id].push(c.status);
    }
    let done = 0, pending = 0, missed = 0;
    for (const statuses of Object.values(byUser)) {
      if (statuses.every((s) => s === 'done')) done++;
      else if (statuses.some((s) => s === 'pending')) pending++;
      else missed++;
    }
    return { done, pending, missed };
  }, [todayCheckins]);

  const progressPct = useMemo(() => {
    const total = members.length + 1; // members + me
    return total ? Math.round((stats.done / total) * 100) : 0;
  }, [members.length, stats.done]);

  // Online members
  const onlineAvatars = useMemo(() => {
    const allSquad = [...members, { user_id: user?.id, display_name: user?.user_metadata?.display_name || 'Tú' }];
    return allSquad.filter((m) => m.user_id && onlineMembers.has(m.user_id)).slice(0, 5);
  }, [members, user, onlineMembers]);

  // Other members, grouped into one card each. Habits are group-wide, so a
  // member with N habits would otherwise render as N "duplicate" cards.
  const squadMembers = useMemo(() => {
    const byUser = new Map<string, { user_id: string; display_name: string; habits: typeof todayCheckins }>();
    for (const c of todayCheckins) {
      if (c.user_id === user?.id) continue;
      const entry = byUser.get(c.user_id);
      if (entry) entry.habits.push(c);
      else byUser.set(c.user_id, { user_id: c.user_id, display_name: c.display_name, habits: [c] });
    }
    return Array.from(byUser.values());
  }, [todayCheckins, user?.id]);

  // A member's overall status: done only if every habit is done, else pending,
  // unless all are missed.
  function memberStatus(habitsList: typeof todayCheckins): string {
    if (habitsList.every((h) => h.status === 'done')) return 'done';
    if (habitsList.some((h) => h.status === 'pending')) return 'pending';
    return 'missed';
  }

  const displayUserName = user?.user_metadata?.display_name || user?.email || 'Tú';

  async function shareInvite() {
    if (!group) return;
    const code = `${group.id}:${group.invite_code}`;
    const text = `¡Únete a mi squad "${group.name}" en Dydi!\nCódigo de invitación: ${code}`;
    try {
      await Share.share({
        message: text,
        title: 'Únete a Dydi',
      });
    } catch (e) {
      console.error(e);
    }
  }

  async function handleRefresh() {
    setRefreshing(true);
    try {
      await loadAllData();
    } catch (err) {
      console.error(err);
    } finally {
      setRefreshing(false);
    }
  }

  function getMemberDayStrip(checkin: Checkin) {
    const dow = new Date().getDay();
    const todayIdx = dow === 0 ? 6 : dow - 1;
    const key = `${checkin.user_id}:${checkin.habit_id}`;
    const dates = weekHistory[key];

    return DAY_LABELS.map((label, i) => {
      if (i > todayIdx) return { label, status: 'future' };
      if (i === todayIdx) return { label, status: checkin.status };
      
      const d = new Date();
      d.setDate(d.getDate() - (todayIdx - i));
      const dateStr = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
      const done = dates ? dates.has(dateStr) : false;
      return { label, status: done ? 'done' : 'missed' };
    });
  }

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center bg-cream">
        <ActivityIndicator size="large" color="#7CA39D" />
      </View>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-cream" edges={['top']}>
      {/* Header */}
      <View className="px-6 py-3 flex-row items-center justify-between border-b border-hairline/30 bg-cream">
        <BrandWordmark size="sm" />
        
        {group && (
          <TouchableOpacity
            activeOpacity={0.8}
            onPress={shareInvite}
            className="flex-row items-center gap-1.5 rounded-full border border-hairline px-3 py-1.5 bg-surface"
          >
            <Text className="text-xs font-bold text-ink truncate max-w-[100px]">{group.name}</Text>
            <Text className="text-xs font-bold text-terracotta">Invitar</Text>
          </TouchableOpacity>
        )}

        <View className={`w-9 h-9 rounded-full flex items-center justify-center ${getAvatarBg(displayUserName)}`}>
          <Text className="text-paper text-sm font-bold">{getInitials(displayUserName)}</Text>
        </View>
      </View>

      <ScrollView
        className="flex-1 px-6"
        contentContainerStyle={{ paddingVertical: 16 }}
        showsVerticalScrollIndicator={false}
      >
        {/* Pulso del squad */}
        <View className="rounded-3xl bg-paper border border-hairline p-4 mb-4 shadow-sm">
          <View className="flex-row justify-between items-center mb-3">
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
              EL SQUAD HOY
            </Text>
            {onlineAvatars.length > 0 && (
              <View className="flex-row items-center gap-1">
                <View className="w-2 h-2 rounded-full bg-sage-deep" />
                <Text className="text-[10px] font-bold text-sage-deep tracking-wider uppercase">
                  {onlineAvatars.length} EN VIVO
                </Text>
              </View>
            )}
          </View>
          <SquadPulse />
        </View>

        {/* En riesgo de ruleta */}
        {riskBanner && (
          <View className="rounded-3xl bg-amber-soft border border-amber/40 p-4 mb-4">
            <View className="flex-row items-center gap-2">
              <TargetGlyph size={16} color="#A57B33" />
              <Text className="text-sm font-semibold text-amber-deep flex-1">
                {riskBanner.text}
              </Text>
            </View>
            {riskBanner.cta && (
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => router.push('/(tabs)/shame')}
                className="mt-3 rounded-full bg-terracotta py-2.5 items-center"
              >
                <Text className="text-paper text-sm font-bold">Ir a la ruleta →</Text>
              </TouchableOpacity>
            )}
          </View>
        )}

        {/* Countdown Card */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-4 shadow-sm">
          <View className="flex-row justify-between items-center mb-3">
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
              {isRouletteWeekend ? 'LA RULETA ESTÁ DESPIERTA' : 'LA RULETA GIRA EN'}
            </Text>
            {!isRouletteWeekend && (
              <Text className="text-xs font-semibold text-terracotta">sáb 00:00</Text>
            )}
          </View>

          {isRouletteWeekend ? (
            <View className="mb-4">
              <Text className="font-serif text-2xl font-semibold text-terracotta leading-tight mb-3">
                Es fin de semana de penitencias
              </Text>
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => router.push('/(tabs)/shame')}
                className="self-start rounded-full bg-terracotta px-5 py-2.5"
              >
                <Text className="text-paper text-sm font-bold">Ir a la ruleta →</Text>
              </TouchableOpacity>
            </View>
          ) : (
          <View className="flex-row items-baseline gap-2 mb-4">
            <View className="items-center">
              <Text className="font-serif text-4xl font-semibold text-terracotta leading-none">{countdown.days}</Text>
              <Text className="text-[10px] text-ink-faint mt-1">días</Text>
            </View>
            <Text className="font-serif text-3xl text-hairline mb-1">:</Text>
            <View className="items-center">
              <Text className="font-serif text-4xl font-semibold text-terracotta leading-none">{countdown.hours}</Text>
              <Text className="text-[10px] text-ink-faint mt-1">hrs</Text>
            </View>
            <Text className="font-serif text-3xl text-hairline mb-1">:</Text>
            <View className="items-center">
              <Text className="font-serif text-4xl font-semibold text-terracotta leading-none">{countdown.mins}</Text>
              <Text className="text-[10px] text-ink-faint mt-1">min</Text>
            </View>
          </View>
          )}

          {/* Progress bar */}
          <View className="flex-row justify-between text-xs mb-1.5">
            <Text className="text-ink-faint">Semana {weekNumber}</Text>
            <Text className="text-terracotta font-semibold">
              {stats.done} de {members.length + 1} al corriente
            </Text>
          </View>
          <View className="h-2 rounded-full bg-hairline overflow-hidden">
            <View className="h-full rounded-full bg-terracotta" style={{ width: `${progressPct}%` }} />
          </View>
        </View>

        {/* Tu Turno Card */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm">
          <View className="flex-row justify-between items-start mb-1">
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mt-1">TU TURNO</Text>
            <View className="items-end">
              <Text className="font-serif text-2xl font-semibold leading-none text-terracotta">{myStreak}</Text>
              <Text className="text-[9px] font-bold text-terracotta tracking-wider uppercase mt-0.5">RACHA</Text>
            </View>
          </View>

          <Text className="font-serif text-2xl font-semibold text-ink mb-3 leading-tight">
            ¿Ya hiciste el tuyo?
          </Text>

          {/* Habit list */}
          {myCheckins.length > 0 ? (
            <View className="gap-2.5 mb-4">
              {myCheckins.map((c) => (
                <View key={c.habit_id}>
                  <View className="flex-row items-center gap-2 flex-wrap">
                    <Text className="text-sm font-semibold text-ink">{c.habit_name}</Text>
                    {c.scheduled_time && (
                      <View className="rounded-full bg-hairline px-2 py-0.5">
                        <Text className="text-[10px] text-ink-soft font-medium">{c.scheduled_time}</Text>
                      </View>
                    )}
                    <View className={`rounded-full px-2 py-0.5 ${STATUS_PILL[c.status]?.cls || 'bg-hairline'}`}>
                      <Text className="text-[10px] font-bold">{STATUS_PILL[c.status]?.label || c.status}</Text>
                    </View>
                  </View>
                  {c.note ? (
                    <Text className="text-xs text-ink-soft italic mt-0.5">“{c.note}”</Text>
                  ) : null}
                </View>
              ))}
            </View>
          ) : (
            <Text className="text-sm text-ink-soft mb-4">
              No tienes un hábito registrado en este grupo todavía.
            </Text>
          )}

          {/* Action button */}
          {(hasPending || myCheckins.length === 0) ? (
            <TouchableOpacity
              activeOpacity={0.8}
              onPress={() => router.push('/(modals)/checkin')}
              className="w-full rounded-full bg-sage-deep py-3.5 items-center shadow-sm"
            >
              <Text className="text-paper font-bold text-sm">Hacer mi check-in →</Text>
            </TouchableOpacity>
          ) : allDone ? (
            <View className="w-full rounded-full bg-sage-soft py-3.5 items-center">
              <Text className="text-sage-deep font-bold text-sm">✓ Ya cumpliste hoy</Text>
            </View>
          ) : anyMissed ? (
            <View className="w-full rounded-full bg-coral-soft py-3.5 items-center">
              <Text className="text-coral-deep font-bold text-sm">Se te fue el día</Text>
            </View>
          ) : null}
        </View>

        {/* Summary Numbers */}
        <View className="rounded-3xl bg-paper border border-hairline flex-row text-center mb-6 overflow-hidden shadow-sm">
          <View className="flex-1 py-4 items-center justify-center">
            <Text className="font-serif text-2xl font-semibold text-sage-deep">{stats.done}</Text>
            <Text className="text-[10px] text-ink-soft mt-0.5">cumplieron</Text>
          </View>
          <View className="w-[1px] bg-hairline" />
          <View className="flex-1 py-4 items-center justify-center">
            <Text className="font-serif text-2xl font-semibold text-amber-deep">{stats.pending}</Text>
            <Text className="text-[10px] text-ink-soft mt-0.5">pendientes</Text>
          </View>
          <View className="w-[1px] bg-hairline" />
          <View className="flex-1 py-4 items-center justify-center">
            <Text className="font-serif text-2xl font-semibold text-coral-deep">{stats.missed}</Text>
            <Text className="text-[10px] text-ink-soft mt-0.5">fallaron</Text>
          </View>
        </View>

        {/* Squad list */}
        <View className="mb-4">
          <View className="flex-row justify-between items-center mb-3">
            <Text className="font-bold text-sm text-ink">El squad hoy</Text>
            <Text className="text-xs text-ink-soft">Lun → Dom</Text>
          </View>

          {squadMembers.length === 0 ? (
            <View className="rounded-3xl bg-surface border border-hairline py-8 px-4 items-center justify-center">
              <Text className="text-sm text-ink-soft text-center leading-snug">
                Propón un hábito en la pestaña Proposals para ver al squad aquí.
              </Text>
            </View>
          ) : (
            <View className="gap-3">
              {squadMembers.map((member) => (
                <View key={member.user_id} className="rounded-3xl bg-surface border border-hairline p-4">
                  {/* Member header (once per member) */}
                  <View className="flex-row items-center gap-3 mb-3">
                    <View className={`w-10 h-10 rounded-full flex-shrink-0 items-center justify-center ${getAvatarBg(member.display_name || '')}`}>
                      <Text className="text-paper text-sm font-bold">{getInitials(member.display_name || '')}</Text>
                    </View>
                    <View className="flex-1 min-w-0 flex-row justify-between items-center">
                      <View className="flex-row items-baseline gap-1.5 max-w-[70%]">
                        <Text className="font-semibold text-sm text-ink truncate">{member.display_name}</Text>
                        <Text className="text-xs text-terracotta font-medium">★ {streaks[member.user_id] ?? 0}</Text>
                      </View>
                      {STATUS_PILL[memberStatus(member.habits)] && (
                        <View className={`rounded-full px-2 py-0.5 ${STATUS_PILL[memberStatus(member.habits)].cls}`}>
                          <Text className="text-[9px] font-bold">{STATUS_PILL[memberStatus(member.habits)].label}</Text>
                        </View>
                      )}
                    </View>
                  </View>

                  {/* One block per assigned habit */}
                  <View className="gap-3">
                    {member.habits.map((row) => (
                      <View key={row.habit_id}>
                        <View className="flex-row justify-between items-center mb-1">
                          <Text className="text-xs text-ink-soft truncate flex-1">{row.habit_name}</Text>
                          {STATUS_PILL[row.status] && (
                            <View className={`rounded-full px-2 py-0.5 ml-2 ${STATUS_PILL[row.status].cls}`}>
                              <Text className="text-[9px] font-bold">{STATUS_PILL[row.status].label}</Text>
                            </View>
                          )}
                        </View>

                        {/* 7-day strip */}
                        <View className="flex-row gap-1">
                          {getMemberDayStrip(row).map((day, i) => (
                            <View key={i} className="items-center gap-1">
                              <View className={`w-7 h-7 rounded-lg items-center justify-center ${STATUS_STYLE[day.status].strip}`}>
                                {STATUS_STYLE[day.status].icon ? (
                                  <Text className={`text-xs font-bold ${STATUS_STYLE[day.status].iconColor}`}>
                                    {STATUS_STYLE[day.status].icon}
                                  </Text>
                                ) : null}
                              </View>
                              <Text className="text-[9px] text-ink-faint font-medium">{day.label}</Text>
                            </View>
                          ))}
                        </View>
                      </View>
                    ))}
                  </View>
                </View>
              ))}
            </View>
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
