import React, { useEffect, useState, useMemo, useRef } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  TextInput,
  Animated,
  Easing,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import Svg, { Path, Circle, G } from 'react-native-svg';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp } from '../../src/contexts/AppContext';

// Espeja spinGraceHours del backend: pasado el deadline + gracia, cualquier
// miembro puede girar por el deudor para que la ruleta nunca muera sin girar.
const SPIN_GRACE_MS = 24 * 3_600_000;

const WHEEL_COLORS = [
  '#C26F4D',
  '#A8C39A',
  '#5C7650',
  '#E9C281',
  '#EDA48F',
  '#BC5C42',
  '#7CA39D',
  '#A57B33',
];

const AVATAR_COLORS = [
  'bg-sage-deep',
  'bg-terracotta',
  'bg-sage',
  'bg-amber',
  'bg-coral',
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

function getShortDate(iso: string) {
  try {
    const d = new Date(iso);
    return d.toLocaleDateString('es-MX', { month: 'short', day: 'numeric' });
  } catch {
    return '';
  }
}

// Polar coordinates calculations for wheel segments
function polarToCartesian(cx: number, cy: number, r: number, deg: number) {
  const rad = ((deg - 90) * Math.PI) / 180;
  return [cx + r * Math.cos(rad), cx + r * Math.sin(rad)];
}

function segmentPath(cx: number, cy: number, r: number, start: number, end: number) {
  const [sx, sy] = polarToCartesian(cx, cy, r, start);
  const [ex, ey] = polarToCartesian(cx, cy, r, end);
  const large = end - start > 180 ? 1 : 0;
  return `M ${cx} ${cy} L ${sx} ${sy} A ${r} ${r} 0 ${large} 1 ${ex} ${ey} Z`;
}

export default function ShameScreen() {
  const router = useRouter();
  const { user } = useAuth();
  const {
    group,
    members,
    debts,
    resolvedDebts,
    eligible,
    openEntries,
    activeEntry,
    suggestions,
    loadDebts,
    loadResolvedDebts,
    loadEligible,
    loadOpenEntries,
    enterEntry,
    openRoulette,
    loadSuggestions,
    submitSuggestion,
    spin,
    completeDebt,
    clearEntry,
  } = useApp();

  const [view, setView] = useState<'list' | 'entry'>('list');
  const [loading, setLoading] = useState(false);
  const [spinning, setSpinning] = useState(false);
  const [spinResult, setSpinResult] = useState<any>(null);
  const [error, setError] = useState('');
  
  // Suggestion form
  const [showForm, setShowForm] = useState(false);
  const [sugText, setSugText] = useState('');
  const [sugEmoji, setSugEmoji] = useState('');

  // Complete-debt confirm state
  const [confirmComplete, setConfirmComplete] = useState<string | null>(null);
  const [completing, setCompleting] = useState<string | null>(null);

  // Historial de deudas (carga perezosa)
  const [showHistory, setShowHistory] = useState(false);
  const [historyLoaded, setHistoryLoaded] = useState(false);

  const DEBT_STATUS_BADGE: Record<string, { label: string; bg: string; text: string }> = {
    completed: { label: 'CUMPLIDA', bg: 'bg-sage-soft', text: 'text-sage-deep' },
    forgiven: { label: 'PERDONADA', bg: 'bg-amber-soft', text: 'text-amber-deep' },
    expired: { label: 'EXPIRÓ', bg: 'bg-cream-2', text: 'text-ink-faint' },
  };

  async function toggleHistory() {
    const next = !showHistory;
    setShowHistory(next);
    if (!next || historyLoaded || !group?.id) return;
    try {
      await loadResolvedDebts(group.id);
      setHistoryLoaded(true);
    } catch (err) {
      console.error(err);
    }
  }

  // Wheel animation refs
  const spinValueRef = useRef(new Animated.Value(0)).current;
  const currentAngleRef = useRef(0);

  // Load list data on mount
  useEffect(() => {
    async function loadData() {
      if (!group?.id) return;
      setLoading(true);
      try {
        await Promise.all([loadEligible(group.id), loadDebts(group.id), loadOpenEntries(group.id)]);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [group?.id]);

  // Derived entry states
  const deadlinePassed = useMemo(() => {
    if (!activeEntry) return false;
    return new Date() > new Date(activeEntry.suggestion_deadline);
  }, [activeEntry]);

  const isDebtor = useMemo(() => {
    return activeEntry?.debtor_id === user?.id;
  }, [activeEntry, user?.id]);

  const graceOver = useMemo(() => {
    if (!activeEntry) return false;
    return Date.now() > new Date(activeEntry.suggestion_deadline).getTime() + SPIN_GRACE_MS;
  }, [activeEntry]);

  const canSpin = useMemo(() => {
    return deadlinePassed && !activeEntry?.spun_at && (isDebtor || graceOver);
  }, [isDebtor, deadlinePassed, graceOver, activeEntry]);

  const hasSuggested = useMemo(() => {
    return suggestions.some((s) => s.suggester_id === user?.id);
  }, [suggestions, user?.id]);

  const canSuggest = useMemo(() => {
    return !deadlinePassed && !hasSuggested && !isDebtor;
  }, [deadlinePassed, hasSuggested, isDebtor]);

  const deadlineLabel = useMemo(() => {
    if (!activeEntry) return '';
    const diff = new Date(activeEntry.suggestion_deadline).getTime() - new Date().getTime();
    if (diff <= 0) return 'Cerrado';
    const hrs = Math.floor(diff / 3600000);
    const mins = Math.floor((diff % 3600000) / 60000);
    if (hrs >= 24) return `${Math.floor(hrs / 24)}d ${hrs % 24}h`;
    if (hrs > 0) return `${hrs}h ${mins}min`;
    return `${mins}min`;
  }, [activeEntry]);

  const debtorName = useMemo(() => {
    if (!activeEntry) return '';
    const found = members.find((m) => m.user_id === activeEntry.debtor_id);
    if (found) return found.display_name;
    if (activeEntry.debtor_id === user?.id) return user?.user_metadata?.display_name || 'Tú';
    return 'Miembro';
  }, [activeEntry, members, user]);

  const getMemberName = (id: string) => {
    if (id === user?.id) return user?.user_metadata?.display_name || 'Tú';
    return members.find((m) => m.user_id === id)?.display_name ?? 'Alguien';
  };

  // Miembros en el bote que aún no tienen ruleta abierta (los que ya la tienen
  // aparecen en la sección "Ruletas abiertas").
  const eligibleWithoutEntry = useMemo(
    () => eligible.filter((m) => !openEntries.some((e) => e.debtor_id === m.user_id)),
    [eligible, openEntries]
  );

  function entryCountdown(e: any) {
    const diff = new Date(e.suggestion_deadline).getTime() - Date.now();
    if (diff <= 0) return '¡Lista para girar!';
    const hrs = Math.floor(diff / 3600000);
    const mins = Math.floor((diff % 3600000) / 60000);
    if (hrs >= 24) return `Sugerencias por ${Math.floor(hrs / 24)}d ${hrs % 24}h`;
    if (hrs > 0) return `Sugerencias por ${hrs}h ${mins}min`;
    return `Sugerencias por ${mins}min`;
  }

  const entryDebtorName = (e: any) => e.debtor_name ?? getMemberName(e.debtor_id);

  const activeSuggestions = useMemo(() => {
    return suggestions.length >= 2 ? suggestions : Array.from({ length: 8 });
  }, [suggestions]);

  // Slices configuration for rendering Svg wheel
  const wheelSegments = useMemo(() => {
    const itemsCount = activeSuggestions.length;
    const anglePer = 360 / itemsCount;
    return activeSuggestions.map((item: any, i) => ({
      path: segmentPath(100, 100, 95, i * anglePer, (i + 1) * anglePer),
      color: WHEEL_COLORS[i % WHEEL_COLORS.length],
      label: item?.text ? item.text.slice(0, 12) : '',
    }));
  }, [activeSuggestions]);

  // Actions
  async function handleOpenRoulette(member: any) {
    if (!group?.id || loading) return;
    setLoading(true);
    setError('');
    try {
      const entry = await openRoulette(group.id, member.user_id);
      await loadSuggestions(entry.id);
      setView('entry');
    } catch (e: any) {
      setError(e?.error ?? 'No se pudo abrir la ruleta');
    } finally {
      setLoading(false);
    }
  }

  // Entra a una ruleta ya abierta sin re-abrirla (POST exige elegibilidad vigente).
  async function handleEnterEntry(entry: any) {
    if (loading) return;
    setLoading(true);
    setError('');
    try {
      enterEntry(entry);
      await loadSuggestions(entry.id);
      setView('entry');
    } catch (e: any) {
      setError(e?.error ?? 'No se pudo abrir la ruleta');
    } finally {
      setLoading(false);
    }
  }

  async function handleCompleteDebt(debt: any) {
    if (confirmComplete !== debt.id) {
      setConfirmComplete(debt.id);
      return;
    }
    setCompleting(debt.id);
    setError('');
    try {
      await completeDebt(debt.id);
    } catch (e: any) {
      setError(e?.error ?? 'No se pudo marcar la deuda');
    } finally {
      setCompleting(null);
      setConfirmComplete(null);
    }
  }

  async function handleSuggest() {
    const text = sugText.trim();
    if (!text || !activeEntry) return;
    setError('');
    try {
      await submitSuggestion(activeEntry.id, text, sugEmoji.trim() || null);
      setSugText('');
      setSugEmoji('');
      setShowForm(false);
    } catch (e: any) {
      setError(e?.error ?? 'No se pudo enviar la sugerencia');
    }
  }

  async function handleSpin() {
    if (spinning || !activeEntry) return;
    setSpinning(true);
    setError('');

    let result: any = null;
    try {
      result = await spin(activeEntry.id);
    } catch (e: any) {
      setError(e?.error ?? 'Error al girar la ruleta');
      setSpinning(false);
      return;
    }

    // Determine target index
    const items = suggestions.length >= 2 ? suggestions : Array.from({ length: 8 });
    let winnerIndex = items.findIndex((s: any) => s?.id === result.winning_suggestion_id);
    if (winnerIndex === -1) {
      winnerIndex = Math.floor(Math.random() * items.length);
    }

    // Calculate rotation angle
    const anglePer = 360 / items.length;
    const centerDeg = (winnerIndex + 0.5) * anglePer;
    // Slight random offset inside slice to make it feel natural
    const offset = (Math.random() - 0.5) * (anglePer * 0.8);
    const targetLanding = centerDeg + offset;

    // Apply multiple full turns for effect
    const extraRotations = 5;
    const totalAngle = extraRotations * 360 + (360 - targetLanding);
    
    currentAngleRef.current = totalAngle;

    // Run the animation
    Animated.timing(spinValueRef, {
      toValue: totalAngle,
      duration: 4200,
      easing: Easing.bezier(0.17, 0.67, 0.12, 0.99),
      useNativeDriver: true,
    }).start(() => {
      setSpinResult(result);
      setSpinning(false);
    });
  }

  function handleBack() {
    setView('list');
    setSpinResult(null);
    setShowForm(false);
    setError('');
    spinValueRef.setValue(0);
    currentAngleRef.current = 0;
    clearEntry();
    if (group?.id) {
      loadEligible(group.id);
      loadDebts(group.id);
      loadOpenEntries(group.id);
    }
  }

  // Interpolation for rotating the SVG wheel
  const rotateInterpolate = spinValueRef.interpolate({
    inputRange: [0, 360],
    outputRange: ['0deg', '360deg'],
  });

  if (loading && view === 'list') {
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
      {/* ═══════════════════ LIST VIEW ═══════════════════════════════════════ */}
      {view === 'list' && (
        <>
          {/* Header */}
          <View className="px-6 py-4 border-b border-hairline/30 bg-cream">
            <View className="flex-row items-center justify-between">
              <Text className="font-serif text-2xl font-semibold text-ink">Ruleta</Text>
              <Text className="text-xs font-bold text-ink-soft uppercase tracking-wider">{group.name}</Text>
            </View>
            <Text className="text-xs text-ink-faint mt-0.5">Muro de consecuencias del ciclo actual</Text>
          </View>

          <ScrollView className="flex-1 px-6 py-4" showsVerticalScrollIndicator={false}>
            {error ? (
              <View className="mb-4 rounded-3xl bg-coral/10 border border-coral/20 px-4 py-3">
                <Text className="text-sm font-semibold text-coral-deep">{error}</Text>
              </View>
            ) : null}

            {/* Ruletas abiertas: cualquiera del squad puede entrar a sugerir */}
            {openEntries.length > 0 && (
              <View className="mb-6">
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">RULETAS ABIERTAS</Text>
                <View className="gap-2.5">
                  {openEntries.map((e) => (
                    <View key={e.id} className="rounded-3xl bg-paper border border-hairline border-l-4 border-l-terracotta p-4 flex-row items-center gap-3 shadow-sm">
                      <View className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(entryDebtorName(e))}`}>
                        <Text className="text-paper text-sm font-bold">{getInitials(entryDebtorName(e))}</Text>
                      </View>
                      <View className="flex-1 min-w-0">
                        <Text className="font-semibold text-sm text-ink truncate">{entryDebtorName(e)}</Text>
                        <Text className="text-xs text-ink-soft mt-0.5">{entryCountdown(e)}</Text>
                      </View>
                      <TouchableOpacity
                        activeOpacity={0.8}
                        onPress={() => handleEnterEntry(e)}
                        className="rounded-full bg-terracotta px-4 py-2"
                      >
                        <Text className="text-paper text-xs font-bold">Entrar →</Text>
                      </TouchableOpacity>
                    </View>
                  ))}
                </View>
              </View>
            )}

            {/* En el Bote Section */}
            <View className="mb-6">
              <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">EN EL BOTE ESTA SEMANA</Text>

              {eligibleWithoutEntry.length === 0 ? (
                <View className="rounded-3xl border border-sage/30 bg-sage-soft/30 p-6 items-center justify-center shadow-sm">
                  {openEntries.length > 0 ? (
                    <>
                      <Text className="text-4xl mb-3">🎡</Text>
                      <Text className="text-sm font-bold text-sage-deep">Todos los del bote ya tienen ruleta</Text>
                      <Text className="text-xs text-ink-soft mt-1">Entra arriba a proponer su penitencia.</Text>
                    </>
                  ) : (
                    <>
                      <Text className="text-4xl mb-3">🎉</Text>
                      <Text className="text-sm font-bold text-sage-deep">Squad limpio esta semana</Text>
                      <Text className="text-xs text-ink-soft mt-1">Nadie falló ningún hábito.</Text>
                    </>
                  )}
                </View>
              ) : (
                <View className="gap-2.5">
                  {eligibleWithoutEntry.map((m) => (
                    <View key={m.user_id} className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3 shadow-sm">
                      <View className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(m.display_name)}`}>
                        <Text className="text-paper text-sm font-bold">{getInitials(m.display_name)}</Text>
                      </View>
                      <View className="flex-1 min-w-0">
                        <Text className="font-semibold text-sm text-ink truncate">{m.display_name}</Text>
                        <Text className="text-xs text-ink-soft mt-0.5">Falló hábitos esta semana</Text>
                      </View>
                      <TouchableOpacity
                        activeOpacity={0.8}
                        onPress={() => handleOpenRoulette(m)}
                        className="rounded-full bg-terracotta px-4 py-2"
                      >
                        <Text className="text-paper text-xs font-bold">Abrir →</Text>
                      </TouchableOpacity>
                    </View>
                  ))}
                </View>
              )}
            </View>

            {/* Deudas Activas Section */}
            <View className="mb-10">
              <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">DEUDAS ACTIVAS EN EL SQUAD</Text>
              
              {debts.length === 0 ? (
                <View className="rounded-3xl bg-surface border border-hairline py-8 px-4 items-center justify-center">
                  <Text className="text-sm text-ink-soft text-center">Sin deudas activas en el grupo.</Text>
                </View>
              ) : (
                <View className="gap-2.5">
                  {debts.map((debt) => (
                    <View
                      key={debt.id}
                      className={`rounded-3xl bg-paper border border-hairline p-4 border-l-4 shadow-sm ${debt.scope === 'collective' ? 'border-l-coral' : 'border-l-terracotta'}`}
                    >
                      <View className="flex-row items-center justify-between mb-2">
                        <View className="flex-row items-center gap-2 flex-wrap">
                          <View className={`w-6 h-6 rounded-full items-center justify-center ${getAvatarBg(getMemberName(debt.debtor_id))}`}>
                            <Text className="text-paper text-[8px] font-bold">{getInitials(getMemberName(debt.debtor_id))}</Text>
                          </View>
                          <Text className="text-xs font-bold text-ink">{getMemberName(debt.debtor_id)}</Text>
                          <View className={`rounded-full px-2 py-0.5 ${debt.scope === 'collective' ? 'bg-coral/10' : 'bg-terracotta/10'}`}>
                            <Text className={`text-[9px] font-bold ${debt.scope === 'collective' ? 'text-coral-deep' : 'text-terracotta'}`}>
                              {debt.scope === 'collective' ? 'colectiva' : 'personal'}
                            </Text>
                          </View>
                        </View>
                        <Text className="text-[9px] text-ink-faint">exp. {getShortDate(debt.expires_at)}</Text>
                      </View>
                      <Text className="text-sm font-semibold text-ink pl-1">
                        {debt.punishment_emoji ?? ''} {debt.punishment_text}
                      </Text>
                      {debt.debtor_id === user?.id && (
                        <TouchableOpacity
                          activeOpacity={0.8}
                          disabled={completing === debt.id}
                          onPress={() => handleCompleteDebt(debt)}
                          className={`mt-3 rounded-full py-2 items-center ${
                            confirmComplete === debt.id
                              ? 'bg-sage-deep'
                              : 'border border-sage-deep bg-paper'
                          }`}
                        >
                          <Text
                            className={`text-xs font-bold ${
                              confirmComplete === debt.id ? 'text-paper' : 'text-sage-deep'
                            }`}
                          >
                            {completing === debt.id
                              ? 'Guardando…'
                              : confirmComplete === debt.id
                                ? '¿Seguro? El squad lo verá'
                                : '✓ Ya cumplí mi penitencia'}
                          </Text>
                        </TouchableOpacity>
                      )}
                    </View>
                  ))}
                </View>
              )}
            </View>

            {/* Historial de deudas */}
            <View className="mb-10">
              <TouchableOpacity activeOpacity={0.7} onPress={toggleHistory} className="flex-row items-center gap-2 mb-3 px-1">
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">HISTORIAL</Text>
                <Text className="text-[10px] text-ink-faint">{showHistory ? '▲' : '▼'}</Text>
              </TouchableOpacity>

              {showHistory &&
                (resolvedDebts.length === 0 ? (
                  <View className="rounded-3xl bg-surface border border-hairline py-6 px-4 items-center justify-center">
                    <Text className="text-xs text-ink-soft text-center">
                      Sin deudas pasadas. El historial del squad aparecerá aquí.
                    </Text>
                  </View>
                ) : (
                  <View className="gap-2.5">
                    {resolvedDebts.map((debt) => {
                      const badge = DEBT_STATUS_BADGE[debt.status ?? ''] ?? {
                        label: (debt.status ?? '').toUpperCase(),
                        bg: 'bg-cream-2',
                        text: 'text-ink-faint',
                      };
                      return (
                        <View key={debt.id} className="rounded-3xl bg-surface border border-hairline p-4">
                          <View className="flex-row items-center justify-between mb-2">
                            <View className="flex-row items-center gap-2">
                              <View className={`w-6 h-6 rounded-full items-center justify-center opacity-70 ${getAvatarBg(getMemberName(debt.debtor_id))}`}>
                                <Text className="text-paper text-[8px] font-bold">{getInitials(getMemberName(debt.debtor_id))}</Text>
                              </View>
                              <Text className="text-xs font-bold text-ink">{getMemberName(debt.debtor_id)}</Text>
                            </View>
                            <View className={`rounded-full px-2.5 py-0.5 ${badge.bg}`}>
                              <Text className={`text-[9px] font-bold ${badge.text}`}>{badge.label}</Text>
                            </View>
                          </View>
                          <Text className="text-sm text-ink-soft pl-1">
                            {debt.punishment_emoji ?? ''} {debt.punishment_text}
                          </Text>
                        </View>
                      );
                    })}
                  </View>
                ))}
            </View>
          </ScrollView>
        </>
      )}

      {/* ═══════════════════ ENTRY VIEW (ROULETTE WHEEL) ══════════════════════ */}
      {view === 'entry' && activeEntry && (
        <View className="flex-1">
          {/* Header */}
          <View className="px-6 py-4 border-b border-hairline/30 bg-cream flex-row items-center gap-3">
            <TouchableOpacity
              activeOpacity={0.8}
              onPress={handleBack}
              className="w-9 h-9 rounded-full bg-surface border border-hairline items-center justify-center"
            >
              <Text className="font-bold text-ink">←</Text>
            </TouchableOpacity>
            <View>
              <Text className="font-serif text-lg font-semibold text-ink">Ruleta de {debtorName}</Text>
              <Text className="text-xs text-ink-soft">Consecuencias y penitencias</Text>
            </View>
          </View>

          <ScrollView className="flex-1 px-6 py-4" showsVerticalScrollIndicator={false}>
            {error ? (
              <View className="mb-4 rounded-3xl bg-coral/10 border border-coral/20 px-4 py-3">
                <Text className="text-sm font-semibold text-coral-deep">{error}</Text>
              </View>
            ) : null}

            {/* SPIN RESULT SPLASH CARD */}
            {spinResult ? (
              <View className="rounded-3xl bg-paper border border-hairline p-6 mb-5 items-center shadow-md">
                <Text className="text-[10px] font-bold text-terracotta tracking-wider uppercase mb-3">
                  {spinResult.scope === 'collective' ? 'DEUDA COLECTIVA' : 'PENITENCIA ASIGNADA'}
                </Text>
                
                <View className={`w-16 h-16 rounded-full items-center justify-center mb-3 ${getAvatarBg(getMemberName(spinResult.debtor_id))}`}>
                  <Text className="text-paper text-xl font-bold">{getInitials(getMemberName(spinResult.debtor_id))}</Text>
                </View>

                <Text className="font-serif text-2xl font-semibold text-ink mb-4">{getMemberName(spinResult.debtor_id)}</Text>
                
                <View className="w-full rounded-2xl bg-terracotta/5 border border-terracotta/20 p-4 mb-4 items-center">
                  <Text className="text-[9px] font-bold text-terracotta tracking-wider uppercase mb-1.5">PENITENCIA</Text>
                  <Text className="font-semibold text-sm text-ink text-center">
                    {spinResult.punishment_emoji ?? ''} {spinResult.punishment_text}
                  </Text>
                </View>

                {spinResult.scope === 'collective' && (
                  <Text className="text-[11px] text-coral-deep font-semibold text-center mb-3 px-4">
                    Nadie propuso penitencia a tiempo — el squad completo paga.
                  </Text>
                )}

                <Text className="text-[10px] text-ink-faint">
                  Expira el {getShortDate(spinResult.expires_at)}
                </Text>

                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={handleBack}
                  className="w-full rounded-full bg-sage-deep py-3.5 items-center mt-6"
                >
                  <Text className="text-paper font-bold text-sm">Volver a la ruleta</Text>
                </TouchableOpacity>
              </View>
            ) : (
              /* PRE-SPIN ROULETTE STAGE */
              <View className="items-center">
                {/* SVG Pointer */}
                <View 
                  className="w-0 h-0 border-l-[9px] border-r-[9px] border-b-[18px] border-l-transparent border-r-transparent border-b-ink z-10"
                  style={{ marginBottom: -2 }}
                />

                {/* Rotating Wheel Container */}
                <Animated.View
                  style={{
                    transform: [{ rotate: rotateInterpolate }],
                  }}
                >
                  <Svg width={220} height={220} viewBox="0 0 200 200">
                    <G>
                      {wheelSegments.map((seg, i) => (
                        <Path key={i} d={seg.path} fill={seg.color} />
                      ))}
                    </G>
                    {/* Outer border & inner circle */}
                    <Circle cx="100" cy="100" r="28" fill="white" opacity="0.9" />
                    <Circle cx="100" cy="100" r="10" fill="#2A251F" />
                  </Svg>
                </Animated.View>

                {/* Deadline indicator */}
                {!deadlinePassed ? (
                  <View className="mt-4 rounded-full bg-amber-soft px-3 py-1 flex-row items-center gap-1.5">
                    <Text className="text-xs font-semibold text-amber-deep">Gira en: {deadlineLabel}</Text>
                  </View>
                ) : (
                  !activeEntry.spun_at && (
                    <View className="mt-4 rounded-full bg-coral-soft px-3 py-1">
                      <Text className="text-xs font-semibold text-coral-deep">¡La ruleta está lista!</Text>
                    </View>
                  )
                )}

                {/* Suggestions List */}
                <View className="w-full mt-6 mb-6">
                  <View className="flex-row justify-between items-center mb-3">
                    <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase">PENITENCIAS EN JUEGO</Text>
                    <Text className="text-xs font-bold text-ink-soft">{suggestions.length}</Text>
                  </View>

                  {suggestions.length === 0 ? (
                    <View className="rounded-3xl bg-surface border border-hairline py-6 px-4 items-center justify-center mb-4">
                      <Text className="text-xs text-ink-soft text-center leading-normal">
                        Nadie ha propuesto una penitencia aún.
                      </Text>
                    </View>
                  ) : (
                    <View className="flex-row flex-wrap gap-1.5 mb-4">
                      {suggestions.map((s, i) => (
                        <View
                          key={s.id}
                          className="rounded-full px-3 py-1.5"
                          style={{ backgroundColor: WHEEL_COLORS[i % WHEEL_COLORS.length] }}
                        >
                          <Text className="text-xs font-semibold text-paper leading-none">
                            {s.emoji ? s.emoji + ' ' : ''}{s.text}
                          </Text>
                        </View>
                      ))}
                    </View>
                  )}

                  {/* Suggestion Form */}
                  {canSuggest && (
                    !showForm ? (
                      <TouchableOpacity
                        activeOpacity={0.8}
                        onPress={() => setShowForm(true)}
                        className="w-full rounded-full border border-sage-deep bg-paper py-3 items-center"
                      >
                        <Text className="text-sage-deep font-bold text-xs">+ Proponer penitencia</Text>
                      </TouchableOpacity>
                    ) : (
                      <View className="rounded-3xl bg-surface border border-hairline p-4 mb-4">
                        <Text className="text-[9px] font-bold text-ink-soft tracking-wider uppercase mb-3">TU PROPUESTA</Text>
                        
                        <View className="flex-row gap-2 mb-3">
                          <TextInput
                            placeholder="😈"
                            placeholderTextColor="#A89C89"
                            maxLength={2}
                            value={sugEmoji}
                            onChangeText={setSugEmoji}
                            className="w-14 rounded-xl border border-hairline bg-paper px-3 py-2.5 text-center text-lg focus:border-sage-deep"
                          />
                          <TextInput
                            placeholder="Ej: 30 sentadillas en público"
                            placeholderTextColor="#A89C89"
                            value={sugText}
                            onChangeText={setSugText}
                            className="flex-1 rounded-xl border border-hairline bg-paper px-3 py-2.5 text-sm focus:border-sage-deep text-ink"
                          />
                        </View>

                        <View className="flex-row gap-2">
                          <TouchableOpacity
                            activeOpacity={0.8}
                            onPress={handleSuggest}
                            className="flex-1 rounded-full bg-sage-deep py-2.5 items-center"
                          >
                            <Text className="text-paper font-bold text-xs">Enviar</Text>
                          </TouchableOpacity>
                          
                          <TouchableOpacity
                            activeOpacity={0.8}
                            onPress={() => setShowForm(false)}
                            className="rounded-full border border-hairline bg-paper px-4 py-2.5 items-center"
                          >
                            <Text className="text-ink-soft font-bold text-xs">Cancelar</Text>
                          </TouchableOpacity>
                        </View>
                      </View>
                    )
                  )}

                  {(!deadlinePassed && isDebtor) && (
                    <View className="rounded-3xl bg-surface border border-hairline py-3 px-4 items-center justify-center">
                      <Text className="text-xs text-ink-soft text-center">
                        Tu squad escribe tus penitencias… tú solo giras. 😈
                      </Text>
                    </View>
                  )}

                  {(!deadlinePassed && hasSuggested) && (
                    <View className="rounded-full bg-sage-soft py-3 items-center justify-center">
                      <Text className="text-sage-deep font-semibold text-xs">✓ Ya propusiste tu penitencia</Text>
                    </View>
                  )}
                </View>

                {/* SPIN BUTTON */}
                {canSpin ? (
                  <TouchableOpacity
                    disabled={spinning}
                    activeOpacity={0.8}
                    onPress={handleSpin}
                    className={`w-full rounded-full bg-terracotta py-4 items-center justify-center mb-10 shadow-sm ${spinning ? 'opacity-60' : ''}`}
                  >
                    {spinning ? (
                      <View className="flex-row items-center gap-2">
                        <ActivityIndicator size="small" color="#FFFFFF" />
                        <Text className="text-paper font-bold text-sm">Girando…</Text>
                      </View>
                    ) : (
                      <Text className="text-paper font-bold text-sm">
                        {isDebtor ? '⊕ Girar la ruleta' : `⊕ Girar por ${debtorName}`}
                      </Text>
                    )}
                  </TouchableOpacity>
                ) : (
                  (deadlinePassed && !isDebtor && !activeEntry.spun_at) && (
                    <View className="w-full rounded-3xl bg-amber-soft/40 border border-amber/30 p-4 items-center mb-10">
                      <Text className="text-sm font-semibold text-amber-deep">Esperando que {debtorName} gire</Text>
                      <Text className="text-[11px] text-ink-soft mt-1 text-center">
                        Si no gira en 24h, cualquiera del squad podrá girar por él.
                      </Text>
                    </View>
                  )
                )}

                {activeEntry.spun_at && !spinResult && (
                  <View className="w-full rounded-3xl bg-surface border border-hairline p-4 items-center mb-10">
                    <Text className="text-sm text-ink-soft">Esta ruleta ya fue girada.</Text>
                  </View>
                )}
              </View>
            )}
          </ScrollView>
        </View>
      )}
    </SafeAreaView>
  );
}
