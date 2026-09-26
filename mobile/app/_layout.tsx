import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { ActivityIndicator } from 'react-native';
import { useEffect } from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { useSessionStore } from '@/store/session-store';
import { Screen } from '@/shared/components/ui/Screen';
import { GluestackUIProvider } from '@/shared/components/ui/gluestack-ui-provider';
import { theme } from '@/shared/theme';
import '../global.css';

function Navigation() {
  const status = useSessionStore((state) => state.status);
  if (status === 'loading') {
    return (
      <Screen>
        <ActivityIndicator
          accessibilityLabel="restoring session"
          color={theme.colors.primary}
        />
      </Screen>
    );
  }
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Protected guard={status === 'unauthenticated'}>
        <Stack.Screen name="(auth)" />
      </Stack.Protected>
      <Stack.Protected guard={status === 'authenticated'}>
        <Stack.Screen name="(app)" />
      </Stack.Protected>
      <Stack.Screen name="+not-found" />
    </Stack>
  );
}

export default function RootLayout() {
  const restore = useSessionStore((state) => state.restore);

  useEffect(() => {
    void restore();
  }, [restore]);

  return (
    <SafeAreaProvider>
      <GluestackUIProvider mode="dark">
        <Navigation />
        <StatusBar style="light" />
      </GluestackUIProvider>
    </SafeAreaProvider>
  );
}
