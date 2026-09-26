import { StatusBar } from 'expo-status-bar';
import { ApolloProvider } from '@apollo/client/react';
import { QueryClient, QueryClientProvider, focusManager, useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { AppState, Linking, Platform, StyleSheet, Text, View } from 'react-native';

import { GluestackUIProvider } from '@/components/ui/gluestack-ui-provider';
import { Button, ButtonText } from '@/components/ui/button';
import '@/global.css';
import { restoreOrAutoPair } from './session';
import { apolloClient } from './src/lib/apollo';

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

export default function App() {
  useEffect(() => {
    if (Platform.OS === 'web') return;
    const subscription = AppState.addEventListener('change', (state) => {
      focusManager.setFocused(state === 'active');
    });
    return () => subscription.remove();
  }, []);

  return (
    <ApolloProvider client={apolloClient}>
      <QueryClientProvider client={queryClient}>
        <GluestackUIProvider mode="light">
          <Home />
        </GluestackUIProvider>
      </QueryClientProvider>
    </ApolloProvider>
  );
}

function Home() {
  const { data: user } = useQuery({
    queryKey: ['mobile', 'session'],
    queryFn: restoreOrAutoPair,
    staleTime: Infinity,
  });

  return (
    <>
      <View style={styles.container}>
        <Text style={styles.eyebrow}>MOBICODE</Text>
        <Text style={styles.title}>{user ? 'Device paired.' : 'Code from anywhere.'}</Text>
        <Text style={styles.description}>
          {user ? `Signed in as ${user.email}` : 'A mobile first coding agent for on the go development.'}
        </Text>
        <Button onPress={() => void Linking.openURL('https://soumajitgh.github.io/mobicode/')}>
          <ButtonText>Read the docs</ButtonText>
        </Button>
      </View>
      <StatusBar style="auto" />
    </>
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
