import { useEffect, useState } from 'react';
import { ActivityIndicator, SafeAreaView, StyleSheet, Text } from 'react-native';
import { Redirect, useLocalSearchParams } from 'expo-router';

import { pairingStatus, submitPairing } from '../features/api/client';
import { useIdentity } from '../features/identity/provider';
import { loadOrCreateIdentity } from '../features/identity/store';

export default function PairRoute() {
  const { server, token } = useLocalSearchParams<{ server?: string; token?: string }>();
  const { refresh } = useIdentity();
  const [complete, setComplete] = useState(false);
  const [message, setMessage] = useState('Preparing secure mobile identity…');

  useEffect(() => {
    if (!server || !token) { setMessage('This pairing link is invalid.'); return; }
    let cancelled = false;
    let statusTimer: ReturnType<typeof setInterval> | undefined;
    const pair = async () => {
      try {
        const identity = await loadOrCreateIdentity();
        await submitPairing(server, token, identity.nsec);
        if (cancelled) return;
        setMessage('Waiting for approval in your browser…');
        statusTimer = setInterval(async () => {
          try {
            if ((await pairingStatus(server, token, identity.nsec)) === 'complete') {
              clearInterval(statusTimer);
              await refresh();
              if (!cancelled) setComplete(true);
            }
          } catch { /* keep waiting while the browser confirms */ }
        }, 2000);
      } catch (error) {
        if (!cancelled) setMessage(error instanceof Error ? error.message : 'Pairing failed.');
      }
    };
    void pair();
    return () => { cancelled = true; if (statusTimer) clearInterval(statusTimer); };
  }, [refresh, server, token]);

  if (complete) return <Redirect href="/(app)" />;
  return <SafeAreaView style={styles.screen}><ActivityIndicator color="#7dd3fc" /><Text style={styles.text}>{message}</Text></SafeAreaView>;
}

const styles = StyleSheet.create({ screen: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: '#0b1020', padding: 24 }, text: { color: '#cbd5e1', textAlign: 'center', fontSize: 16, marginTop: 16 } });
