import { Tabs, Redirect } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { useApp } from '../../src/contexts/AppContext';
import { ActivityIndicator, Text, View, type ColorValue } from 'react-native';
import Svg, { Path } from 'react-native-svg';

function TabIcon({ d, color, size = 24 }: { d: string; color: ColorValue; size?: number }) {
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <Path
        d={d}
        stroke={color}
        strokeWidth={1.8}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </Svg>
  );
}

const ICONS = {
  home: 'M2.25 12l8.954-8.955a1.126 1.126 0 011.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75',
  squad: 'M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z',
  proposals: 'M10.05 4.575a1.575 1.575 0 10-3.15 0v3m3.15-3v-1.5a1.575 1.575 0 013.15 0v1.5m-3.15 0l.075 5.925m3.075.75V4.575m0 0a1.575 1.575 0 013.15 0V15M6.9 7.575a1.575 1.575 0 10-3.15 0v8.175a6.75 6.75 0 0013.5 0v-5.1',
  shame: 'M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z',
};

export default function TabLayout() {
  const { session, loading } = useAuth();
  const { group, wsConnected } = useApp();

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
    <View className="flex-1">
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
        tabBarLabelStyle: {
          fontSize: 10,
          fontWeight: '600',
          letterSpacing: 0.3,
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Hoy',
          tabBarIcon: ({ color }) => <TabIcon d={ICONS.home} color={color} />,
        }}
      />
      <Tabs.Screen
        name="squad"
        options={{
          title: 'Squad',
          tabBarIcon: ({ color }) => <TabIcon d={ICONS.squad} color={color} />,
        }}
      />
      <Tabs.Screen
        name="proposals"
        options={{
          title: 'Votar',
          tabBarIcon: ({ color }) => <TabIcon d={ICONS.proposals} color={color} />,
        }}
      />
      <Tabs.Screen
        name="shame"
        options={{
          title: 'Ruleta',
          tabBarIcon: ({ color }) => <TabIcon d={ICONS.shame} color={color} />,
        }}
      />
    </Tabs>

    {/* Aviso de tiempo real caído (el socket reintenta con backoff) */}
    {group && !wsConnected && (
      <View pointerEvents="none" className="absolute bottom-24 left-0 right-0 items-center">
        <View className="rounded-full bg-amber-soft border border-amber/40 px-4 py-2">
          <Text className="text-xs font-semibold text-amber-deep">
            Sin conexión en vivo — reconectando…
          </Text>
        </View>
      </View>
    )}
    </View>
  );
}
