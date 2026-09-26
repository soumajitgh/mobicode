import { create } from 'zustand';

import { AppError } from '@/api/errors';
import {
  persistSession,
  removeSavedSession,
  restoreSession,
  type Session,
} from '@/features/auth/session';
import { autoPairDevelopmentDevice } from '@/features/pairing/api/auto-pair';

type SessionStatus = 'loading' | 'unauthenticated' | 'authenticated';

type SessionStore = {
  status: SessionStatus;
  error: AppError | null;
  session: Session | null;
  restore: () => Promise<void>;
  signIn: (session: Session) => Promise<void>;
  signOut: () => Promise<void>;
};

let restorePromise: Promise<void> | null = null;

export const useSessionStore = create<SessionStore>((set) => ({
  status: 'loading',
  error: null,
  session: null,

  restore: () => {
    if (restorePromise) return restorePromise;

    restorePromise = (async () => {
      set({ status: 'loading', error: null, session: null });
      let restoredSession: Session | null = null;
      let restoreError: AppError | null = null;
      try {
        restoredSession = await restoreSession();
      } catch (error) {
        restoreError =
          error instanceof AppError
            ? error
            : new AppError(
                'server_unavailable',
                'Could not restore your session. Try again.',
              );
      }

      try {
        const autoPair = restoredSession
          ? null
          : await autoPairDevelopmentDevice();
        const session = restoredSession ?? autoPair?.session ?? null;
        set(
          session
            ? { status: 'authenticated', error: null, session }
            : {
                status: 'unauthenticated',
                error: autoPair?.error ?? restoreError,
                session: null,
              },
        );
      } finally {
        restorePromise = null;
      }
    })();

    return restorePromise;
  },

  signIn: async (session) => {
    await persistSession(session);
    set({ status: 'authenticated', error: null, session });
  },

  signOut: async () => {
    await removeSavedSession();
    set({ status: 'unauthenticated', error: null, session: null });
  },
}));
