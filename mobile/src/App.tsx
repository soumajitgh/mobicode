import { StatusBar } from 'expo-status-bar';
import { SafeAreaView, StyleSheet, Text, View } from 'react-native';

export default function App() {
  return (
    <SafeAreaView style={styles.screen}>
      <StatusBar style="light" />
      <View style={styles.content}>
        <Text style={styles.eyebrow}>MOBICODE</Text>
        <Text style={styles.title}>Your coding workspace, wherever you are.</Text>
        <Text style={styles.description}>
          The mobile client is ready to connect to the MobiCode Go API.
        </Text>
        <View style={styles.status}>
          <View style={styles.statusDot} />
          <Text style={styles.statusText}>Mobile workspace initialized</Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: '#0b1020',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 28,
  },
  eyebrow: {
    color: '#7dd3fc',
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 3,
    marginBottom: 18,
  },
  title: {
    color: '#f8fafc',
    fontSize: 42,
    fontWeight: '700',
    lineHeight: 48,
  },
  description: {
    color: '#94a3b8',
    fontSize: 18,
    lineHeight: 28,
    marginTop: 20,
  },
  status: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    backgroundColor: '#111a31',
    borderColor: '#243354',
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: 'row',
    marginTop: 36,
    paddingHorizontal: 16,
    paddingVertical: 11,
  },
  statusDot: {
    backgroundColor: '#34d399',
    borderRadius: 5,
    height: 10,
    marginRight: 10,
    width: 10,
  },
  statusText: {
    color: '#cbd5e1',
    fontSize: 14,
    fontWeight: '600',
  },
});
