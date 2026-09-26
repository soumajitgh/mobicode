import { Link } from 'expo-router';
import { StyleSheet, Text } from 'react-native';

import { Screen } from '@/shared/components/ui/Screen';
import { theme } from '@/shared/theme';

export default function NotFound() {
  return (
    <Screen>
      <Text style={styles.text}>Page not found.</Text>
      <Link href="/" style={styles.link}>
        Go home
      </Link>
    </Screen>
  );
}

const styles = StyleSheet.create({
  text: {
    color: theme.colors.text,
  },
  link: {
    color: theme.colors.primary,
  },
});
