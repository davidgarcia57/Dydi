import { Tabs, Redirect } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { ActivityIndicator, View } from 'react-native';

export default function TabLayout() {
  const { session, loading } = useAuth();

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center bg-cream">
        <ActivityIndicator size="large" color="#7CA39D" />
      </View>
    );
  }

  if (!session) {
    return <Redirect href="/(auth)/login" />;
  }

  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#C26F4D', // terracotta
        tabBarInactiveTintColor: '#A89C89', // ink-faint
        tabBarStyle: {
          backgroundColor: '#FCF9F3', // surface
          borderTopWidth: 1,
          borderTopColor: '#E7DECD', // hairline
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Today',
        }}
      />
      <Tabs.Screen
        name="squad"
        options={{
          title: 'Squad',
        }}
      />
      <Tabs.Screen
        name="proposals"
        options={{
          title: 'Proposals',
        }}
      />
      <Tabs.Screen
        name="shame"
        options={{
          title: 'Shame',
        }}
      />
    </Tabs>
  );
}
