import { ApolloClient, HttpLink, InMemoryCache } from '@apollo/client';
import { SetContextLink } from '@apollo/client/link/context';

import { serverURL } from './api';

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

const authLink = new SetContextLink((context) => ({
  headers: {
    ...context.headers,
    ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
  },
}));

export const apolloClient = new ApolloClient({
  link: authLink.concat(new HttpLink({ uri: `${serverURL ?? 'http://localhost:8080'}/mobile/graphql` })),
  cache: new InMemoryCache(),
});
