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
  { width: 375, height: 667 },
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

/**
 * #511 slice 2: the example's card and the rhythm are folds, closed at
 * first with their gist on the line, opened one at a time, and kept as
 * they were set when another example opens.
 */
test('the card and the rhythm are folded, each on its own', async ({ page }) => {
  await page.goto('/');
  const { result } = await openExample(page, '長い雨の日', '窓辺の雨（長め）');
  const explanation = result.getByRole('region', { name: '解説', exact: true });
  const card = explanation.getByText('この例（AI の提案）', { exact: true });
  const rhythm = explanation.getByRole('heading', { name: 'リズム' });
  await expect(card).toBeVisible();
  await expect(rhythm).toBeVisible();
  const gists = explanation.locator('.fold__gist');
  await expect(gists.filter({ hasText: 'A minor · 72 BPM · 16 小節' })).toBeVisible();
  await expect(gists.filter({ hasText: '伴奏: アルペジオ · ドラム: バラード' })).toBeVisible();
  await expect(explanation.getByText('イメージ: 長い雨の日')).toBeHidden();
  await expect(explanation.getByText('全体', { exact: true })).toBeVisible();

  // Opening the rhythm leaves the card closed.
  await rhythm.click();
  await expect(explanation.getByRole('region', { name: /^伴奏: / })).toBeVisible();
  await expect(explanation.getByText('イメージ: 長い雨の日')).toBeHidden();

  // Another example: the rhythm stays open, the card closed.
  await page.getByRole('article', { name: '遠い記憶' }).getByRole('button').click();
  await expect(result.getByRole('heading', { name: '遠い記憶' })).toBeVisible();
  await expect(explanation.getByRole('region', { name: /^伴奏: / })).toBeVisible();
  await expect(explanation.getByText(/^イメージ: /)).toBeHidden();

  // The keyboard opens a fold too.
  await explanation.locator('summary').filter({ hasText: 'この例（AI の提案）' }).focus();
  await page.keyboard.press('Enter');
  await expect(explanation.getByText(/^イメージ: /)).toBeVisible();
});
