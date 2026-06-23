import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Clipboard,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { useApp } from '../src/contexts/AppContext';
import BrandWordmark from '../src/components/ui/BrandWordmark';

export default function OnboardingScreen() {
  const router = useRouter();
  const { createGroup, joinGroup } = useApp();
  
  // 'home' | 'create' | 'join' | 'created'
  const [step, setStep] = useState<'home' | 'create' | 'join' | 'created'>('home');
  const [groupName, setGroupName] = useState('');
  const [joinCode, setJoinCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [errMsg, setErrMsg] = useState('');
  const [createdGroup, setCreatedGroup] = useState<any>(null);
  const [copied, setCopied] = useState(false);

  function goHome() {
    setStep('home');
    setErrMsg('');
    setGroupName('');
    setJoinCode('');
  }

  async function submitCreate() {
    if (!groupName.trim() || loading) return;
    setLoading(true);
    setErrMsg('');
    try {
      const g = await createGroup(groupName.trim());
      setCreatedGroup(g);
      setStep('created');
    } catch (e: any) {
      setErrMsg(e?.error ?? e?.message ?? 'No se pudo crear el grupo.');
    } finally {
      setLoading(false);
    }
  }

  async function submitJoin() {
    if (!joinCode.trim() || loading) return;
    setErrMsg('');

    // Expected format: "{groupID}:{inviteCode}"
    const parts = joinCode.trim().split(':');
    if (parts.length !== 2 || !parts[0] || !parts[1]) {
      setErrMsg('Formato inválido. Debe ser el código completo que te compartieron.');
      return;
    }

    setLoading(true);
    try {
      await joinGroup(parts[0], parts[1]);
      router.replace('/(tabs)');
    } catch (e: any) {
      setErrMsg(e?.error ?? e?.message ?? 'Código inválido o grupo no encontrado.');
    } finally {
      setLoading(false);
    }
  }

  function copyInviteCode() {
    if (!createdGroup) return;
    const code = `${createdGroup.id}:${createdGroup.invite_code}`;
    try {
      Clipboard.setString(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.warn('Could not copy to clipboard:', err);
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
          className="px-6 py-12"
          keyboardShouldPersistTaps="handled"
        >
          {/* Home step */}
          {step === 'home' && (
            <View className="flex-1 items-center justify-center">
              <BrandWordmark size="xl" />
              <Text className="font-serif text-2xl font-semibold text-ink mt-6 mb-2">
                Bienvenido
              </Text>
              <Text className="text-sm text-ink-soft text-center font-sans px-4 mb-10">
                Únete a tu squad o crea uno nuevo para empezar.
              </Text>

              <View className="w-full gap-3">
                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={() => setStep('create')}
                  className="w-full rounded-full bg-sage-deep py-4 items-center"
                >
                  <Text className="text-paper font-bold text-sm">Crear grupo →</Text>
                </TouchableOpacity>

                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={() => setStep('join')}
                  className="w-full rounded-full border border-hairline bg-paper py-4 items-center"
                >
                  <Text className="text-ink font-bold text-sm">Unirme con código</Text>
                </TouchableOpacity>
              </View>
            </View>
          )}

          {/* Create step */}
          {step === 'create' && (
            <View className="flex-1 justify-center">
              <TouchableOpacity onPress={goHome} className="mb-8">
                <Text className="text-sm text-ink-soft font-bold">← Volver</Text>
              </TouchableOpacity>

              <Text className="text-xs font-bold text-ink-soft tracking-wider uppercase mb-2">
                NUEVO GRUPO
              </Text>
              <Text className="font-serif text-3xl font-semibold text-ink leading-tight mb-8">
                ¿Cómo se llama tu squad?
              </Text>

              <TextInput
                maxLength={40}
                placeholder="Los Incumplidos, El Squad, …"
                placeholderTextColor="#A89C89"
                value={groupName}
                onChangeText={setGroupName}
                className="w-full rounded-[14px] border border-hairline bg-paper px-4 py-3.5 text-sm text-ink font-sans mb-6"
                autoFocus
              />

              {errMsg ? (
                <Text className="text-sm text-coral-deep font-medium mb-4">{errMsg}</Text>
              ) : null}

              <TouchableOpacity
                disabled={!groupName.trim() || loading}
                onPress={submitCreate}
                className={`w-full rounded-full bg-sage-deep py-4 items-center ${
                  !groupName.trim() || loading ? 'opacity-40' : ''
                }`}
              >
                {loading ? (
                  <View className="flex-row items-center gap-2">
                    <ActivityIndicator size="small" color="#FFFFFF" />
                    <Text className="text-paper font-bold text-sm">Creando…</Text>
                  </View>
                ) : (
                  <Text className="text-paper font-bold text-sm">Crear grupo →</Text>
                )}
              </TouchableOpacity>
            </View>
          )}

          {/* Created step */}
          {step === 'created' && (
            <View className="flex-1 items-center justify-center">
              <View className="w-16 h-16 rounded-full bg-sage/20 flex items-center justify-center mb-6">
                <Text className="text-2xl text-sage-deep">✓</Text>
              </View>

              <Text className="text-xs font-bold text-sage-deep tracking-wider uppercase mb-1">
                GRUPO CREADO
              </Text>
              <Text className="font-serif text-3xl font-semibold text-ink text-center mb-2">
                {createdGroup?.name}
              </Text>
              <Text className="text-sm text-ink-soft text-center px-4 mb-8">
                Comparte este código con tu squad para que se unan.
              </Text>

              {/* Invite code box */}
              <View className="w-full rounded-3xl bg-paper border border-hairline p-5 mb-3">
                <Text className="text-xs font-bold text-ink-soft tracking-wider uppercase mb-2">
                  CÓDIGO DE INVITACIÓN
                </Text>
                <Text className="font-mono text-xs text-ink break-all leading-relaxed select-all">
                  {createdGroup?.id}:{createdGroup?.invite_code}
                </Text>
              </View>

              <TouchableOpacity
                onPress={copyInviteCode}
                className={`w-full rounded-full border border-hairline py-3 items-center mb-6 ${
                  copied ? 'bg-sage-soft/30 border-sage/40' : 'bg-surface'
                }`}
              >
                <Text className={`font-semibold text-sm ${copied ? 'text-sage-deep' : 'text-ink-soft'}`}>
                  {copied ? '¡Copiado! ✓' : 'Copiar código'}
                </Text>
              </TouchableOpacity>

              <TouchableOpacity
                onPress={() => router.replace('/(tabs)')}
                className="w-full rounded-full bg-sage-deep py-4 items-center"
              >
                <Text className="text-paper font-bold text-sm">Ir a Hoy →</Text>
              </TouchableOpacity>
            </View>
          )}

          {/* Join step */}
          {step === 'join' && (
            <View className="flex-1 justify-center">
              <TouchableOpacity onPress={goHome} className="mb-8">
                <Text className="text-sm text-ink-soft font-bold">← Volver</Text>
              </TouchableOpacity>

              <Text className="text-xs font-bold text-ink-soft tracking-wider uppercase mb-2">
                UNIRSE A UN GRUPO
              </Text>
              <Text className="font-serif text-3xl font-semibold text-ink leading-tight mb-2">
                Pega el código que te compartieron
              </Text>
              <Text className="text-xs text-ink-soft mb-8">
                El código completo tiene el formato:{'\n'}
                <Text className="font-mono">id-del-grupo:código-acceso</Text>
              </Text>

              <TextInput
                multiline
                numberOfLines={3}
                placeholder="Pega aquí el código completo…"
                placeholderTextColor="#A89C89"
                value={joinCode}
                onChangeText={setJoinCode}
                className="w-full rounded-[14px] border border-hairline bg-paper px-4 py-3.5 text-sm text-ink font-mono mb-6"
                style={{ textAlignVertical: 'top', height: 80 }}
              />

              {errMsg ? (
                <Text className="text-sm text-coral-deep font-medium mb-4">{errMsg}</Text>
              ) : null}

              <TouchableOpacity
                disabled={!joinCode.trim() || loading}
                onPress={submitJoin}
                className={`w-full rounded-full bg-sage-deep py-4 items-center ${
                  !joinCode.trim() || loading ? 'opacity-40' : ''
                }`}
              >
                {loading ? (
                  <View className="flex-row items-center gap-2">
                    <ActivityIndicator size="small" color="#FFFFFF" />
                    <Text className="text-paper font-bold text-sm">Uniéndome…</Text>
                  </View>
                ) : (
                  <Text className="text-paper font-bold text-sm">Unirme al squad →</Text>
                )}
              </TouchableOpacity>
            </View>
          )}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
