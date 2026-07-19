// aqui comienza control de perfil (juanito)
import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, ScrollView, Alert, Share, ActivityIndicator } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter, Stack } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp } from '../../src/contexts/AppContext';
import { api } from '../../lib/api';
import { supabase } from '../../lib/supabase';

export default function ProfileModal() {
  const router = useRouter();
  const { user, signOut, updateDisplayName } = useAuth();
  const { group, leaveGroup, propose, resetGroup } = useApp();

  const [displayName, setDisplayName] = useState(user?.user_metadata?.display_name || user?.email?.split('@')[0] || 'Tú');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [feedback, setFeedback] = useState({ type: '', message: '' });

  const displayEmail = user?.email || '';

  async function handleSaveProfile() {
    if (!displayName.trim()) return;
    setLoading(true);
    setFeedback({ type: '', message: '' });
    try {
      await updateDisplayName(displayName.trim());
      await api('/api/users/sync', {
        method: 'POST',
        body: JSON.stringify({ display_name: displayName.trim(), avatar_url: null }),
      });
      setFeedback({ type: 'success', message: 'Nombre actualizado.' });
    } catch (err: any) {
      setFeedback({ type: 'error', message: err?.message || 'Error al actualizar perfil' });
    } finally {
      setLoading(false);
    }
  }

  async function handleChangePassword() {
    if (password.length < 6) {
      setFeedback({ type: 'error', message: 'Mínimo 6 caracteres.' });
      return;
    }
    if (password !== confirmPassword) {
      setFeedback({ type: 'error', message: 'Las contraseñas no coinciden.' });
      return;
    }
    setLoading(true);
    setFeedback({ type: '', message: '' });
    try {
      const { error } = await supabase.auth.updateUser({ password });
      if (error) throw error;
      setPassword('');
      setConfirmPassword('');
      setFeedback({ type: 'success', message: 'Contraseña actualizada.' });
    } catch (err: any) {
      setFeedback({ type: 'error', message: err?.message || 'Error al cambiar contraseña' });
    } finally {
      setLoading(false);
    }
  }

  async function handleDeleteAccount() {
    Alert.alert(
      "¿Borrar cuenta?",
      "Esta acción es irreversible. Se borrará todo tu progreso.",
      [
        { text: "Cancelar", style: "cancel" },
        { 
          text: "Sí, borrar", 
          style: "destructive",
          onPress: async () => {
            setLoading(true);
            try {
              await api('/api/users/me', { method: 'DELETE' });
              await signOut();
              router.replace('/(auth)/login');
            } catch (err: any) {
              setFeedback({ type: 'error', message: err?.message || 'Error al borrar cuenta' });
              setLoading(false);
            }
          }
        }
      ]
    );
  }

  async function handleLogout() {
    setLoading(true);
    try {
      await signOut();
      resetGroup();
      router.replace('/(auth)/login');
    } catch (err) {
      setLoading(false);
    }
  }

  async function shareInvite() {
    if (!group) return;
    const code = `${group.id}:${group.invite_code}`;
    const text = `¡Únete a mi squad "${group.name}" en Dydi!\nCódigo: ${code}`;
    try {
      await Share.share({ message: text, title: 'Únete a Dydi' });
    } catch (e) {
      console.error(e);
    }
  }

  async function handleLeaveGroup() {
    Alert.alert(
      "¿Salir del squad?",
      "Perderás tus hábitos y progreso en este grupo.",
      [
        { text: "Cancelar", style: "cancel" },
        { 
          text: "Sí, salir", 
          style: "destructive",
          onPress: async () => {
            setLoading(true);
            try {
              await leaveGroup();
              router.replace('/onboarding');
            } catch (err: any) {
              setFeedback({ type: 'error', message: err?.message || 'Error al salir' });
              setLoading(false);
            }
          }
        }
      ]
    );
  }

  async function proposeDissolve() {
    if (!group) return;
    Alert.alert(
      "¿Proponer disolver squad?",
      "El squad deberá votar por mayoría para que se elimine el grupo.",
      [
        { text: "Cancelar", style: "cancel" },
        { 
          text: "Sí, proponer", 
          style: "destructive",
          onPress: async () => {
            setLoading(true);
            try {
              await propose(group.id, 'delete_group');
              setFeedback({ type: 'success', message: 'Propuesta enviada a Votar.' });
            } catch (err: any) {
              setFeedback({ type: 'error', message: err?.message || 'Error al proponer' });
            } finally {
              setLoading(false);
            }
          }
        }
      ]
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-cream">
      <Stack.Screen options={{ headerShown: false }} />
      <View className="px-6 py-4 flex-row justify-between items-center border-b border-hairline/30">
        <Text className="font-serif text-2xl font-bold text-ink">Configuración</Text>
        <TouchableOpacity onPress={() => router.back()} className="p-2 -mr-2 bg-surface rounded-full">
          <Text className="text-lg text-ink font-bold leading-none">✕</Text>
        </TouchableOpacity>
      </View>

      <ScrollView className="flex-1 px-6" contentContainerStyle={{ paddingVertical: 20 }}>
        
        {feedback.message ? (
          <View className={`p-4 rounded-xl mb-4 border ${feedback.type === 'success' ? 'bg-sage-soft/30 border-sage/40' : 'bg-coral-soft/30 border-coral/40'}`}>
            <Text className={`text-sm font-semibold ${feedback.type === 'success' ? 'text-sage-deep' : 'text-coral-deep'}`}>
              {feedback.message}
            </Text>
          </View>
        ) : null}

        <View className="bg-paper p-5 rounded-3xl border border-hairline shadow-sm mb-5">
          <Text className="font-serif text-xl font-semibold text-ink mb-1">Perfil</Text>
          <Text className="text-xs text-ink-soft mb-4">Este nombre aparece en tu squad.</Text>
          
          <View className="mb-4">
            <Text className="text-xs font-bold text-ink mb-1.5">Nombre o apodo</Text>
            <TextInput
              className="bg-surface border border-hairline rounded-xl px-4 py-3 text-sm text-ink"
              value={displayName}
              onChangeText={setDisplayName}
              autoCapitalize="words"
            />
          </View>
          <TouchableOpacity 
            onPress={handleSaveProfile} 
            disabled={loading}
            className="bg-sage-deep py-3 rounded-full items-center"
          >
            <Text className="text-paper font-bold text-sm">Guardar perfil</Text>
          </TouchableOpacity>
        </View>

        <View className="bg-paper p-5 rounded-3xl border border-hairline shadow-sm mb-5">
          <Text className="font-serif text-xl font-semibold text-ink mb-1">Seguridad</Text>
          <Text className="text-xs text-ink-soft mb-4">{displayEmail}</Text>

          <View className="mb-3">
            <Text className="text-xs font-bold text-ink mb-1.5">Nueva contraseña</Text>
            <TextInput
              secureTextEntry
              className="bg-surface border border-hairline rounded-xl px-4 py-3 text-sm text-ink"
              value={password}
              onChangeText={setPassword}
            />
          </View>
          <View className="mb-4">
            <Text className="text-xs font-bold text-ink mb-1.5">Confirmar contraseña</Text>
            <TextInput
              secureTextEntry
              className="bg-surface border border-hairline rounded-xl px-4 py-3 text-sm text-ink"
              value={confirmPassword}
              onChangeText={setConfirmPassword}
            />
          </View>
          <TouchableOpacity 
            onPress={handleChangePassword} 
            disabled={loading}
            className="bg-ink py-3 rounded-full items-center"
          >
            <Text className="text-paper font-bold text-sm">Actualizar contraseña</Text>
          </TouchableOpacity>
        </View>

        {group && (
          <View className="bg-paper p-5 rounded-3xl border border-hairline shadow-sm mb-5">
            <Text className="font-serif text-xl font-semibold text-ink mb-1">Tu Squad</Text>
            <Text className="text-sm font-semibold text-ink mt-2 mb-1">{group.name}</Text>
            
            <TouchableOpacity onPress={shareInvite} className="bg-surface border border-hairline py-3 rounded-full items-center mt-3 mb-3">
              <Text className="text-ink font-bold text-sm">Compartir invitación</Text>
            </TouchableOpacity>

            <TouchableOpacity onPress={handleLeaveGroup} className="py-3 rounded-full items-center border border-coral/40 bg-coral/5 mb-3">
              <Text className="text-coral-deep font-bold text-sm">Salir del grupo</Text>
            </TouchableOpacity>

            <TouchableOpacity onPress={proposeDissolve} className="py-3 rounded-full items-center">
              <Text className="text-coral font-bold text-xs">Proponer disolver squad</Text>
            </TouchableOpacity>
          </View>
        )}

        <View className="bg-paper p-5 rounded-3xl border border-hairline shadow-sm mb-10">
          <Text className="font-serif text-xl font-semibold text-coral mb-4">Zona Delicada</Text>
          
          <TouchableOpacity 
            onPress={handleLogout} 
            disabled={loading}
            className="border border-hairline py-3 rounded-full items-center mb-3"
          >
            <Text className="text-ink-soft font-bold text-sm">Cerrar Sesión</Text>
          </TouchableOpacity>

          <TouchableOpacity 
            onPress={handleDeleteAccount} 
            disabled={loading}
            className="border border-coral/50 py-3 rounded-full items-center"
          >
            <Text className="text-coral-deep font-bold text-sm">Borrar Cuenta Definitivamente</Text>
          </TouchableOpacity>
        </View>

      </ScrollView>
    </SafeAreaView>
  );
}
// aqui acaba control de perfil (juanito)
