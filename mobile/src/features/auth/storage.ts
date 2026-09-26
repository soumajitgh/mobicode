import * as SecureStore from 'expo-secure-store';
import { z } from 'zod';

const sessionKey = 'mobicode.mobile.session';
const legacyTokenKey = 'mobicode.mobile.accessToken';

const savedSessionSchema = z.object({
  serverBaseURL: z.url().refine((value) => /^https?:\/\//.test(value)),
  accessToken: z.string().min(1),
});

export type SavedSession = z.infer<typeof savedSessionSchema>;

export async function loadSavedSession(): Promise<SavedSession | null> {
  // The previous token-only format cannot identify its paired server.
  await SecureStore.deleteItemAsync(legacyTokenKey);
  const raw = await SecureStore.getItemAsync(sessionKey);
  if (!raw) return null;
  try {
    return savedSessionSchema.parse(JSON.parse(raw));
  } catch {
    await SecureStore.deleteItemAsync(sessionKey);
    return null;
  }
}

export async function persistSession(session: SavedSession): Promise<void> {
  await SecureStore.setItemAsync(
    sessionKey,
    JSON.stringify(savedSessionSchema.parse(session)),
  );
}

export async function removeSavedSession(): Promise<void> {
  await SecureStore.deleteItemAsync(sessionKey);
  await SecureStore.deleteItemAsync(legacyTokenKey);
}
