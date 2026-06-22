import { Tabs } from 'expo-router';

export default function TabLayout() {
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
