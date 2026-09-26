import { createClient, fetchExchange } from 'urql';

import { config } from '@/shared/config';

export function createGraphQLClient(
  serverBaseURL: string,
  accessToken?: string,
) {
  return createClient({
    url: `${serverBaseURL.replace(/\/+$/, '')}${config.graphqlPath}`,
    exchanges: [fetchExchange],
    fetchOptions: () => {
      const headers: Record<string, string> = accessToken
        ? { Authorization: `Bearer ${accessToken}` }
        : {};
      return { headers };
    },
  });
}
