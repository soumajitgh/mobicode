import * as SecureStore from 'expo-secure-store';

import { createIdentity, MobileIdentity, publicKeyFromNsec } from './nostr';

const identityKey = 'mobicode.nostr.nsec';
const secureOptions: SecureStore.SecureStoreOptions = {
  keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
  requireAuthentication: true,
};

export async function loadIdentity(): Promise<MobileIdentity | null> {
	const developmentSeed = getDevelopmentSeed();
	if (developmentSeed) return { nsec: developmentSeed, publicKey: publicKeyFromNsec(developmentSeed) };
	const nsec = await SecureStore.getItemAsync(identityKey, secureOptions);
  return nsec ? { nsec, publicKey: publicKeyFromNsec(nsec) } : null;
}

export async function loadOrCreateIdentity(): Promise<MobileIdentity> {
	const developmentSeed = getDevelopmentSeed();
	if (developmentSeed) return { nsec: developmentSeed, publicKey: publicKeyFromNsec(developmentSeed) };
	const existing = await loadIdentity();
  if (existing) return existing;
  const identity = createIdentity();
  await SecureStore.setItemAsync(identityKey, identity.nsec, secureOptions);
  return identity;
}

function getDevelopmentSeed(): string | null {
	if (!__DEV__) return null;
	const nsec = process.env.EXPO_PUBLIC_DEV_NSEC?.trim();
	return nsec || null;
}
