import { ActivityIndicator, SafeAreaView, StyleSheet, Text } from 'react-native';
import { Redirect } from 'expo-router';

import { useIdentity } from '../features/identity/provider';

export default function IndexRoute() {
  const { publicKey, loading } = useIdentity();
  if (loading) return <SafeAreaView style={styles.screen}><ActivityIndicator color="#7dd3fc" /><Text style={styles.text}>Unlocking signing identity…</Text></SafeAreaView>;
  if (publicKey) return <Redirect href="/(app)" />;
  return <SafeAreaView style={styles.screen}><Text style={styles.text}>Open the server setup QR code to pair this mobile.</Text></SafeAreaView>;
}

const styles = StyleSheet.create({ screen: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: '#0b1020', padding: 24 }, text: { color: '#cbd5e1', textAlign: 'center', fontSize: 16, marginTop: 16 } });
