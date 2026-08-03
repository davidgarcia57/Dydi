import '../global.css';
import { Stack } from 'expo-router';
import { useFonts } from 'expo-font';
import {
  Newsreader_400Regular,
  Newsreader_700Bold,
  Newsreader_400Regular_Italic
} from '@expo-google-fonts/newsreader';
import {
  HankenGrotesk_400Regular,
  HankenGrotesk_700Bold,
  HankenGrotesk_600SemiBold
} from '@expo-google-fonts/hanken-grotesk';
import { useEffect } from 'react';
import { View, ActivityIndicator } from 'react-native';

import { AuthProvider } from '../src/contexts/AuthContext';
import { AppProvider } from '../src/contexts/AppContext';
import ServerWakeup from '../src/components/ServerWakeup';

export default function RootLayout() {
  const [loaded, error] = useFonts({
    Newsreader: Newsreader_400Regular,
    NewsreaderBold: Newsreader_700Bold,
    NewsreaderItalic: Newsreader_400Regular_Italic,
    HankenGrotesk: HankenGrotesk_400Regular,
    HankenGroteskSemiBold: HankenGrotesk_600SemiBold,
    HankenGroteskBold: HankenGrotesk_700Bold,
  });

  useEffect(() => {
    if (error) {
      console.error(error);
    }
  }, [error]);

  if (!loaded) {
    return (
      <View className="flex-1 items-center justify-center bg-cream">
        <ActivityIndicator size="large" color="#7CA39D" />
      </View>
    );
  }

  return (
    <AuthProvider>
      <AppProvider>
        <Stack>
          <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
          <Stack.Screen name="(auth)" options={{ headerShown: false }} />
          <Stack.Screen name="onboarding" options={{ headerShown: false }} />
          {/* (modals) no es una ruta: el directorio no tiene _layout, asi que
              expo-router solo conoce las pantallas de dentro. Declararlo como
              grupo tiraba "No route named (modals) exists in nested children". */}
          <Stack.Screen
            name="(modals)/checkin"
            options={{ presentation: 'modal', headerShown: false }}
          />
          <Stack.Screen
            name="(modals)/profile"
            options={{ presentation: 'modal', headerShown: false }}
          />
        </Stack>
        {/* Despues del Stack a proposito: en RN el hermano posterior pinta
            encima, y este splash tiene que cubrir tambien el login — si el
            gateway esta dormido, no sirve dejar al usuario teclear y esperar. */}
        <ServerWakeup />
      </AppProvider>
    </AuthProvider>
  );
}
