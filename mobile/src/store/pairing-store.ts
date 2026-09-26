import { create } from 'zustand';

import { AppError } from '@/api/errors';
import { claimPairing } from '@/features/pairing/api/claim';
import { useSessionStore } from './session-store';

type PairingStore = {
  payload: string;
  busy: boolean;
  error: string | null;
  setPayload: (payload: string) => void;
  submit: () => Promise<void>;
};

export const usePairingStore = create<PairingStore>((set, get) => ({
  payload: '',
  busy: false,
  error: null,

  setPayload: (payload) => set({ payload, error: null }),

  submit: async () => {
    if (get().busy) return;
    set({ busy: true, error: null });
    try {
      const session = await claimPairing(get().payload);
      await useSessionStore.getState().signIn(session);
      set({ payload: '', error: null });
    } catch (error) {
      set({
        error:
          error instanceof AppError
            ? error.message
            : 'could not complete pairing. try again.',
      });
    } finally {
      set({ busy: false });
    }
  },
}));
