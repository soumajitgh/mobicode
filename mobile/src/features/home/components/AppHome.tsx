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
      <Text style={styles.eyebrow}>MOBICODE</Text>
      <Text style={styles.title}>Device paired.</Text>
      <Text style={styles.description}>Signed in as {session?.user.email}</Text>
      <Button onPress={() => void signOut()}>
        <ButtonText>Unpair device</ButtonText>
      </Button>
    </Screen>
  );
}

const styles = StyleSheet.create({
  eyebrow: {
    color: theme.colors.primary,
    fontSize: theme.typography.caption,
    fontWeight: '700',
    letterSpacing: 3,
  },
  title: {
    color: theme.colors.text,
    fontSize: theme.typography.title,
    fontWeight: '700',
  },
  description: { color: theme.colors.muted, fontSize: theme.typography.body },
});
