import { View, Text } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Link } from 'expo-router';

export default function TodayScreen() {
  return (
    <SafeAreaView className="flex-1 bg-cream items-center justify-center">
      <Text className="text-3xl font-serif text-ink">Hoy</Text>
      <Text className="text-base text-ink-soft mt-2 text-center font-sans">
        Tus hábitos del día
      </Text>
      <Link href="/(auth)/login" className="mt-8 px-4 py-2 bg-terracotta rounded-full">
        <Text className="text-paper font-bold font-sans">Ir al Auth Demo</Text>
      </Link>
    </SafeAreaView>
  );
}
