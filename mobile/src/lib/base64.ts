const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';

export function base64Encode(value: Uint8Array): string {
  let result = '';
  for (let index = 0; index < value.length; index += 3) {
    const first = value[index];
    const second = value[index + 1];
    const third = value[index + 2];
    result += alphabet[first >> 2];
    result += alphabet[((first & 0x03) << 4) | ((second ?? 0) >> 4)];
    result += second === undefined ? '=' : alphabet[((second & 0x0f) << 2) | ((third ?? 0) >> 6)];
    result += third === undefined ? '=' : alphabet[third & 0x3f];
  }
  return result;
}
