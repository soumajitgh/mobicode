import { StyleSheet, Text } from 'react-native';

import { useSessionStore } from '@/store/session-store';
import { Button, ButtonText } from '@/shared/components/ui/button';
import { Screen } from '@/shared/components/ui/Screen';
import { theme } from '@/shared/theme';

export function AppHome() {
  const session = useSessionStore((state) => state.session);
  const signOut = useSessionStore((state) => state.signOut);
  return (
    <Screen>
      <Text style={styles.eyebrow}>mobicode</Text>
      <Text style={styles.title}>device paired.</Text>
      <Text style={styles.description}>signed in as {session?.user.email}</Text>
      <Button onPress={() => void signOut()}>
        <ButtonText style={styles.buttonText}>unpair device</ButtonText>
      </Button>
    </Screen>
  );
}

const styles = StyleSheet.create({
  eyebrow: {
    color: theme.colors.primary,
    fontSize: theme.typography.caption,
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontWeight: '700',
    letterSpacing: 2,
  },
  title: {
    color: theme.colors.text,
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontSize: 26,
    fontWeight: '700',
    letterSpacing: 1.5,
  },
  description: {
    color: theme.colors.muted,
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontSize: 14,
    letterSpacing: 0.5,
  },
  buttonText: {
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    letterSpacing: 1,
  },
});
