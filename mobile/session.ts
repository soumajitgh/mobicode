import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const tokenKey = 'mobicode.mobile.accessToken';
const serverURL = process.env.EXPO_PUBLIC_MOBICODE_SERVER_URL?.replace(/\/$/, '');

export type PairedUser = { id: string; email: string };

async function viewer(token: string): Promise<PairedUser | null> {
  const response = await fetch(`${serverURL}/mobile/graphql`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ query: '{ viewer { id email } }' }),
  });
  if (!response.ok) throw new Error('Could not check mobile session');
  const result: { data?: { viewer?: PairedUser | null }; errors?: unknown[] } = await response.json();
  if (result.errors?.length) return null;
  return result.data?.viewer ?? null;
}

export async function restoreOrAutoPair(): Promise<PairedUser | null> {
  if (!serverURL || Platform.OS === 'web') return null;

  const savedToken = await SecureStore.getItemAsync(tokenKey);
  if (savedToken) {
    try {
      const user = await viewer(savedToken);
      if (user) return user;
      await SecureStore.deleteItemAsync(tokenKey);
    } catch {
      // Keep the token when the server cannot be reached.
      return null;
    }
  }

  const response = await fetch(`${serverURL}/mobile/dev/auto-pair`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ platform: Platform.OS === 'ios' ? 'IOS' : 'ANDROID' }),
  });
  if (!response.ok) return null;
  const session: { accessToken?: string; user?: PairedUser } = await response.json();
  if (!session.accessToken || !session.user?.email) return null;
  await SecureStore.setItemAsync(tokenKey, session.accessToken);
  return session.user;
}
