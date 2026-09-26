import { CombinedGraphQLErrors } from '@apollo/client/errors';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';
import { z } from 'zod';

import { ViewerDocument } from './src/graphql/generated/graphql';
import { api, serverURL } from './src/lib/api';
import { apolloClient, setAccessToken } from './src/lib/apollo';

const tokenKey = 'mobicode.mobile.accessToken';
const sessionSchema = z.object({
  accessToken: z.string().min(1),
  user: z.object({ id: z.coerce.string(), email: z.email() }),
});

export type PairedUser = z.infer<typeof sessionSchema>['user'];

export async function restoreOrAutoPair(): Promise<PairedUser | null> {
  if (!serverURL || Platform.OS === 'web') return null;

  const savedToken = await SecureStore.getItemAsync(tokenKey);
  if (savedToken) {
    setAccessToken(savedToken);
    try {
      const result = await apolloClient.query({ query: ViewerDocument, fetchPolicy: 'no-cache' });
      if (result.data?.viewer) return result.data.viewer;
    } catch (error) {
      const invalidToken = CombinedGraphQLErrors.is(error) &&
        error.errors.some((item) => item.extensions?.code === 'INVALID_ACCESS_TOKEN');
      if (!invalidToken) return null;
    }
    setAccessToken(null);
    await SecureStore.deleteItemAsync(tokenKey);
  }

  const { data } = await api.post<unknown>('/mobile/dev/auto-pair', {
    platform: Platform.OS === 'ios' ? 'IOS' : 'ANDROID',
  });
  const session = sessionSchema.parse(data);
  await SecureStore.setItemAsync(tokenKey, session.accessToken);
  setAccessToken(session.accessToken);
  return session.user;
}
