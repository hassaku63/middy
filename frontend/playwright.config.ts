import { defineConfig, devices } from '@playwright/test';

/**
 * End-to-end scenarios against a running middy on :8080 by default: the
 * production-like compose stack (ADR-0027), or the single binary.
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
