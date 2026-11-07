import { expect, test, type Locator, type Page } from '@playwright/test';
import { openExample } from './helpers';

/**
 * BL-061 (docs/product/experience.md): while an example plays, the timeline
 * and the explanation of the chord that is sounding are seen together.
 */
async function inViewport(page: Page, locator: Locator) {
  const box = await locator.boundingBox();
  const vp = page.viewportSize();
  expect(box).not.toBeNull();
  expect(vp).not.toBeNull();
  if (!box || !vp) return;
  expect(box.y).toBeGreaterThanOrEqual(0);
  expect(box.y + box.height).toBeLessThanOrEqual(vp.height);
  expect(box.x + box.width).toBeLessThanOrEqual(vp.width);
}

test.describe('at 1280×800', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('the timeline and the sounding chord are visible together while playing', async ({
    page,
  }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    await page.getByRole('button', { name: '再生' }).click();
    const timeline = result.getByRole('img', { name: 'タイムライン' });
    await timeline.scrollIntoViewIfNeeded();
    await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
    await inViewport(page, timeline);
    await inViewport(page, result.getByRole('heading', { level: 2 }));
    await page.getByRole('button', { name: '停止' }).click();
  });
});

test.describe('at phone width', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('nothing scrolls sideways and the timeline keeps its proportions', async ({ page }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
      390,
    );
    const box = await result.getByRole('img', { name: 'タイムライン' }).boundingBox();
    const [w, h] = await result
      .getByRole('img', { name: 'タイムライン' })
      .evaluate((el) => [Number(el.getAttribute('width')), Number(el.getAttribute('height'))]);
    expect(box).not.toBeNull();
    // Scaled down to fit, the drawing keeps its aspect ratio instead of a tall empty box.
    expect(Math.abs((box?.height ?? 0) / (box?.width ?? 1) - (h ?? 0) / (w ?? 1))).toBeLessThan(
      0.05,
    );
  });
});
