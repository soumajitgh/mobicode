import { z } from 'zod';

import { AppError } from '@/api/errors';

const payloadSchema = z.object({
  serverBaseURL: z.url().refine((value) => /^https?:\/\//.test(value)),
  token: z.string().min(1),
});

export function parsePairingPayload(value: string) {
  try {
    const url = new URL(value.trim());
    if (
      url.protocol !== 'mobicode:' ||
      url.hostname !== 'pair' ||
      url.pathname !== '' ||
      url.hash !== ''
    ) {
      throw new Error('Unexpected pairing URL');
    }
    const serverBaseURL = url.searchParams.get('server');
    const token = url.searchParams.get('token');
    if (
      !serverBaseURL ||
      !token ||
      url.searchParams.getAll('server').length !== 1 ||
      url.searchParams.getAll('token').length !== 1 ||
      [...url.searchParams.keys()].some(
        (key) => key !== 'server' && key !== 'token',
      )
    ) {
      throw new Error('Unexpected pairing fields');
    }
    return payloadSchema.parse({
      serverBaseURL: serverBaseURL.replace(/\/+$/, ''),
      token,
    });
  } catch {
    throw new AppError(
      'invalid_pairing_payload',
      'Invalid Mobicode pairing code.',
    );
  }
}
