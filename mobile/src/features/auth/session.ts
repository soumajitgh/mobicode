import { AppError, fromGraphQLError } from '@/api/errors';
import { createGraphQLClient } from '@/api/graphql/client';
import {
  ViewerDocument,
  type ViewerQuery,
} from '@/api/graphql/generated/graphql';

import {
  loadSavedSession,
  persistSession,
  removeSavedSession,
  type SavedSession,
} from './storage';

export type Session = SavedSession & {
  user: NonNullable<ViewerQuery['viewer']>;
};

export async function restoreSession(): Promise<Session | null> {
  try {
    const saved = await loadSavedSession();
    if (!saved) return null;

    const result = await createGraphQLClient(
      saved.serverBaseURL,
      saved.accessToken,
    )
      .query(ViewerDocument, {}, { requestPolicy: 'network-only' })
      .toPromise();

    if (result.error) throw fromGraphQLError(result.error);
    if (!result.data?.viewer) {
      throw new AppError(
        'authentication_failure',
        'your mobile session has expired. pair again.',
      );
    }

    return { ...saved, user: result.data.viewer };
  } catch (error) {
    const appError =
      error instanceof AppError
        ? error
        : new AppError(
            'server_unavailable',
            'could not restore your session. try again.',
          );
    if (appError.code === 'authentication_failure') await removeSavedSession();
    throw appError;
  }
}

export { persistSession, removeSavedSession };
