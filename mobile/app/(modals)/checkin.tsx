import React, { useEffect, useState, useMemo } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  ActivityIndicator,
  TextInput,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp } from '../../src/contexts/AppContext';
import WaterBottle from '../../src/components/ui/WaterBottle';
import HabitIcon from '../../src/components/ui/HabitIcon';

export default function CheckinModal() {
  const router = useRouter();
  const { user } = useAuth();
  const { group, todayCheckins, streaks, checkin, loadToday, loadStreaks } = useApp();

  // 'loading' | 'error' | 'no-habit' | 'select' | 'confirm' | 'success' | 'done'
  const [step, setStep] = useState<'loading' | 'error' | 'no-habit' | 'select' | 'confirm' | 'success' | 'done'>('loading');
  const [selectedHabit, setSelectedHabit] = useState<any>(null);
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [errMsg, setErrMsg] = useState('');
  const [prevStreak, setPrevStreak] = useState(0);
  const [newStreak, setNewStreak] = useState(0);
  const [showPlus, setShowPlus] = useState(false);
  const [currentTime, setCurrentTime] = useState(formatTime());

  function formatTime() {
    const d = new Date();
    return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  }

  // Update clock
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(formatTime());
    }, 10000);
    return () => clearInterval(timer);
  }, []);

  const myHabits = useMemo(() => {
    return todayCheckins.filter((c) => c.user_id === user?.id);
  }, [todayCheckins, user?.id]);

  const myPending = useMemo(() => {
    return myHabits.filter((c) => c.status === 'pending');
  }, [myHabits]);

  function isWaterHabit(h: any) {
    return h?.icon_key === 'water' || /agua|water/i.test(h?.habit_name || '');
  }

  async function loadData() {
    if (!group?.id || !user?.id) {
      setStep('error');
      return;
    }
    setStep('loading');
    setErrMsg('');
    try {
      await loadToday(group.id);
      await loadStreaks(user.id);
      const currentStreak = streaks[user.id] ?? 0;
      setPrevStreak(currentStreak);
    } catch (e: any) {
      setErrMsg(e?.message || 'No pudimos cargar tus hábitos.');
      setStep('error');
    }
  }

  // Initial loading
  useEffect(() => {
    loadData();
  }, []);

  // Solve the starting screen step when checkins load
  useEffect(() => {
    if (step === 'loading' && todayCheckins.length > 0) {
      resolveStep();
    }
  }, [todayCheckins, step]);

  function resolveStep() {
    if (myHabits.length === 0) {
      setStep('no-habit');
    } else if (myPending.length === 0) {
      setStep('done');
    } else if (myPending.length === 1) {
      setSelectedHabit(myPending[0]);
      setStep('confirm');
    } else {
      setStep('select');
    }
  }

  function pickHabit(habit: any) {
    setSelectedHabit(habit);
    setStep('confirm');
  }

  async function handleSubmit() {
    if (!selectedHabit || !group?.id || !user?.id || submitting) return;
    setSubmitting(true);
    setErrMsg('');
    try {
      await checkin(group.id, selectedHabit.habit_id, note.trim());
      await loadStreaks(user.id);
      
      // Compute new streak
      const updatedStreaks = useApp().streaks; // Get freshest values
      const currentStreak = updatedStreaks[user.id] ?? (prevStreak + 1);
      setNewStreak(currentStreak);

      setStep('success');
      setTimeout(() => {
        setShowPlus(true);
      }, 600);
    } catch (e: any) {
      setErrMsg(e?.message || 'Algo salió mal, intenta de nuevo.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <SafeAreaView className="flex-1 bg-cream">
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        className="flex-1"
      >
        <ScrollView
          contentContainerStyle={{ flexGrow: 1 }}
          className="flex-grow-1"
          keyboardShouldPersistTaps="handled"
        >
          {/* ── LOADING ─────────────────────────────────────────────────────── */}
          {step === 'loading' && (
            <View className="flex-1 items-center justify-center">
              <ActivityIndicator size="large" color="#7CA39D" />
            </View>
          )}

          {/* ── ERROR ──────────────────────────────────────────────────────── */}
          {step === 'error' && (
            <View className="flex-1 items-center justify-center px-8 text-center">
              <Text className="text-[10px] font-bold text-coral-deep tracking-wider uppercase mb-2">ALGO FALLÓ</Text>
              <Text className="font-serif text-2xl font-semibold text-ink text-center mb-2">
                No pudimos cargar tus hábitos
              </Text>
              <Text className="text-sm text-ink-soft text-center mb-8">{errMsg}</Text>
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={loadData}
                className="rounded-full bg-sage-deep px-8 py-3.5"
              >
                <Text className="text-paper font-bold text-sm">Reintentar</Text>
              </TouchableOpacity>
            </View>
          )}

          {/* ── NO HABIT ────────────────────────────────────────────────────── */}
          {step === 'no-habit' && (
            <View className="flex-1 items-center justify-center px-8 text-center">
              <View className="w-16 h-16 rounded-full bg-amber-soft flex items-center justify-center mb-6">
                <Text className="text-2xl text-amber-deep">⚠</Text>
              </View>
              <Text className="text-[10px] font-bold text-amber-deep tracking-wider uppercase mb-2">SIN HÁBITO</Text>
              <Text className="font-serif text-2xl font-semibold text-ink text-center mb-2">
                No tienes un hábito asignado
              </Text>
              <Text className="text-sm text-ink-soft text-center mb-8">
                Pídele al squad que proponga uno — se aprueba por votación.
              </Text>
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => router.back()}
                className="rounded-full bg-ink px-8 py-3.5"
              >
                <Text className="text-paper font-bold text-sm">Volver</Text>
              </TouchableOpacity>
            </View>
          )}

          {/* ── ALREADY DONE ────────────────────────────────────────────────── */}
          {step === 'done' && (
            <View className="flex-1 items-center justify-center px-8 text-center">
              {isWaterHabit(myHabits[0]) ? (
                <View className="mb-6">
                  <WaterBottle size={120} />
                </View>
              ) : (
                <View className="w-16 h-16 rounded-full bg-sage-soft flex items-center justify-center mb-6">
                  <Text className="text-2xl text-sage-deep">✓</Text>
                </View>
              )}
              <Text className="text-[10px] font-bold text-sage-deep tracking-wider uppercase mb-2">HOY YA CUMPLISTE</Text>
              <Text className="font-serif text-3xl font-semibold text-ink text-center mb-1">
                {myHabits[0]?.habit_name ?? 'Tu hábito'}
              </Text>
              <Text className="text-sm text-ink-soft text-center mb-8">
                Racha actual: <Text className="font-bold text-terracotta">{prevStreak} días</Text>
              </Text>
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => router.back()}
                className="rounded-full bg-ink px-8 py-3.5"
              >
                <Text className="text-paper font-bold text-sm">Volver al squad</Text>
              </TouchableOpacity>
            </View>
          )}

          {/* ── SELECT HABIT (MULTIPLE PENDING) ──────────────────────────────── */}
          {step === 'select' && (
            <View className="flex-1 px-6 pt-8">
              <TouchableOpacity
                onPress={() => router.back()}
                className="w-9 h-9 rounded-full bg-surface border border-hairline items-center justify-center mb-8"
              >
                <Text className="font-bold text-ink">×</Text>
              </TouchableOpacity>
              
              <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-2">ELIGE TU HÁBITO DE HOY</Text>
              <Text className="font-serif text-3xl font-semibold text-ink mb-6">¿Cuál cumpliste?</Text>

              <View className="gap-3">
                {myPending.map((h) => (
                  <TouchableOpacity
                    key={h.habit_id}
                    activeOpacity={0.9}
                    onPress={() => pickHabit(h)}
                    className="w-full rounded-3xl bg-paper border border-hairline p-5 flex-row items-center gap-4 shadow-sm"
                  >
                    <View
                      className="w-12 h-12 rounded-full items-center justify-center"
                      style={{ backgroundColor: h.color || '#A8C39A' }}
                    >
                      <HabitIcon iconKey={h.icon_key} size={26} color="#FFFFFF" />
                    </View>
                    <View className="flex-1 min-w-0">
                      <Text className="font-semibold text-base text-ink truncate">{h.habit_name}</Text>
                      {h.scheduled_time && (
                        <Text className="text-xs text-ink-soft mt-0.5">{h.scheduled_time}</Text>
                      )}
                    </View>
                    <Text className="text-xl text-ink-faint">›</Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>
          )}

          {/* ── CONFIRM CHECKIN ────────────────────────────────────────────── */}
          {step === 'confirm' && selectedHabit && (
            <View className="flex-1 px-6 pt-8 pb-10 justify-between">
              <View>
                {/* Top header row */}
                <View className="flex-row items-center justify-between mb-8">
                  <TouchableOpacity
                    onPress={() => myPending.length > 1 ? setStep('select') : router.back()}
                    className="w-9 h-9 rounded-full bg-surface border border-hairline items-center justify-center"
                  >
                    <Text className="font-bold text-ink">{myPending.length > 1 ? '←' : '×'}</Text>
                  </TouchableOpacity>
                  
                  {/* Streak chip */}
                  <View className="flex-row items-center gap-1 rounded-full bg-amber-soft px-3 py-1.5 border border-amber/10">
                    <Text className="text-xs font-bold text-amber-deep">★ {prevStreak} días de racha</Text>
                  </View>
                </View>

                {/* Habit title */}
                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-2">TU HÁBITO DE HOY</Text>
                <Text className="font-serif text-3xl font-semibold text-ink leading-tight mb-5">
                  {selectedHabit.habit_name}
                </Text>

                {/* Meta details */}
                <View className="flex-row items-center gap-2 flex-wrap mb-10">
                  <View className="flex-row items-center gap-1 rounded-full bg-surface border border-hairline px-3 py-1.5">
                    <Text className="text-xs font-semibold text-ink-soft">🕒 {currentTime}</Text>
                  </View>
                  {selectedHabit.scheduled_time ? (
                    <View className="rounded-full bg-amber-soft px-3 py-1.5">
                      <Text className="text-xs font-semibold text-amber-deep">Meta: {selectedHabit.scheduled_time}</Text>
                    </View>
                  ) : (
                    <View className="rounded-full bg-sage-soft px-3 py-1.5">
                      <Text className="text-xs font-semibold text-sage-deep">Hoy hasta medianoche</Text>
                    </View>
                  )}
                </View>

                {/* Big Illustration */}
                <View className="items-center justify-center py-6 mb-4">
                  {isWaterHabit(selectedHabit) ? (
                    <WaterBottle size={140} />
                  ) : (
                    <View
                      className="w-24 h-24 rounded-full items-center justify-center shadow-sm"
                      style={{ backgroundColor: selectedHabit.color || '#A8C39A' }}
                    >
                      <HabitIcon iconKey={selectedHabit.icon_key} size={48} color="#FFFFFF" />
                    </View>
                  )}
                </View>

                {/* Note Area */}
                <View className="w-full mt-4">
                  <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-2">NOTA OPCIONAL</Text>
                  <TextInput
                    placeholder="¿Algo que quieras contarle al squad?"
                    placeholderTextColor="#A89C89"
                    multiline
                    numberOfLines={2}
                    value={note}
                    onChangeText={setNote}
                    className="w-full rounded-[14px] border border-hairline bg-surface px-4 py-3 text-sm text-ink font-sans"
                    style={{ height: 60, textAlignVertical: 'top' }}
                  />
                </View>
                
                {errMsg ? (
                  <Text className="text-sm text-coral-deep font-semibold mt-4 text-center">{errMsg}</Text>
                ) : null}
              </View>

              {/* Submit CTA button */}
              <View className="items-center mt-8">
                <TouchableOpacity
                  disabled={submitting}
                  activeOpacity={0.9}
                  onPress={handleSubmit}
                  className="w-24 h-24 rounded-full bg-sage-deep items-center justify-center shadow-lg active:scale-95 transition-all"
                >
                  {submitting ? (
                    <ActivityIndicator size="large" color="#FFFFFF" />
                  ) : (
                    <Text className="text-paper text-3xl font-bold">✓</Text>
                  )}
                </TouchableOpacity>
                <Text className="text-sm font-semibold text-ink-soft mt-3 mb-1">
                  {submitting ? 'Registrando…' : 'Toca para registrar'}
                </Text>
                <Text className="text-[10px] text-ink-faint">El squad verá tu check-in al instante</Text>
              </View>
            </View>
          )}

          {/* ── SUCCESS SCREEN ─────────────────────────────────────────────── */}
          {step === 'success' && selectedHabit && (
            <View className="flex-1 items-center justify-center px-8 text-center">
              {/* Checkring illustration */}
              {isWaterHabit(selectedHabit) ? (
                <View className="mb-8">
                  <WaterBottle size={130} />
                </View>
              ) : (
                <View className="w-28 h-28 rounded-full bg-sage-soft items-center justify-center mb-8 relative">
                  <View className="w-20 h-20 rounded-full bg-sage-deep items-center justify-center">
                    <Text className="text-paper text-4xl font-bold">✓</Text>
                  </View>
                </View>
              )}

              <Text className="text-[10px] font-bold text-sage-deep tracking-wider uppercase mb-2">¡LO LOGRASTE!</Text>
              <Text className="font-serif text-3xl font-semibold text-ink text-center mb-1">
                {selectedHabit.habit_name}
              </Text>
              <Text className="text-sm text-ink-soft text-center mb-8">Check-in registrado hoy</Text>

              {/* Streak card */}
              <View className="rounded-3xl bg-paper border border-hairline px-8 py-6 mb-10 w-full max-w-xs items-center relative shadow-sm">
                {showPlus && (
                  <View className="absolute -top-3 -right-2 rounded-full bg-terracotta px-3 py-1 shadow-sm">
                    <Text className="text-paper text-[10px] font-bold">+1 🔥</Text>
                  </View>
                )}

                <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-2">TU RACHA</Text>
                <Text className="font-serif text-5xl font-semibold text-terracotta leading-none mb-1">{newStreak}</Text>
                <Text className="text-xs text-ink-soft text-center mt-1">
                  {newStreak === 1 ? 'día — ¡arrancaste!' : 'días seguidos'}
                </Text>
              </View>

              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => router.back()}
                className="w-full max-w-xs rounded-full bg-ink py-4 items-center"
              >
                <Text className="text-paper font-bold text-sm">Ver el squad →</Text>
              </TouchableOpacity>
            </View>
          )}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
