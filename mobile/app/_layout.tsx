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
      <Stack>
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen name="(auth)" options={{ headerShown: false }} />
        <Stack.Screen name="(modals)" options={{ presentation: 'modal', headerShown: false }} />
      </Stack>
    </AuthProvider>
  );
}
