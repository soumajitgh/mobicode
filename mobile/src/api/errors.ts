import type { CombinedError } from 'urql';

export type AppErrorCode =
  | 'network_unavailable'
  | 'server_unavailable'
  | 'authentication_failure'
  | 'invalid_pairing_payload'
  | 'graphql_error'
  | 'incompatible_server';

export class AppError extends Error {
  constructor(
    public readonly code: AppErrorCode,
    message: string,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

export function fromGraphQLError(error: CombinedError): AppError {
  const code = error.graphQLErrors[0]?.extensions?.code;
  if (code === 'INVALID_ACCESS_TOKEN' || code === 'UNAUTHENTICATED') {
    return new AppError(
      'authentication_failure',
      'Your mobile session has expired. Pair again.',
    );
  }
  if (
    code === 'INVALID_PAIRING_TOKEN' ||
    code === 'PAIRING_EXPIRED' ||
    code === 'PAIRING_ALREADY_CLAIMED'
  ) {
    return new AppError(
      'invalid_pairing_payload',
      'This pairing code is invalid or has expired.',
    );
  }
  if (error.networkError) {
    return new AppError(
      'network_unavailable',
      'Cannot reach the Mobicode server. Check your connection.',
    );
  }
  return new AppError('graphql_error', error.message);
}
