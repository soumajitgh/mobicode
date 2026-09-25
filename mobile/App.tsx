import { StatusBar } from 'expo-status-bar';
import { Linking, StyleSheet, Text, View } from 'react-native';

import { GluestackUIProvider } from '@/components/ui/gluestack-ui-provider';
import { Button, ButtonText } from '@/components/ui/button';
import '@/global.css';

export default function App() {
  return (
    <GluestackUIProvider mode="light">
      <View style={styles.container}>
        <Text style={styles.eyebrow}>MOBICODE</Text>
        <Text style={styles.title}>Code from anywhere.</Text>
        <Text style={styles.description}>
          A mobile first coding agent for on the go development.
        </Text>
        <Button onPress={() => void Linking.openURL('https://soumajitgh.github.io/mobicode/')}>
          <ButtonText>Read the docs</ButtonText>
        </Button>
        <StatusBar style="auto" />
      </View>
    </GluestackUIProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 32,
    gap: 16,
    backgroundColor: '#f8fafc',
  },
  eyebrow: {
    color: '#3659a4',
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 3,
  },
  title: {
    color: '#101b35',
    fontSize: 38,
    fontWeight: '700',
    lineHeight: 44,
  },
  description: {
    color: '#475569',
    fontSize: 17,
    lineHeight: 25,
    marginBottom: 12,
  },
});
