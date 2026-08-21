import { PropsWithChildren, createContext, useContext, useEffect, useMemo, useState } from 'react';

import { loadIdentity } from './store';

type IdentityState = {
  publicKey: string | null;
  loading: boolean;
  refresh: () => Promise<void>;
};

const IdentityContext = createContext<IdentityState | null>(null);

export function IdentityProvider({ children }: PropsWithChildren) {
  const [publicKey, setPublicKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    setLoading(true);
    try {
      const identity = await loadIdentity();
      setPublicKey(identity?.publicKey ?? null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void refresh(); }, []);
  const value = useMemo(() => ({ publicKey, loading, refresh }), [publicKey, loading]);
  return <IdentityContext.Provider value={value}>{children}</IdentityContext.Provider>;
}

export function useIdentity(): IdentityState {
  const state = useContext(IdentityContext);
  if (!state) throw new Error('useIdentity must be used inside IdentityProvider.');
  return state;
}
