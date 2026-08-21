import { createNip98Authorization } from '../identity/nostr';

export type GraphQLResponse<T> = { data?: T; errors?: Array<{ message: string }> };

function configuredBaseURL(): string {
  const baseURL = process.env.EXPO_PUBLIC_API_BASE_URL?.replace(/\/$/, '');
  if (!baseURL) throw new Error('EXPO_PUBLIC_API_BASE_URL is required.');
  return baseURL;
}

export async function signedJSON<T>(path: string, body: unknown, nsec: string, baseURL = configuredBaseURL()): Promise<T> {
  const url = `${baseURL}${path}`;
  const rawBody = JSON.stringify(body);
  const authorization = await createNip98Authorization(url, 'POST', rawBody, nsec);
  const response = await fetch(url, {
    method: 'POST',
    headers: { Authorization: authorization, 'Content-Type': 'application/json' },
    body: rawBody,
  });
  if (!response.ok) throw new Error(`Server authentication failed (${response.status}).`);
  return (await response.json()) as T;
}

export async function queryViewer(nsec: string): Promise<string> {
  const response = await signedJSON<GraphQLResponse<{ viewer: { publicKey: string } }>>('/graphql', { query: 'query Viewer { viewer { publicKey } }' }, nsec);
  if (response.errors?.length || !response.data?.viewer) throw new Error(response.errors?.[0]?.message ?? 'Viewer query failed.');
  return response.data.viewer.publicKey;
}

export async function submitPairing(baseURL: string, pairingToken: string, nsec: string): Promise<'pending_confirmation' | 'complete'> {
  const response = await signedJSON<{ status: 'pending_confirmation' | 'complete' }>('/setup/pair', { pairingToken }, nsec, baseURL.replace(/\/$/, ''));
  return response.status;
}

export async function pairingStatus(baseURL: string, pairingToken: string, nsec: string): Promise<'pending_confirmation' | 'complete'> {
  const response = await signedJSON<{ status: 'pending_confirmation' | 'complete' }>('/setup/pair/status', { pairingToken }, nsec, baseURL.replace(/\/$/, ''));
  return response.status;
}
