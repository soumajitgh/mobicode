import { AppError, fromGraphQLError } from '@/api/errors';
import { createGraphQLClient } from '@/api/graphql/client';
import { Platform } from 'react-native';

import { ClaimDevicePairingDocument } from '@/api/graphql/generated/graphql';
import { parsePairingPayload } from '../schemas';

export async function claimPairing(payload: string) {
  const { serverBaseURL, token } = parsePairingPayload(payload);
  const result = await createGraphQLClient(serverBaseURL)
    .mutation(ClaimDevicePairingDocument, {
      input: {
        token,
        device: {
          name: 'Mobicode mobile device',
          platform: Platform.OS === 'ios' ? 'IOS' : 'ANDROID',
        },
      },
    })
    .toPromise();
  if (result.error) throw fromGraphQLError(result.error);
  if (!result.data?.claimDevicePairing)
    throw new AppError(
      'incompatible_server',
      'This server did not return a mobile session.',
    );
  return {
    serverBaseURL,
    accessToken: result.data.claimDevicePairing.accessToken,
    user: result.data.claimDevicePairing.user,
  };
}
