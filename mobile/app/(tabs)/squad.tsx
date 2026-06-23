import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Share,
  Clipboard,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp } from '../../src/contexts/AppContext';

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

export default function SquadScreen() {
  const router = useRouter();
  const { signOut, user } = useAuth();
  const { group, members, leaveGroup } = useApp();
  
  const [copied, setCopied] = useState(false);
  const [confirmLeave, setConfirmLeave] = useState(false);
  const [leaving, setLeaving] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);

  const displayUserName = user?.user_metadata?.display_name || user?.email || 'Tú';

  async function handleLogout() {
    setLoggingOut(true);
    try {
      await signOut();
      router.replace('/(auth)/login');
    } catch (err) {
      console.error(err);
    } finally {
      setLoggingOut(false);
    }
  }

  async function handleLeaveGroup() {
    setLeaving(true);
    try {
      await leaveGroup();
      router.replace('/onboarding');
    } catch (err) {
      console.error(err);
      setConfirmLeave(false);
    } finally {
      setLeaving(false);
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
      await Share.share({
        message: text,
        title: 'Únete a Dydi',
      });
    } catch (e) {
      console.error(e);
    }
  }

  return (
    <SafeAreaView className="flex-1 bg-cream" edges={['top']}>
      {/* Header */}
      <View className="px-6 py-4 border-b border-hairline/30 bg-cream">
        <Text className="font-serif text-2xl font-semibold text-ink">Mi Squad</Text>
        <Text className="text-xs text-ink-soft mt-0.5">Administración y miembros del grupo</Text>
      </View>

      <ScrollView className="flex-1 px-6 py-4" showsVerticalScrollIndicator={false}>
        {group ? (
          <>
            {/* Squad Info Card */}
            <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm">
              <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-1">MI GRUPO</Text>
              <Text className="font-serif text-2xl font-semibold text-ink mb-4">{group.name}</Text>
              
              <View className="rounded-2xl bg-cream-2 border border-hairline/60 p-4 mb-4">
                <Text className="text-[9px] font-bold text-ink-soft tracking-wider uppercase mb-1">CÓDIGO DE INVITACIÓN</Text>
                <Text className="font-mono text-xs text-ink break-all select-all leading-normal">
                  {group.id}:{group.invite_code}
                </Text>
              </View>

              <View className="flex-row gap-2">
                <TouchableOpacity
                  activeOpacity={0.8}
                  onPress={copyInviteCode}
                  className={`flex-1 rounded-full border border-hairline py-3 items-center ${copied ? 'bg-sage-soft/30 border-sage/40' : 'bg-surface'}`}
                >
                  <Text className={`font-bold text-xs ${copied ? 'text-sage-deep' : 'text-ink-soft'}`}>
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

            {/* Members Section */}
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">MIEMBROS DEL SQUAD ({members.length + 1})</Text>
            
            <View className="gap-2 mb-6">
              {/* Me */}
              <View className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3">
                <View className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(displayUserName)}`}>
                  <Text className="text-paper text-sm font-bold">{getInitials(displayUserName)}</Text>
                </View>
                <View className="flex-1 min-w-0">
                  <Text className="font-semibold text-sm text-ink truncate">{displayUserName} (Tú)</Text>
                  <Text className="text-xs text-ink-soft truncate mt-0.5">{user?.email}</Text>
                </View>
              </View>

              {/* Others */}
              {members.map((member) => (
                <View key={member.user_id} className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3">
                  <View className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(member.display_name)}`}>
                    <Text className="text-paper text-sm font-bold">{getInitials(member.display_name)}</Text>
                  </View>
                  <View className="flex-1 min-w-0">
                    <Text className="font-semibold text-sm text-ink truncate">{member.display_name}</Text>
                  </View>
                </View>
              ))}
            </View>

            {/* Leave Group Section */}
            {!confirmLeave ? (
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => setConfirmLeave(true)}
                className="w-full rounded-full border border-hairline bg-paper py-3.5 items-center mb-4 shadow-sm"
              >
                <Text className="text-coral-deep font-bold text-sm">Salir del grupo</Text>
              </TouchableOpacity>
            ) : (
              <View className="rounded-3xl border border-coral/40 bg-coral-soft/10 p-5 mb-5">
                <Text className="text-sm font-bold text-ink mb-1">
                  ¿Seguro que quieres salir de <Text className="text-coral-deep">{group.name}</Text>?
                </Text>
                <Text className="text-xs text-ink-soft mb-4">Perderás tus hábitos y rachas en este grupo.</Text>
                <View className="flex-row gap-2">
                  <TouchableOpacity
                    disabled={leaving}
                    activeOpacity={0.8}
                    onPress={handleLeaveGroup}
                    className="flex-1 rounded-full bg-coral-deep py-2.5 items-center"
                  >
                    {leaving ? (
                      <ActivityIndicator size="small" color="#FFFFFF" />
                    ) : (
                      <Text className="text-paper font-bold text-xs">Sí, salir</Text>
                    )}
                  </TouchableOpacity>

                  <TouchableOpacity
                    activeOpacity={0.8}
                    onPress={() => setConfirmLeave(false)}
                    className="flex-1 rounded-full border border-hairline bg-paper py-2.5 items-center"
                  >
                    <Text className="text-ink-soft font-bold text-xs">Cancelar</Text>
                  </TouchableOpacity>
                </View>
              </View>
            )}
          </>
        ) : (
          <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm items-center justify-center py-8">
            <Text className="text-sm text-ink-soft mb-4">No estás asociado a ningún grupo.</Text>
            <TouchableOpacity
              activeOpacity={0.8}
              onPress={() => router.replace('/onboarding')}
              className="rounded-full bg-sage-deep px-6 py-2.5"
            >
              <Text className="text-paper font-bold text-xs">Crear o Unirme</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* Logout Button */}
        <TouchableOpacity
          disabled={loggingOut}
          activeOpacity={0.8}
          onPress={handleLogout}
          className="w-full rounded-full border border-hairline bg-paper py-3.5 items-center mb-10 shadow-sm"
        >
          {loggingOut ? (
            <ActivityIndicator size="small" color="#6F6557" />
          ) : (
            <Text className="text-ink-soft font-bold text-sm">Cerrar sesión</Text>
          )}
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}
