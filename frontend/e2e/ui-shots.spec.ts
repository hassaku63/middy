import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { test, type Page } from '@playwright/test';
import { openExample } from './helpers';
import { SIZES } from './sizes';

/**
 * Screens for the ui-review skill (BL-183, #286): each state at five sizes
 * and in both themes, as an image and as the accessibility tree. Not a
 * test: it runs only with UI_SHOTS set (`npm run ui:shots`), never in CI,
 * and writes to UI_SHOTS_DIR (default ui-review-out/, ignored by git).
 */
test.skip(!process.env.UI_SHOTS, 'run with npm run ui:shots');

const out = process.env.UI_SHOTS_DIR ?? 'ui-review-out';
const sizes = SIZES;
const layout = process.env.UI_SHOTS_LAYOUT ?? ''; // a query for a trial, e.g. "lab"

/** The states the result screen is reviewed in. */
const states: { name: string; reach: (page: Page) => Promise<void> }[] = [
  { name: '1-first', reach: () => Promise.resolve() },
  {
    name: '2-opened',
    reach: async (page) => {
      await openExample(page, '雨の日曜の午後', '窓辺の雨');
    },
  },
  {
    name: '3-pinned',
    reach: async (page) => {
      const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
      await result.getByRole('option', { name: 'V7 B7' }).click();
    },
  },
  {
    name: '4-refined',
    reach: async (page) => {
      await openExample(page, '雨の日曜の午後', '窓辺の雨');
      await openExample(page, 'もう少し明るく', '朝の光');
    },
  },
];

for (const scheme of ['light', 'dark'] as const) {
  for (const size of sizes) {
    test.describe(`${scheme} ${size.name}`, () => {
      test.use({ colorScheme: scheme, viewport: { width: size.width, height: size.height } });
      for (const state of states) {
        test(state.name, async ({ page }) => {
          await page.goto(layout ? `/?${layout}` : '/');
          await state.reach(page);
          mkdirSync(out, { recursive: true });
          const base = join(out, `${state.name}-${size.name}-${scheme}`);
          await page.screenshot({ path: `${base}.png`, fullPage: true });
          writeFileSync(`${base}.aria.yaml`, await page.locator('body').ariaSnapshot());
        });
      }
    });
  }
}
