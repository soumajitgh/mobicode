import Constants from 'expo-constants';
import { Platform } from 'react-native';
import { z } from 'zod';

import { persistSession, type Session } from '@/features/auth/session';

const autoPairResponse = z.object({
  accessToken: z.string().min(1),
  user: z.object({
    id: z.union([z.string(), z.number().int().nonnegative()]).transform(String),
    email: z.email(),
  }),
});

function developmentServerURL(): string | null {
  const configured = process.env.EXPO_PUBLIC_MOBICODE_SERVER_URL;
  if (configured) {
    try {
      const url = new URL(configured);
      if (url.protocol === 'http:' || url.protocol === 'https:') {
        return url.origin;
      }
    } catch {
      return null;
    }
    return null;
  }

  const hostUri = Constants.expoConfig?.hostUri;
  if (!hostUri) return null;
  try {
    const metroHost = new URL(`http://${hostUri}`).hostname;
    const host =
      Platform.OS === 'android' && metroHost === 'localhost'
        ? '10.0.2.2'
        : metroHost;
    const port = process.env.EXPO_PUBLIC_MOBICODE_SERVER_PORT ?? '8080';
    return `http://${host}:${port}`;
  } catch {
    return null;
  }
}

export async function autoPairDevelopmentDevice(): Promise<Session | null> {
  if (!__DEV__) return null;
  const serverBaseURL = developmentServerURL();
  if (!serverBaseURL) return null;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 5000);
  try {
    const response = await fetch(`${serverBaseURL}/mobile/dev/auto-pair`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        platform: Platform.OS === 'ios' ? 'IOS' : 'ANDROID',
      }),
      signal: controller.signal,
    });
    if (!response.ok) return null;
    const data = autoPairResponse.parse(await response.json());
    const session = { serverBaseURL, ...data };
    await persistSession(session);
    return session;
  } catch {
    return null;
  } finally {
    clearTimeout(timeout);
  }
}
