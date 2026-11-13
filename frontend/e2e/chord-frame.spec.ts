import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * #511 (ADR-0074): while it plays, the sounding chord's header and the
 * start of its text are on screen with the timeline, the result's top at
 * the top of the screen, at every size. On a narrow stage the text comes
 * before the degree map; on the band (760 px and wider) the map stays on
 * the left of the text.
 */
for (const viewport of [
  { width: 1024, height: 768 },
  { width: 390, height: 844 },
  { width: 1280, height: 800 },
  { width: 820, height: 1180 },
]) {
  test.describe(`at ${String(viewport.width)}×${String(viewport.height)}`, () => {
    test.use({ viewport });

    test('the sounding chord is read without scrolling while it plays', async ({ page }) => {
      await page.goto('/');
      const { result } = await openExample(page, '長い雨の日', '窓辺の雨（長め）');
      await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
      await result.evaluate((el) => {
        el.scrollIntoView({ block: 'start' });
      });
      const frame = result.getByRole('region', { name: 'コードの解説' });
      await expect(frame.getByRole('heading', { level: 2 })).toBeInViewport({ ratio: 1 });
      // The text's first two lines are on screen.
      const room = await frame.locator('.explanation__chord .explanation__text').evaluate((el) => {
        const lineHeight = parseFloat(getComputedStyle(el).lineHeight);
        return window.innerHeight - (el.getBoundingClientRect().top + 2 * lineHeight);
      });
      expect(room).toBeGreaterThanOrEqual(0);
      await expect(result.getByRole('group', { name: 'タイムライン' }).first()).toBeInViewport();

      const [map, text] = await Promise.all([
        frame.locator('.degree-map').boundingBox(),
        frame.locator('.explanation__chord').boundingBox(),
      ]);
      if (!map || !text) throw new Error('no map or text');
      if (viewport.width >= 1280 || viewport.width === 820) {
        expect(map.x).toBeLessThan(text.x); // the band: map, then text
      } else {
        expect(text.y).toBeLessThan(map.y); // stacked: text, then map
      }
    });
  });
}
