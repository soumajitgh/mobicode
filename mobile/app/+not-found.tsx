import { Link } from 'expo-router';
import { Text } from 'react-native';

import { Screen } from '@/shared/components/ui/Screen';

export default function NotFound() {
  return (
    <Screen>
      <Text>Page not found.</Text>
      <Link href="/">Go home</Link>
    </Screen>
  );
}
