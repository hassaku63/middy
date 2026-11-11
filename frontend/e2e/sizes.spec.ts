import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';
import { openExample } from './helpers';
import { SIZES } from './sizes';

/**
 * The layout at every size of the matrix (#359, ADR-0042), checked by
 * machine, not by eye: the page never scrolls sideways, axe finds nothing,
 * and the same operations are there at every width ("幅で、できることを
 * 失わない", docs/dev/conventions.md). Resizing keeps the state: a pinned
 * chord stays pinned. These run in the CI of every pull request that
 * changes the frontend.
 */
async function violations(page: Page) {
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
    .analyze();
  return results.violations.map((v) => `${v.id}: ${v.help} (${String(v.nodes.length)})`);
}

/** The names of the operations on the result, in order of the page. */
async function operations(page: Page) {
  const result = page.getByRole('region', { name: '結果' });
  const buttons = await result.getByRole('button').allInnerTexts();
  const fields = await result.getByRole('combobox').count();
  return { buttons: buttons.map((b) => b.replace(/[▶■]/g, '').trim()), fields };
}

for (const size of SIZES) {
  test.describe(`at ${size.name} (${String(size.width)}×${String(size.height)})`, () => {
    test.use({ viewport: { width: size.width, height: size.height } });

    test('nothing scrolls sideways, axe finds nothing, and the operations are those of 1280', async ({
      page,
    }) => {
      await page.goto('/');
      const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
      const stop = page.getByRole('button', { name: '停止' });
      if (await stop.isVisible()) await stop.click();
      await result.getByRole('option', { name: 'V7 B7' }).click();
      await expect(page.getByRole('button', { name: /固定を解除/ })).toBeVisible();

      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
        size.width,
      );
      expect(await violations(page)).toEqual([]);
      const here = await operations(page);

      // The same page at 1280: the same operations, and the pinned chord
      // still pinned (the layout changes, the state stays).
      await page.setViewportSize({ width: 1280, height: 800 });
      await expect(page.getByRole('button', { name: /固定を解除/ })).toBeVisible();
      expect(await operations(page)).toEqual(here);
    });
  });
}
