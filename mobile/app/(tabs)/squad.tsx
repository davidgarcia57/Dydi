import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  TextInput,
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
import { api } from '../../lib/api';

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
  const { signOut, user, updateDisplayName } = useAuth();
  const { group, members, myGroups, loadMyGroups, switchGroup, leaveGroup, propose } = useApp();

  const [copied, setCopied] = useState(false);
  const [confirmLeave, setConfirmLeave] = useState(false);
  const [leaving, setLeaving] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);

  // Perfil
  const [nameInput, setNameInput] = useState('');
  const [savingName, setSavingName] = useState(false);
  const [profileMsg, setProfileMsg] = useState('');

  // Kick
  const [confirmKick, setConfirmKick] = useState<string | null>(null);
  const [kicking, setKicking] = useState<string | null>(null);
  const [kickMsg, setKickMsg] = useState('');

  // Switcher
  const [switching, setSwitching] = useState<string | null>(null);

  const displayUserName = user?.user_metadata?.display_name || user?.email || 'Tú';

  useEffect(() => {
    setNameInput(user?.user_metadata?.display_name || '');
  }, [user?.user_metadata?.display_name]);

  useEffect(() => {
    if (!myGroups.length) loadMyGroups();
  }, []);

  async function handleSaveName() {
    const name = nameInput.trim();
    if (!name || savingName || name === user?.user_metadata?.display_name) return;
    setSavingName(true);
    setProfileMsg('');
    try {
      await updateDisplayName(name);
      // Sincroniza public.users para que el resto del squad vea el nombre nuevo.
      await api('/api/users/sync', {
        method: 'POST',
        body: JSON.stringify({ display_name: name, avatar_url: null }),
      });
      setProfileMsg('✓ Nombre actualizado');
    } catch (e: any) {
      setProfileMsg(e?.error ?? e?.message ?? 'No se pudo actualizar');
    } finally {
      setSavingName(false);
      setTimeout(() => setProfileMsg(''), 2500);
    }
  }

  async function handleSwitchGroup(id: string) {
    if (switching || id === group?.id) return;
    setSwitching(id);
    try {
      await switchGroup(id);
    } catch (err) {
      console.error(err);
    } finally {
      setSwitching(null);
    }
  }

  async function handleKick(member: any) {
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
        {/* Perfil */}
        <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm">
          <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3">TU PERFIL</Text>
          <View className="flex-row items-center gap-3 mb-3">
            <View className={`w-12 h-12 rounded-full items-center justify-center ${getAvatarBg(displayUserName)}`}>
              <Text className="text-paper text-base font-bold">{getInitials(displayUserName)}</Text>
            </View>
            <View className="flex-1 min-w-0">
              <Text className="font-semibold text-sm text-ink truncate">{displayUserName}</Text>
              <Text className="text-xs text-ink-soft truncate mt-0.5">{user?.email}</Text>
            </View>
          </View>
          <View className="flex-row gap-2">
            <TextInput
              value={nameInput}
              onChangeText={setNameInput}
              placeholder="Tu nombre"
              placeholderTextColor="#A89C89"
              maxLength={40}
              className="flex-1 rounded-xl border border-hairline bg-cream-2 px-3 py-2.5 text-sm text-ink"
            />
            <TouchableOpacity
              activeOpacity={0.8}
              disabled={savingName || !nameInput.trim() || nameInput.trim() === (user?.user_metadata?.display_name || '')}
              onPress={handleSaveName}
              className={`rounded-full px-4 items-center justify-center ${
                savingName || !nameInput.trim() || nameInput.trim() === (user?.user_metadata?.display_name || '')
                  ? 'bg-cream-2'
                  : 'bg-sage-deep'
              }`}
            >
              {savingName ? (
                <ActivityIndicator size="small" color="#5C7650" />
              ) : (
                <Text
                  className={`text-xs font-bold ${
                    !nameInput.trim() || nameInput.trim() === (user?.user_metadata?.display_name || '')
                      ? 'text-ink-faint'
                      : 'text-paper'
                  }`}
                >
                  Guardar
                </Text>
              )}
            </TouchableOpacity>
          </View>
          {profileMsg ? (
            <Text className="text-xs font-semibold text-sage-deep mt-2">{profileMsg}</Text>
          ) : null}
        </View>

        {/* Mis grupos (switcher) */}
        {myGroups.length > 1 && (
          <View className="rounded-3xl bg-paper border border-hairline p-5 mb-5 shadow-sm">
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3">MIS GRUPOS</Text>
            <View className="gap-2">
              {myGroups.map((g) => (
                <TouchableOpacity
                  key={g.id}
                  activeOpacity={0.8}
                  onPress={() => handleSwitchGroup(g.id)}
                  className={`rounded-2xl border px-4 py-3 flex-row items-center justify-between ${
                    g.id === group?.id ? 'border-sage-deep bg-sage-soft/30' : 'border-hairline bg-cream-2'
                  }`}
                >
                  <Text
                    className={`text-sm font-semibold ${g.id === group?.id ? 'text-sage-deep' : 'text-ink'}`}
                  >
                    {g.name}
                  </Text>
                  {switching === g.id ? (
                    <ActivityIndicator size="small" color="#5C7650" />
                  ) : g.id === group?.id ? (
                    <Text className="text-xs font-bold text-sage-deep">Activo</Text>
                  ) : null}
                </TouchableOpacity>
              ))}
            </View>
          </View>
        )}

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
            <Text className="text-[10px] font-bold text-ink-soft tracking-wider uppercase mb-3 px-1">
              MIEMBROS DEL SQUAD ({members.filter((m) => m.user_id !== user?.id).length + 1})
            </Text>
            
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
              {members
                .filter((m) => m.user_id !== user?.id)
                .map((member) => (
                  <View key={member.user_id} className="rounded-3xl bg-paper border border-hairline p-4 flex-row items-center gap-3">
                    <View className={`w-10 h-10 rounded-full items-center justify-center ${getAvatarBg(member.display_name)}`}>
                      <Text className="text-paper text-sm font-bold">{getInitials(member.display_name)}</Text>
                    </View>
                    <View className="flex-1 min-w-0">
                      <Text className="font-semibold text-sm text-ink truncate">{member.display_name}</Text>
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
