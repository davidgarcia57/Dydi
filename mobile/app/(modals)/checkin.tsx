import { View, Text, TouchableOpacity } from 'react-native';
import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function CheckinModal() {
  const router = useRouter();

  return (
    <SafeAreaView className="flex-1 bg-surface items-center justify-center">
      <Text className="text-3xl font-serif text-ink">Check-in</Text>
      <Text className="text-base text-ink-soft mt-2 text-center font-sans px-4">
        ¿Completaste el hábito?
      </Text>
      <TouchableOpacity 
        onPress={() => router.back()}
        className="mt-8 px-6 py-3 bg-sage-deep rounded-full shadow-sm"
      >
        <Text className="text-paper font-bold font-sans">Cerrar</Text>
      </TouchableOpacity>
    </SafeAreaView>
  );
}
