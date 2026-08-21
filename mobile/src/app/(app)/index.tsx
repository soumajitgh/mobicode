import { SafeAreaView, StyleSheet, Text } from 'react-native';

import { useIdentity } from '../../features/identity/provider';

export default function WorkspaceRoute() {
  const { publicKey } = useIdentity();
  return <SafeAreaView style={styles.screen}><Text style={styles.label}>MOBICODE</Text><Text style={styles.title}>Mobile identity paired</Text><Text style={styles.description}>{publicKey}</Text></SafeAreaView>;
}

const styles = StyleSheet.create({ screen: { flex: 1, justifyContent: 'center', backgroundColor: '#0b1020', padding: 28 }, label: { color: '#7dd3fc', fontWeight: '700', letterSpacing: 3 }, title: { color: '#f8fafc', fontSize: 32, fontWeight: '700', marginTop: 16 }, description: { color: '#94a3b8', fontSize: 13, marginTop: 16 } });
