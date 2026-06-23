import '../../global.css';
import { StatusBar } from 'expo-status-bar';
import React, { useState } from 'react';
import {
  Text,
  View,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { supabase } from '../../lib/supabase';
import { useRouter } from 'expo-router';

export default function LoginScreen() {
  const router = useRouter();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  // Form fields
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const isRegister = mode === 'register';

  function resetFeedback() {
    setErrorMessage('');
    setSuccessMessage('');
  }

  function switchMode() {
    setMode(isRegister ? 'login' : 'register');
    setPassword('');
    setConfirmPassword('');
    resetFeedback();
  }

  function translateAuthError(error: any) {
    const msg = error?.message?.toLowerCase() ?? '';
    if (
      msg.includes('email_taken') ||
      msg.includes('already registered') ||
      msg.includes('user already')
    ) {
      return 'Ese correo ya está registrado. Inicia sesión.';
    }
    if (msg.includes('invalid login credentials')) {
      return 'El correo o la contraseña no coinciden.';
    }
    if (msg.includes('email')) {
      return 'Revisa que el correo esté bien escrito.';
    }
    if (msg.includes('password')) {
      return 'La contraseña debe tener al menos 6 caracteres.';
    }
    return error?.message || 'Algo salió mal. Intenta de nuevo.';
  }

  const submit = async () => {
    resetFeedback();

    if (isRegister && password !== confirmPassword) {
      setErrorMessage('Las contraseñas no coinciden.');
      return;
    }

    setLoading(true);

    try {
      if (isRegister) {
        const { data, error } = await supabase.auth.signUp({
          email: email.trim(),
          password,
          options: {
            data: {
              display_name: displayName.trim(),
            },
          },
        });

        if (error) throw error;

        // Anti-enumeration check (same as web app)
        if (data.user && Array.isArray(data.user.identities) && data.user.identities.length === 0) {
          throw new Error('email_taken');
        }

        if (data.session) {
          router.replace('/(tabs)');
        } else {
          setSuccessMessage('Cuenta creada. Revisa tu correo para confirmar el acceso.');
        }
      } else {
        const { data, error } = await supabase.auth.signInWithPassword({
          email: email.trim(),
          password,
        });

        if (error) throw error;

        if (data.session) {
          router.replace('/(tabs)');
        }
      }
    } catch (error: any) {
      setErrorMessage(translateAuthError(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-cream">
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        className="flex-1"
      >
        <ScrollView
          contentContainerStyle={{ flexGrow: 1 }}
          className="px-6 py-8"
          keyboardShouldPersistTaps="handled"
        >
          {/* Header/Logo */}
          <View className="items-center mt-6 mb-8">
            <Text className="text-5xl font-serif text-terracotta tracking-tight">Dydi</Text>
            <Text className="text-xs font-bold text-ink-soft tracking-[0.15em] uppercase mt-2">
              Hábitos con consecuencias
            </Text>
          </View>

          {/* Card Form Container */}
          <View className="bg-paper border border-hairline rounded-3xl p-6 shadow-sm mb-6">
            
            {/* Tab Switcher */}
            <View className="flex-row bg-cream-2 rounded-full p-1 mb-6">
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => {
                  setMode('login');
                  resetFeedback();
                }}
                className={`flex-1 py-2.5 rounded-full items-center ${
                  !isRegister ? 'bg-paper shadow-sm' : ''
                }`}
              >
                <Text className={`text-sm font-bold ${!isRegister ? 'text-ink' : 'text-ink-soft'}`}>
                  Entrar
                </Text>
              </TouchableOpacity>
              
              <TouchableOpacity
                activeOpacity={0.8}
                onPress={() => {
                  setMode('register');
                  resetFeedback();
                }}
                className={`flex-1 py-2.5 rounded-full items-center ${
                  isRegister ? 'bg-paper shadow-sm' : ''
                }`}
              >
                <Text className={`text-sm font-bold ${isRegister ? 'text-ink' : 'text-ink-soft'}`}>
                  Registro
                </Text>
              </TouchableOpacity>
            </View>

            {/* Form Title & Subtitle */}
            <View className="mb-5">
              <Text className="font-serif text-2xl font-semibold text-ink leading-none">
                {isRegister ? 'Crea tu cuenta' : 'Inicia sesión'}
              </Text>
              <Text className="text-xs text-ink-soft mt-1.5 leading-snug">
                {isRegister
                  ? 'Únete a tu grupo y empieza el reto.'
                  : 'Vuelve con tu squad y marca el día.'}
              </Text>
            </View>

            {/* Inputs */}
            <View className="flex-col gap-4">
              {isRegister && (
                <View className="flex-col gap-1.5">
                  <Text className="text-xs font-bold text-ink">Nombre</Text>
                  <TextInput
                    className="w-full bg-surface border border-hairline rounded-xl px-4 py-3 text-[15px] text-ink font-sans"
                    placeholder="Tu nombre o apodo"
                    placeholderTextColor="#A89C89"
                    value={displayName}
                    onChangeText={setDisplayName}
                    autoCapitalize="words"
                    autoCorrect={false}
                  />
                </View>
              )}

              <View className="flex-col gap-1.5">
                <Text className="text-xs font-bold text-ink">Correo</Text>
                <TextInput
                  className="w-full bg-surface border border-hairline rounded-xl px-4 py-3 text-[15px] text-ink font-sans"
                  placeholder="tu@correo.com"
                  placeholderTextColor="#A89C89"
                  value={email}
                  onChangeText={setEmail}
                  autoCapitalize="none"
                  keyboardType="email-address"
                  autoCorrect={false}
                />
              </View>

              <View className="flex-col gap-1.5">
                <Text className="text-xs font-bold text-ink">Contraseña</Text>
                <TextInput
                  className="w-full bg-surface border border-hairline rounded-xl px-4 py-3 text-[15px] text-ink font-sans"
                  placeholder="Mínimo 6 caracteres"
                  placeholderTextColor="#A89C89"
                  value={password}
                  onChangeText={setPassword}
                  secureTextEntry
                  autoCapitalize="none"
                  autoCorrect={false}
                />
              </View>

              {isRegister && (
                <View className="flex-col gap-1.5">
                  <Text className="text-xs font-bold text-ink">Confirmar contraseña</Text>
                  <TextInput
                    className="w-full bg-surface border border-hairline rounded-xl px-4 py-3 text-[15px] text-ink font-sans"
                    placeholder="Repítela una vez"
                    placeholderTextColor="#A89C89"
                    value={confirmPassword}
                    onChangeText={setConfirmPassword}
                    secureTextEntry
                    autoCapitalize="none"
                    autoCorrect={false}
                  />
                </View>
              )}

              {/* Feedback messages */}
              {errorMessage ? (
                <View className="bg-coral-soft border border-coral/30 rounded-xl px-4 py-3 mt-1">
                  <Text className="text-xs font-medium text-coral-deep">{errorMessage}</Text>
                </View>
              ) : null}

              {successMessage ? (
                <View className="bg-sage-soft border border-sage/30 rounded-xl px-4 py-3 mt-1">
                  <Text className="text-xs font-medium text-sage-deep">{successMessage}</Text>
                </View>
              ) : null}

              {/* Submit Button */}
              <TouchableOpacity
                onPress={submit}
                disabled={loading}
                activeOpacity={0.9}
                className="w-full bg-sage-deep py-4 rounded-full items-center justify-center mt-2"
              >
                {loading ? (
                  <ActivityIndicator size="small" color="#FFFFFF" />
                ) : (
                  <Text className="text-paper font-bold text-sm">
                    {isRegister ? 'Crear cuenta' : 'Entrar'}
                  </Text>
                )}
              </TouchableOpacity>
            </View>
          </View>

          {/* Bottom Switch Mode Link */}
          <View className="flex-row justify-center items-center mt-2">
            <Text className="text-xs text-ink-soft">
              {isRegister ? '¿Ya tienes cuenta?' : '¿Eres nuevo?'}
            </Text>
            <TouchableOpacity onPress={switchMode} className="ml-1">
              <Text className="text-xs font-bold text-sage-deep">
                {isRegister ? 'Inicia sesión' : 'Crea una cuenta'}
              </Text>
            </TouchableOpacity>
          </View>

        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
