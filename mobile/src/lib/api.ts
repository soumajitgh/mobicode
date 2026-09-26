import { create } from 'axios';

export const serverURL = process.env.EXPO_PUBLIC_MOBICODE_SERVER_URL?.replace(/\/+$/, '');

export const api = create({
  baseURL: serverURL,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
});
