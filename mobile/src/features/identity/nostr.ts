import * as ExpoCrypto from 'expo-crypto';
import { finalizeEvent, generateSecretKey, getPublicKey } from 'nostr-tools/pure';
import { decode, nsecEncode } from 'nostr-tools/nip19';

import { base64Encode } from '../../lib/base64';

const encoder = new TextEncoder();

function installRandomValues(): void {
  if (!globalThis.crypto?.getRandomValues) {
    Object.defineProperty(globalThis, 'crypto', {
      configurable: true,
      value: { getRandomValues: ExpoCrypto.getRandomValues },
    });
  }
}

export type MobileIdentity = {
  nsec: string;
  publicKey: string;
};

export function createIdentity(): MobileIdentity {
  installRandomValues();
  const secretKey = generateSecretKey();
  return { nsec: nsecEncode(secretKey), publicKey: getPublicKey(secretKey) };
}

export function publicKeyFromNsec(nsec: string): string {
  installRandomValues();
  const decoded = decode(nsec);
  if (decoded.type !== 'nsec') {
    throw new Error('Stored signing key is not an nsec.');
  }
  return getPublicKey(decoded.data);
}

export async function createNip98Authorization(url: string, method: string, body: string, nsec: string): Promise<string> {
  installRandomValues();
  const decoded = decode(nsec);
  if (decoded.type !== 'nsec') {
    throw new Error('Stored signing key is not an nsec.');
  }
  const payload = await ExpoCrypto.digestStringAsync(ExpoCrypto.CryptoDigestAlgorithm.SHA256, body, {
    encoding: ExpoCrypto.CryptoEncoding.HEX,
  });
  const event = finalizeEvent(
    {
      kind: 27235,
      created_at: Math.floor(Date.now() / 1000),
      tags: [
        ['u', url],
        ['method', method],
        ['payload', payload],
      ],
      content: '',
    },
    decoded.data,
  );
  return `Nostr ${base64Encode(encoder.encode(JSON.stringify(event)))}`;
}
