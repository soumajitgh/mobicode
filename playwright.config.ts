import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { defineConfig, devices } from '@playwright/test';

const baseURL = 'http://127.0.0.1:18080';
const testDirectory = mkdtempSync(join(tmpdir(), 'mobicode-e2e-'));

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'go run ./cmd/server',
    url: `${baseURL}/healthz`,
    reuseExistingServer: false,
    timeout: 120_000,
    env: {
      ...process.env,
      MOBICODE_SERVER_ENV: 'development',
      MOBICODE_SERVER_PORT: '18080',
      MOBICODE_SERVER_DATA_DIR: testDirectory,
      MOBICODE_SERVER_DB_LOG_LEVEL: 'silent',
      MOBICODE_SERVER_SECRET_TOKEN: 'e2e-recovery-token-with-at-least-32-bytes',
    },
  },
});
