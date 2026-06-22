import '../../global.css';
import { StatusBar } from 'expo-status-bar';
import React, { useEffect, useState } from 'react';
import { Text, View, TouchableOpacity, ActivityIndicator, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { supabase } from '../../lib/supabase';

export default function App() {
  const [apiStatus, setApiStatus] = useState<'idle' | 'checking' | 'ok' | 'error'>('idle');
  const [supabaseStatus, setSupabaseStatus] = useState<'idle' | 'checking' | 'ok' | 'error'>('idle');
  const [apiResponse, setApiResponse] = useState<string>('');
  const [supabaseDetails, setSupabaseDetails] = useState<string>('');

  const checkConnectivity = async () => {
    // 1. Check API Gateway
    setApiStatus('checking');
    setApiResponse('Consultando...');
    try {
      const apiUrl = process.env.EXPO_PUBLIC_API_URL || 'https://dydi-25hj.onrender.com';
      const controller = new AbortController();
      const id = setTimeout(() => controller.abort(), 8000); // 8s timeout

      const res = await fetch(`${apiUrl}/health`, { signal: controller.signal });
      clearTimeout(id);

      if (res.ok) {
        const text = await res.text();
        setApiStatus('ok');
        setApiResponse(text || 'API activa (OK)');
      } else {
        setApiStatus('error');
        setApiResponse(`Error HTTP: ${res.status}`);
      }
    } catch (err: any) {
      setApiStatus('error');
      setApiResponse(err.message || 'Error de red / Timeout');
    }

    // 2. Check Supabase
    setSupabaseStatus('checking');
    setSupabaseDetails('Consultando...');
    try {
      const { data, error } = await supabase.auth.getSession();
      if (error) {
        setSupabaseStatus('error');
        setSupabaseDetails(error.message);
      } else {
        setSupabaseStatus('ok');
        setSupabaseDetails(data.session ? 'Sesión existente' : 'Conectado (Sin sesión)');
      }
    } catch (err: any) {
      setSupabaseStatus('error');
      setSupabaseDetails(err.message || 'Error de conexión');
    }
  };

  useEffect(() => {
    checkConnectivity();
  }, []);

  return (
    <SafeAreaView className="flex-1 bg-cream">
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={{ flexGrow: 1 }} className="px-6 py-8">
        {/* Header */}
        <View className="items-center my-8">
          <Text className="text-5xl font-serif text-terracotta tracking-tight">Dydi</Text>
          <Text className="text-xs font-bold text-ink-soft tracking-[0.15em] uppercase mt-2">
            Accountability Social · Mobile
          </Text>
        </View>

        {/* Info Card */}
        <View className="bg-surface border border-hairline rounded-3xl p-6 shadow-sm mb-6">
          <Text className="text-lg font-bold text-ink mb-2">Diagnóstico de Conectividad</Text>
          <Text className="text-sm text-ink-soft mb-4">
            Esta pantalla verifica la conexión directa con el backend y base de datos en producción.
          </Text>

          {/* API Gateway Status */}
          <View className="mb-4 pb-4 border-b border-hairline/60">
            <View className="flex-row justify-between items-center mb-1">
              <Text className="font-bold text-sm text-ink-soft uppercase tracking-wider">
                API Gateway (Go)
              </Text>
              <View className="flex-row items-center">
                {apiStatus === 'checking' && <ActivityIndicator size="small" color="#7CA39D" />}
                {apiStatus === 'ok' && (
                  <View className="w-2.5 h-2.5 rounded-full bg-sage mr-1.5 animate-pulse" />
                )}
                {apiStatus === 'error' && (
                  <View className="w-2.5 h-2.5 rounded-full bg-coral mr-1.5" />
                )}
                <Text className="text-xs font-semibold capitalize text-ink">
                  {apiStatus === 'idle' ? 'Inactivo' : apiStatus === 'checking' ? 'Buscando...' : apiStatus}
                </Text>
              </View>
            </View>
            <Text className="text-xs font-mono text-ink-soft bg-paper p-2 rounded-lg border border-hairline/40">
              {process.env.EXPO_PUBLIC_API_URL}
            </Text>
            <Text className="text-xs text-ink-soft mt-1">
              Respuesta: <Text className="font-semibold text-ink">{apiResponse}</Text>
            </Text>
          </View>

          {/* Supabase Status */}
          <View className="mb-2">
            <View className="flex-row justify-between items-center mb-1">
              <Text className="font-bold text-sm text-ink-soft uppercase tracking-wider">
                Supabase Auth & DB
              </Text>
              <View className="flex-row items-center">
                {supabaseStatus === 'checking' && <ActivityIndicator size="small" color="#7CA39D" />}
                {supabaseStatus === 'ok' && (
                  <View className="w-2.5 h-2.5 rounded-full bg-sage mr-1.5" />
                )}
                {supabaseStatus === 'error' && (
                  <View className="w-2.5 h-2.5 rounded-full bg-coral mr-1.5" />
                )}
                <Text className="text-xs font-semibold capitalize text-ink">
                  {supabaseStatus === 'idle' ? 'Inactivo' : supabaseStatus === 'checking' ? 'Buscando...' : supabaseStatus}
                </Text>
              </View>
            </View>
            <Text className="text-xs font-mono text-ink-soft bg-paper p-2 rounded-lg border border-hairline/40">
              {process.env.EXPO_PUBLIC_SUPABASE_URL}
            </Text>
            <Text className="text-xs text-ink-soft mt-1">
              Estado: <Text className="font-semibold text-ink">{supabaseDetails}</Text>
            </Text>
          </View>
        </View>

        {/* UI Demo Preview Card */}
        <View className="bg-surface border border-hairline rounded-3xl p-6 shadow-sm mb-6">
          <Text className="text-lg font-bold text-ink mb-1">Diseño Dydi Preview</Text>
          <Text className="text-xs text-ink-soft mb-4">
            Muestra de estilos NativeWind
          </Text>

          <View className="flex-row justify-between items-center bg-paper p-4 rounded-2xl border border-hairline/50">
            <View>
              <Text className="text-eyebrow text-ink-soft text-[10px] tracking-wider uppercase font-bold">
                Tu Racha Actual
              </Text>
              <Text className="text-4xl font-serif text-terracotta mt-1">13</Text>
              <Text className="text-xs text-ink-soft">días de racha</Text>
            </View>
            <View className="bg-wash p-3 rounded-full">
              <Text className="text-sage text-2xl">🔥</Text>
            </View>
          </View>
        </View>

        {/* Buttons */}
        <TouchableOpacity
          onPress={checkConnectivity}
          className="bg-sage-deep py-4 rounded-full items-center justify-center active:opacity-90 shadow-sm"
        >
          <Text className="text-paper font-bold text-base">Probar Conexión</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}
