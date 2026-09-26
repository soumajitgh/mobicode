import { StyleSheet, Text } from 'react-native';

import { usePairingStore } from '@/store/pairing-store';
import { useSessionStore } from '@/store/session-store';
import { Button, ButtonText } from '@/shared/components/ui/button';
import { Input, InputField } from '@/shared/components/ui/input';
import { Screen } from '@/shared/components/ui/Screen';
import { theme } from '@/shared/theme';

export function PairingScreen() {
  const payload = usePairingStore((state) => state.payload);
  const setPayload = usePairingStore((state) => state.setPayload);
  const busy = usePairingStore((state) => state.busy);
  const error = usePairingStore((state) => state.error);
  const submit = usePairingStore((state) => state.submit);
  const restore = useSessionStore((state) => state.restore);
  const restoreError = useSessionStore((state) => state.error);

  return (
    <Screen>
      <Text style={styles.eyebrow}>MOBICODE</Text>
      <Text style={styles.title}>Pair your device</Text>
      <Text style={styles.description}>
        Paste the pairing link from your Mobicode QR code.
      </Text>
      <Input>
        <InputField
          accessibilityLabel="Pairing link"
          autoCapitalize="none"
          autoCorrect={false}
          onChangeText={setPayload}
          value={payload}
          placeholder="mobicode://pair?..."
        />
      </Input>
      <Button isDisabled={busy} onPress={() => void submit()}>
        <ButtonText>{busy ? 'Pairing…' : 'Pair device'}</ButtonText>
      </Button>
      {(error || restoreError) && (
        <Text style={styles.error}>{error ?? restoreError?.message}</Text>
      )}
      {restoreError && (
        <Button onPress={() => void restore()}>
          <ButtonText>Retry session restore</ButtonText>
        </Button>
      )}
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
  error: { color: theme.colors.danger, fontSize: theme.typography.body },
});
