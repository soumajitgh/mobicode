import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';

import { IdentityProvider } from '../features/identity/provider';

export default function RootLayout() {
  return (
    <IdentityProvider>
      <StatusBar style="light" />
      <Stack screenOptions={{ headerShown: false }} />
    </IdentityProvider>
  );
}
