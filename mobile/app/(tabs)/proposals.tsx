import React, { useEffect, useState, useMemo } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { useApp } from '../../src/contexts/AppContext';
import HabitIcon from '../../src/components/ui/HabitIcon';

const PROPOSAL_LABEL: Record<string, string> = {
  add_habit: 'Agregar hábito',
  remove_habit: 'Quitar hábito',
  kick_member: 'Expulsar miembro',
  delete_group: 'Disolver grupo',
};

export default function ProposalsScreen() {
  const router = useRouter();
  const {
    group,
    members,
    catalog,
    proposals,
    resolvedProposals,
    voted,
    todayCheckins,
    loadCatalog,
    loadProposals,
    loadResolvedProposals,
    propose,
    vote,
  } = useApp();

  const [tab, setTab] = useState<'catalogo' | 'propuestas' | 'historial'>('catalogo');
  const [loading, setLoading] = useState(true);
  const [historyLoaded, setHistoryLoaded] = useState(false);
  const [proposingID, setProposingID] = useState<string | null>(null);
  const [proposeErr, setProposeErr] = useState('');
  const [proposeOkID, setProposeOkID] = useState<string | null>(null);
  const [votingID, setVotingID] = useState<string | null>(null);
  const [voteErr, setVoteErr] = useState('');

  // Auto-load data
  useEffect(() => {
    async function loadData() {
      if (!group?.id) {
        setLoading(false);
        return;
      }
      setLoading(true);
      try {
        await Promise.all([loadCatalog(), loadProposals(group.id)]);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [group?.id]);

  // Identify assigned habit IDs
  const assignedHabitIDs = useMemo(() => {
    return new Set(todayCheckins.map((c) => c.habit_id));
  }, [todayCheckins]);

  // Split catalog into available and active
  const availableHabits = useMemo(() => {
    return catalog.filter((h) => !assignedHabitIDs.has(h.id));
  }, [catalog, assignedHabitIDs]);

  const activeHabits = useMemo(() => {
    return catalog.filter((h) => assignedHabitIDs.has(h.id));
  }, [catalog, assignedHabitIDs]);

  function getHabitName(habitID: string) {
    return catalog.find((h) => h.id === habitID)?.name ?? habitID;
  }

  function getMemberName(userID: string) {
    return members.find((m) => m.user_id === userID)?.display_name ?? 'Miembro';
  }

  const STATUS_BADGE: Record<string, { label: string; bg: string; text: string }> = {
    approved: { label: 'APROBADA', bg: 'bg-sage-soft', text: 'text-sage-deep' },
    rejected: { label: 'RECHAZADA', bg: 'bg-coral-soft', text: 'text-coral-deep' },
    expired: { label: 'EXPIRÓ', bg: 'bg-cream-2', text: 'text-ink-faint' },
  };

  // Carga perezosa: el historial solo se pide al abrir su tab.
  async function openHistory() {
    setTab('historial');
    if (historyLoaded || !group?.id) return;
    try {
      await loadResolvedProposals(group.id);
      setHistoryLoaded(true);
    } catch (err) {
      console.error(err);
    }
  }

  function getVoteProgress(p: any) {
    if (!p.member_count) return 0;
    return Math.round((p.vote_count / p.member_count) * 100);
  }

  function getQuorumLabel(p: any) {
    const need = Math.ceil(p.member_count / 2);
    return `${p.vote_count} de ${need} votos necesarios`;
  }

  async function handlePropose(habit: any, type: 'add_habit' | 'remove_habit') {
    if (!group?.id || proposingID || proposeOkID === habit.id) return;
    setProposingID(habit.id);
    setProposeErr('');
    try {
      await propose(group.id, type, habit.id);
      setProposeOkID(habit.id);
      setTab('propuestas');
    } catch (e: any) {
      setProposeErr(e?.error ?? e?.message ?? 'No se pudo crear la propuesta.');
    } finally {
      setProposingID(null);
    }
  }

  async function castVote(proposalID: string, approved: boolean) {
    setVotingID(proposalID);
    setVoteErr('');
    try {
      await vote(proposalID, approved);
    } catch (e: any) {
      setVoteErr(e?.error ?? e?.message ?? 'No se pudo registrar el voto.');
    } finally {
      setVotingID(null);
    }
  }

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center bg-cream">
        <ActivityIndicator size="large" color="#7CA39D" />
      </View>
    );
  }

  if (!group) {
    return (
      <SafeAreaView className="flex-1 bg-cream items-center justify-center px-6">
        <Text className="text-sm text-ink-soft mb-4">No estás asociado a ningún grupo.</Text>
        <TouchableOpacity
          activeOpacity={0.8}
          onPress={() => router.replace('/onboarding')}
          className="rounded-full bg-sage-deep px-6 py-2.5"
        >
          <Text className="text-paper font-bold text-xs">Crear o Unirme</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-cream" edges={['top']}>
      {/* Header */}
      <View className="px-6 py-4 border-b border-hairline/30 bg-cream">
        <View className="flex-row items-center justify-between">
          <Text className="font-serif text-2xl font-semibold text-ink">Propuestas</Text>
          <Text className="text-xs font-bold text-ink-soft uppercase tracking-wider">{group.name}</Text>
        </View>
        <Text className="text-xs text-ink-faint mt-0.5">Propón y vota hábitos para el squad</Text>
      </View>

      <ScrollView className="flex-1 px-6 py-4" showsVerticalScrollIndicator={false}>
        {/* Tab switcher */}
        <View className="flex-row bg-cream-2 rounded-[14px] p-1 mb-5">
          <TouchableOpacity
            activeOpacity={0.8}
            onPress={() => setTab('catalogo')}
            className={`flex-1 py-2 rounded-[10px] items-center ${tab === 'catalogo' ? 'bg-paper shadow-sm' : ''}`}
          >
            <Text className={`text-sm font-semibold ${tab === 'catalogo' ? 'text-ink' : 'text-ink-soft'}`}>
              Catálogo
            </Text>
          </TouchableOpacity>

          <TouchableOpacity
            activeOpacity={0.8}
            onPress={() => setTab('propuestas')}
            className={`flex-1 py-2 rounded-[10px] items-center relative ${tab === 'propuestas' ? 'bg-paper shadow-sm' : ''}`}
          >
            <Text className={`text-sm font-semibold ${tab === 'propuestas' ? 'text-ink' : 'text-ink-soft'}`}>
              Propuestas
            </Text>
            {proposals.length > 0 && (
              <View className="absolute top-1 right-3 w-4 h-4 rounded-full bg-terracotta items-center justify-center">
                <Text className="text-paper text-[8px] font-bold">{proposals.length}</Text>
              </View>
            )}
          </TouchableOpacity>

          <TouchableOpacity
            activeOpacity={0.8}
            onPress={openHistory}
            className={`flex-1 py-2 rounded-[10px] items-center ${tab === 'historial' ? 'bg-paper shadow-sm' : ''}`}
          >
            <Text className={`text-sm font-semibold ${tab === 'historial' ? 'text-ink' : 'text-ink-soft'}`}>
              Historial
            </Text>
          </TouchableOpacity>
        </View>

        {/* Catalog Tab */}
        {tab === 'catalogo' && (
          <View className="gap-5">
            <Text className="text-xs text-ink-soft px-1">
              Propón un hábito para todo el squad. Requiere votación mayoritaria.
            </Text>

            {proposeErr ? (
              <Text className="text-sm text-coral-deep font-semibold px-1">{proposeErr}</Text>
            ) : null}

            {/* Available to Add */}
            {availableHabits.length > 0 && (
              <View>
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">DISPONIBLES PARA AÑADIR</Text>
                <View className="gap-2">
                  {availableHabits.map((habit) => (
                    <View key={habit.id} className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3 shadow-sm">
                      <View
                        className="w-10 h-10 rounded-full items-center justify-center"
                        style={{ backgroundColor: habit.color || '#A8C39A' }}
                      >
                        <HabitIcon iconKey={habit.icon_key} size={22} color="#FFFFFF" />
                      </View>

                      <View className="flex-1 min-w-0">
                        <Text className="font-semibold text-sm text-ink truncate">{habit.name}</Text>
                        {habit.description && (
                          <Text className="text-xs text-ink-soft truncate mt-0.5">{habit.description}</Text>
                        )}
                      </View>

                      {proposeOkID === habit.id ? (
                        <View className="bg-sage-soft rounded-full px-3 py-1.5">
                          <Text className="text-[10px] font-bold text-sage-deep">✓ Propuesto</Text>
                        </View>
                      ) : (
                        <TouchableOpacity
                          disabled={proposingID === habit.id}
                          activeOpacity={0.8}
                          onPress={() => handlePropose(habit, 'add_habit')}
                          className="rounded-full border border-hairline bg-surface px-3 py-1.5"
                        >
                          {proposingID === habit.id ? (
                            <ActivityIndicator size="small" color="#6F6557" />
                          ) : (
                            <Text className="text-ink-soft text-xs font-bold">+ Proponer</Text>
                          )}
                        </TouchableOpacity>
                      )}
                    </View>
                  ))}
                </View>
              </View>
            )}

            {/* Active Habits */}
            {activeHabits.length > 0 && (
              <View className="mb-6">
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">ACTIVOS EN EL SQUAD</Text>
                <View className="gap-2">
                  {activeHabits.map((habit) => (
                    <View key={habit.id} className="rounded-3xl bg-surface border border-hairline p-4 flex-row items-center gap-3">
                      <View
                        className="w-10 h-10 rounded-full items-center justify-center opacity-60"
                        style={{ backgroundColor: habit.color || '#A8C39A' }}
                      >
                        <HabitIcon iconKey={habit.icon_key} size={22} color="#FFFFFF" />
                      </View>

                      <View className="flex-1 min-w-0">
                        <Text className="font-semibold text-sm text-ink truncate">{habit.name}</Text>
                        <View className="bg-sage-soft rounded-full px-2 py-0.5 self-start mt-1">
                          <Text className="text-[9px] font-semibold text-sage-deep">Ya en el grupo</Text>
                        </View>
                      </View>

                      {proposeOkID === habit.id ? (
                        <View className="bg-sage-soft rounded-full px-3 py-1.5">
                          <Text className="text-[10px] font-bold text-sage-deep">✓ Propuesto</Text>
                        </View>
                      ) : (
                        <TouchableOpacity
                          disabled={proposingID === habit.id}
                          activeOpacity={0.8}
                          onPress={() => handlePropose(habit, 'remove_habit')}
                          className="rounded-full border border-coral/30 bg-coral-soft/50 px-3 py-1.5"
                        >
                          {proposingID === habit.id ? (
                            <ActivityIndicator size="small" color="#BC5C42" />
                          ) : (
                            <Text className="text-coral-deep text-xs font-bold">- Quitar</Text>
                          )}
                        </TouchableOpacity>
                      )}
                    </View>
                  ))}
                </View>
              </View>
            )}
          </View>
        )}

        {/* Proposals Tab */}
        {tab === 'propuestas' && (
          <View className="mb-8">
            {proposals.length === 0 ? (
              <View className="rounded-3xl bg-paper border border-hairline py-12 items-center justify-center shadow-sm">
                <Text className="text-[10px] font-bold text-ink-faint tracking-wider uppercase mb-2">SIN PROPUESTAS</Text>
                <Text className="font-serif text-xl font-semibold text-ink mb-1">Todo tranquilo</Text>
                <Text className="text-xs text-ink-soft text-center mt-1">Propón un hábito desde el catálogo.</Text>
              </View>
            ) : (
              <View className="gap-3">
                {voteErr ? (
                  <Text className="text-sm text-coral-deep font-medium px-1">{voteErr}</Text>
                ) : null}

                {proposals.map((p) => {
                  const hasUserVoted = voted.has(p.id);
                  const progress = getVoteProgress(p);

                  return (
                    <View key={p.id} className="rounded-3xl bg-paper border border-hairline p-5 shadow-sm">
                      <View className="flex-row items-start justify-between gap-2 mb-3">
                        <View className="flex-1">
                          <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
                            {PROPOSAL_LABEL[p.type] ?? p.type}
                          </Text>
                          {p.habit_id ? (
                            <Text className="font-semibold text-sm text-ink mt-0.5">
                              {getHabitName(p.habit_id)}
                            </Text>
                          ) : p.target_user_id ? (
                            <Text className="font-semibold text-sm text-ink mt-0.5">
                              {getMemberName(p.target_user_id)}
                            </Text>
                          ) : null}
                        </View>
                        <View className="rounded-full bg-amber-soft px-2.5 py-0.5">
                          <Text className="text-[9px] font-bold text-amber-deep">ABIERTA</Text>
                        </View>
                      </View>

                      {/* Progress Bar */}
                      <View className="mb-4">
                        <View className="flex-row justify-between text-xs text-ink-soft mb-1">
                          <Text className="text-xs text-ink-soft">{getQuorumLabel(p)}</Text>
                          <Text className="text-xs text-ink-soft font-bold">{progress}%</Text>
                        </View>
                        <View className="h-1.5 rounded-full bg-hairline overflow-hidden">
                          <View className="h-full rounded-full bg-sage-deep" style={{ width: `${progress}%` }} />
                        </View>
                      </View>

                      {/* Vote actions */}
                      {!hasUserVoted ? (
                        <View className="flex-row gap-2">
                          <TouchableOpacity
                            disabled={votingID === p.id}
                            activeOpacity={0.8}
                            onPress={() => castVote(p.id, true)}
                            className="flex-1 rounded-full bg-sage-deep py-2.5 items-center justify-center"
                          >
                            {votingID === p.id ? (
                              <ActivityIndicator size="small" color="#FFFFFF" />
                            ) : (
                              <Text className="text-paper font-bold text-xs">✓ Aprobar</Text>
                            )}
                          </TouchableOpacity>

                          <TouchableOpacity
                            disabled={votingID === p.id}
                            activeOpacity={0.8}
                            onPress={() => castVote(p.id, false)}
                            className="flex-1 rounded-full border border-hairline bg-paper py-2.5 items-center justify-center"
                          >
                            <Text className="text-ink-soft font-bold text-xs">✗ Rechazar</Text>
                          </TouchableOpacity>
                        </View>
                      ) : (
                        <View className="rounded-full bg-sage-soft py-2.5 items-center justify-center">
                          <Text className="text-sage-deep font-bold text-xs">✓ Ya votaste</Text>
                        </View>
                      )}
                    </View>
                  );
                })}
              </View>
            )}
          </View>
        )}

        {/* Historial Tab */}
        {tab === 'historial' && (
          <View className="mb-8">
            {resolvedProposals.length === 0 ? (
              <View className="rounded-3xl bg-paper border border-hairline py-12 items-center justify-center shadow-sm">
                <Text className="text-[10px] font-bold text-ink-faint tracking-wider uppercase mb-2">SIN HISTORIAL</Text>
                <Text className="font-serif text-xl font-semibold text-ink mb-1">Nada decidido aún</Text>
                <Text className="text-xs text-ink-soft text-center mt-1">Las propuestas cerradas aparecerán aquí.</Text>
              </View>
            ) : (
              <View className="gap-3">
                {resolvedProposals.map((p) => {
                  const badge = STATUS_BADGE[p.status ?? ''] ?? {
                    label: (p.status ?? '').toUpperCase(),
                    bg: 'bg-cream-2',
                    text: 'text-ink-faint',
                  };
                  return (
                    <View key={p.id} className="rounded-3xl bg-surface border border-hairline p-5">
                      <View className="flex-row items-start justify-between gap-2 mb-2">
                        <View className="flex-1">
                          <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">
                            {PROPOSAL_LABEL[p.type] ?? p.type}
                          </Text>
                          {p.habit_id ? (
                            <Text className="font-semibold text-sm text-ink mt-0.5">
                              {getHabitName(p.habit_id)}
                            </Text>
                          ) : p.target_user_id ? (
                            <Text className="font-semibold text-sm text-ink mt-0.5">
                              {getMemberName(p.target_user_id)}
                            </Text>
                          ) : null}
                        </View>
                        <View className={`rounded-full px-2.5 py-0.5 ${badge.bg}`}>
                          <Text className={`text-[9px] font-bold ${badge.text}`}>{badge.label}</Text>
                        </View>
                      </View>
                      <Text className="text-xs text-ink-soft">
                        {p.vote_count} de {p.member_count} votos a favor
                      </Text>
                    </View>
                  );
                })}
              </View>
            )}
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}
