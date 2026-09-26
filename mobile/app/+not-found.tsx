import { Link } from 'expo-router';
import { StyleSheet, Text } from 'react-native';

import { Screen } from '@/shared/components/ui/Screen';
import { theme } from '@/shared/theme';

export default function NotFound() {
  return (
    <Screen>
      <Text style={styles.text}>page not found.</Text>
      <Link href="/" style={styles.link}>
        go home
      </Link>
    </Screen>
  );
}

const styles = StyleSheet.create({
  text: {
    color: theme.colors.text,
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    letterSpacing: 1,
  },
  link: {
    color: theme.colors.primary,
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    letterSpacing: 1,
  },
});
