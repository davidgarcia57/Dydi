import { View, Text } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function ShameScreen() {
  return (
    <SafeAreaView className="flex-1 bg-cream items-center justify-center">
      <Text className="text-3xl font-serif text-ink">Vergüenza</Text>
      <Text className="text-base text-ink-soft mt-2 text-center font-sans">
        Muro de la vergüenza y ruleta
      </Text>
    </SafeAreaView>
  );
}
