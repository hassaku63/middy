import { defineConfig, devices } from '@playwright/test';

/**
 * End-to-end scenarios against a running middy server (ADR-0008): by
 * default the production-like single binary / container on :8080.
 *
 *   make build && make run        # or: make up
 *   make e2e                      # E2E_BASE_URL overrides the target
 *
 * Unit tests live in src/ and run with Vitest; this config only sees e2e/.
 */
const baseURL = process.env['E2E_BASE_URL'] ?? 'http://localhost:8080';
const executablePath = process.env['E2E_CHROMIUM_PATH'];

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  retries: process.env['CI'] ? 1 : 0,
  reporter: process.env['CI'] ? [['list'], ['html', { open: 'never' }]] : [['list']],
  use: {
    baseURL,
    trace: 'retain-on-failure',
    ...devices['Desktop Chrome'],
    ...(executablePath ? { launchOptions: { executablePath } } : {}),
  },
});
