import React, { useState } from 'react';
import { View, Text, ScrollView, TouchableOpacity, Share, Clipboard } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp, type Checkin } from '../../src/contexts/AppContext';
import { mondayIndex } from '../../src/weekStatus';

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

// ── Matriz semanal L–V ───────────────────────────────────────────────────────
// Espejo de la vista Squad de la web. Solo L–V porque eso es lo que juzga la
// ruleta; el fin de semana es su fin de semana.
const MATRIX_LABELS = ['L', 'M', 'M', 'J', 'V'];

type CellStatus = 'done' | 'partial' | 'pending' | 'missed' | 'future' | 'untracked';

const MATRIX_CELL: Record<CellStatus, string> = {
  done: 'bg-sage',
  partial: 'bg-sage/50',
  pending: 'bg-amber',
  missed: 'bg-coral',
  future: 'border border-dashed border-hairline',
  untracked: 'bg-hairline/40',
};

const MATRIX_TEXT: Record<CellStatus, string> = {
  done: 'text-sage-deep',
  partial: 'text-sage-deep',
  pending: 'text-amber-deep',
  missed: 'text-coral-deep',
  future: 'text-transparent',
  untracked: 'text-transparent',
};

const MATRIX_ICON: Partial<Record<CellStatus, string>> = {
  done: '✓',
  partial: '~',
  missed: '✗',
};

type SquadRow = {
  user_id: string;
  display_name: string;
  habits: Checkin[];
};

function dateForIdx(i: number, todayIdx: number): string {
  const d = new Date();
  d.setDate(d.getDate() - (todayIdx - i));
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(
    d.getDate()
  ).padStart(2, '0')}`;
}

function memberStatus(row: SquadRow): CellStatus {
  if (row.habits.every((h) => h.status === 'done')) return 'done';
  if (row.habits.some((h) => h.status === 'pending')) return 'pending';
  return 'missed';
}

function weekMatrixRow(row: SquadRow, weekHistory: Record<string, Set<string>>): CellStatus[] {
  const todayIdx = mondayIndex();
  return MATRIX_LABELS.map((_, i) => {
    if (i > todayIdx) return 'future';
    if (i === todayIdx) return memberStatus(row);
    const date = dateForIdx(i, todayIdx);
    // Un miembro puede tener hábitos asignados en fechas distintas: ese día solo
    // se juzgan los que ya contaban. Si ninguno contaba, el día no se juzga.
    const tracked = row.habits.filter((h) => !h.tracked_since || date >= h.tracked_since);
    if (!tracked.length) return 'untracked';
    let done = 0;
    for (const h of tracked) {
      if (weekHistory[`${row.user_id}:${h.habit_id}`]?.has(date)) done++;
    }
    if (done === tracked.length) return 'done';
    return done > 0 ? 'partial' : 'missed';
  });
}

// Semana perfecta: todos los días L–V transcurridos (incluido hoy) en verde.
function isPerfectWeek(cells: CellStatus[]): boolean {
  return (
    cells.some((c) => c === 'done') &&
    cells.every((c) => c === 'done' || c === 'future' || c === 'untracked')
  );
}

export default function SquadScreen() {
  const { user } = useAuth();
  const { group, members, onlineMembers, todayCheckins, weekHistory, streaks, propose } = useApp();

  const [copied, setCopied] = useState(false);
  const [confirmKick, setConfirmKick] = useState<string | null>(null);
  const [kicking, setKicking] = useState<string | null>(null);
  const [kickMsg, setKickMsg] = useState('');

  const displayUserName =
    user?.user_metadata?.display_name || user?.email?.split('@')[0] || 'Tú';

  // Filas del squad agrupadas por miembro — me incluye, así que mi propia
  // semana también sale aquí.
  const squadRows: SquadRow[] = Object.values(
    todayCheckins.reduce<Record<string, SquadRow>>((acc, c) => {
      if (!acc[c.user_id]) {
        acc[c.user_id] = { user_id: c.user_id, display_name: c.display_name, habits: [] };
      }
      acc[c.user_id].habits.push(c);
      return acc;
    }, {})
  ).sort((a, b) => a.display_name.localeCompare(b.display_name, 'es'));

  async function handleKick(member: { user_id: string; display_name: string }) {
    if (confirmKick !== member.user_id) {
      setConfirmKick(member.user_id);
      return;
    }
    if (!group?.id) return;
    setKicking(member.user_id);
    setKickMsg('');
    try {
      await propose(group.id, 'kick_member', null, member.user_id);
      setKickMsg(`Propuesta creada: expulsar a ${member.display_name}. El squad vota en "Votar".`);
    } catch (e: any) {
      setKickMsg(e?.error ?? 'No se pudo crear la propuesta');
    } finally {
      setKicking(null);
      setConfirmKick(null);
    }
  }

  function copyInviteCode() {
    if (!group) return;
    const code = `${group.id}:${group.invite_code}`;
    try {
      Clipboard.setString(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.warn('Could not copy to clipboard:', err);
    }
  }

  async function shareInvite() {
    if (!group) return;
    const code = `${group.id}:${group.invite_code}`;
    const text = `¡Únete a mi squad "${group.name}" en Dydi!\nCódigo de invitación: ${code}`;
    try {
      await Share.share({ message: text, title: 'Únete a Dydi' });
    } catch (e) {
      console.error(e);
    }
  }

  return (
    <SafeAreaView className="flex-1 bg-cream" edges={['top']}>
      <View className="px-6 py-4 border-b border-hairline/30 bg-cream">
        <View className="flex-row items-baseline justify-between">
          <Text className="font-serif text-2xl font-semibold text-ink">Squad</Text>
          <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
            {group?.name ?? ''}
          </Text>
        </View>
        <Text className="text-xs text-ink-soft mt-0.5">
          La semana del equipo · presencia en vivo
        </Text>
      </View>

      <ScrollView className="flex-1 px-6 py-4" showsVerticalScrollIndicator={false}>
        {!group ? (
          <View className="rounded-3xl bg-paper border border-hairline py-10 items-center">
            <Text className="text-sm text-ink-soft">Todavía no estás en un squad.</Text>
          </View>
        ) : (
          <>
            {/* ── La semana del squad: miembros × L–V ───────────────────────── */}
            <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm">
              <View className="flex-row items-center justify-between mb-4">
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
                  LA SEMANA DEL SQUAD
                </Text>
                <Text className="text-[9px] text-ink-faint">L–V cuentan</Text>
              </View>

              {!squadRows.length ? (
                <Text className="text-xs text-ink-soft py-4 text-center">
                  Ningún miembro tiene hábitos asignados todavía.
                </Text>
              ) : (
                <>
                  {/* Encabezado de días */}
                  <View className="flex-row items-center mb-1">
                    <View className="flex-1" />
                    <View className="flex-row gap-1">
                      {MATRIX_LABELS.map((l, i) => (
                        <Text
                          key={i}
                          className="w-7 text-center text-[9px] font-medium text-ink-faint"
                        >
                          {l}
                        </Text>
                      ))}
                    </View>
                  </View>

                  <View className="gap-1">
                    {squadRows.map((row) => {
                      const cells = weekMatrixRow(row, weekHistory);
                      const perfect = isPerfectWeek(cells);
                      return (
                        <View
                          key={row.user_id}
                          className={`flex-row items-center rounded-2xl px-2 py-1.5 ${
                            perfect ? 'bg-amber-soft/60' : ''
                          }`}
                        >
                          <View
                            className={`w-7 h-7 rounded-full items-center justify-center ${getAvatarBg(row.display_name)}`}
                          >
                            <Text className="text-paper text-[10px] font-bold">
                              {getInitials(row.display_name)}
                            </Text>
                          </View>

                          <View className="flex-1 min-w-0 px-2">
                            <Text className="text-xs font-semibold text-ink" numberOfLines={1}>
                              {row.user_id === user?.id ? 'Tú' : row.display_name}
                            </Text>
                            {perfect ? (
                              <Text className="text-[9px] font-bold text-amber-deep">
                                semana perfecta
                              </Text>
                            ) : null}
                          </View>

                          <View className="flex-row gap-1">
                            {cells.map((status, i) => (
                              <View
                                key={i}
                                className={`w-7 h-7 rounded-lg items-center justify-center ${MATRIX_CELL[status]}`}
                              >
                                <Text className={`text-[10px] font-bold ${MATRIX_TEXT[status]}`}>
                                  {MATRIX_ICON[status] ?? ''}
                                </Text>
                              </View>
                            ))}
                          </View>
                        </View>
                      );
                    })}
                  </View>
                </>
              )}
            </View>

            {/* ── Miembros ──────────────────────────────────────────────────── */}
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">
              MIEMBROS DEL SQUAD ({members.filter((m) => m.user_id !== user?.id).length + 1})
            </Text>

            <View className="gap-2 mb-6">
              {/* Yo */}
              <View className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3">
                <View
                  className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(displayUserName)}`}
                >
                  <Text className="text-paper text-sm font-bold">
                    {getInitials(displayUserName)}
                  </Text>
                </View>
                <View className="flex-1 min-w-0">
                  <Text className="font-semibold text-sm text-ink" numberOfLines={1}>
                    {displayUserName} (Tú)
                  </Text>
                  <Text className="text-xs text-ink-soft mt-0.5" numberOfLines={1}>
                    {user?.email}
                  </Text>
                </View>
                <Text className="text-xs font-medium text-terracotta">
                  ★ {streaks[user?.id ?? ''] ?? 0}
                </Text>
              </View>

              {/* Los demás */}
              {members
                .filter((m) => m.user_id !== user?.id)
                .map((member) => (
                  <View
                    key={member.user_id}
                    className={`rounded-3xl bg-paper border p-4 flex-row items-center gap-3 ${
                      onlineMembers.has(member.user_id) ? 'border-sage/50' : 'border-hairline'
                    }`}
                  >
                    <View className="relative">
                      <View
                        className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(member.display_name)}`}
                      >
                        <Text className="text-paper text-sm font-bold">
                          {getInitials(member.display_name)}
                        </Text>
                      </View>
                      {onlineMembers.has(member.user_id) ? (
                        <View className="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-sage-deep border-2 border-paper" />
                      ) : null}
                    </View>

                    <View className="flex-1 min-w-0">
                      <Text className="font-semibold text-sm text-ink" numberOfLines={1}>
                        {member.display_name}
                      </Text>
                      <Text className="text-xs font-medium text-terracotta mt-0.5">
                        ★ {streaks[member.user_id] ?? 0}
                      </Text>
                    </View>

                    <TouchableOpacity
                      activeOpacity={0.8}
                      disabled={kicking === member.user_id}
                      onPress={() => handleKick(member)}
                      className={`rounded-full px-3 py-1.5 border ${
                        confirmKick === member.user_id
                          ? 'bg-coral-deep border-coral-deep'
                          : 'bg-paper border-hairline'
                      }`}
                    >
                      <Text
                        className={`text-[10px] font-bold ${
                          confirmKick === member.user_id ? 'text-paper' : 'text-ink-faint'
                        }`}
                      >
                        {kicking === member.user_id
                          ? '…'
                          : confirmKick === member.user_id
                            ? '¿Proponer expulsión?'
                            : 'Expulsar'}
                      </Text>
                    </TouchableOpacity>
                  </View>
                ))}
            </View>

            {kickMsg ? (
              <View className="rounded-3xl bg-amber-soft/40 border border-amber/30 px-4 py-3 mb-5">
                <Text className="text-xs font-semibold text-amber-deep">{kickMsg}</Text>
              </View>
            ) : null}

            {/* ── Invitar ───────────────────────────────────────────────────── */}
            <View className="rounded-3xl bg-paper border border-hairline p-5 mb-8 shadow-sm">
              <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3">
                INVITAR AL SQUAD
              </Text>

              <View className="rounded-2xl bg-cream-2 border border-hairline/60 p-4 mb-4">
                <Text className="text-[9px] font-bold text-ink-soft tracking-wider uppercase mb-1">
                  CÓDIGO DE INVITACIÓN
                </Text>
                <Text className="font-mono text-xs text-ink leading-normal">
                  {group.id}:{group.invite_code}
                </Text>
              </View>

              <View className="flex-row gap-2">
                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={copyInviteCode}
                  className={`flex-1 rounded-full border py-3 items-center ${
                    copied ? 'bg-sage-soft/30 border-sage/40' : 'bg-surface border-hairline'
                  }`}
                >
                  <Text
                    className={`font-bold text-xs ${copied ? 'text-sage-deep' : 'text-ink-soft'}`}
                  >
                    {copied ? '¡Copiado! ✓' : 'Copiar código'}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={shareInvite}
                  className="flex-1 rounded-full bg-terracotta py-3 items-center justify-center"
                >
                  <Text className="text-paper font-bold text-xs">Compartir</Text>
                </TouchableOpacity>
              </View>
            </View>
          </>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}
