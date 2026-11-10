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
  // Layout lands on fractional pixels (800.0156 for a box that ends at the
  // bottom edge); allow the rounding, not a visible overflow.
  const px = 1;
  expect(box.y).toBeGreaterThanOrEqual(-px);
  expect(box.y + box.height).toBeLessThanOrEqual(vp.height + px);
  expect(box.x + box.width).toBeLessThanOrEqual(vp.width + px);
}

test.describe('at 1280×800', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('the timeline and the sounding chord are visible together while playing', async ({
    page,
  }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    await page.getByRole('button', { name: '再生' }).click();
    const timeline = result.getByRole('group', { name: 'タイムライン' });
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
    const box = await result.getByRole('group', { name: 'タイムライン' }).boundingBox();
    const [w, h] = await result
      .getByRole('group', { name: 'タイムライン' })
      .evaluate((el) => [Number(el.getAttribute('width')), Number(el.getAttribute('height'))]);
    expect(box).not.toBeNull();
    // Scaled down to fit, the drawing keeps its aspect ratio instead of a tall empty box.
    expect(Math.abs((box?.height ?? 0) / (box?.width ?? 1) - h / w)).toBeLessThan(0.05);
  });
});

test.describe('in the light theme', () => {
  test.use({ colorScheme: 'light' });

  test('an example card keeps its spec items apart and unshaded (BL-069)', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel('イメージ', { exact: true }).fill('雨の日曜の午後');
    await page.getByRole('button', { name: '送る' }).click();
    const spec = page
      .getByRole('article', { name: '窓辺の雨' })
      .locator('.candidate__spec')
      .first();
    await expect(spec).toHaveCSS('display', 'flex');
    await expect(spec).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
  });
});

/**
 * BL-212: a line marks the boundary between the conversation and the open
 * example, across the page when they are stacked and upright when they sit
 * side by side; before an example opens there is nothing to divide.
 */
test.describe('the boundary between the conversation and the example', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  // The width of a border, read in the page (the side goes as an argument).
  const border = (locator: Locator, side: 'top' | 'left') =>
    locator.evaluate(
      (el, s) => parseFloat(getComputedStyle(el).getPropertyValue(`border-${s}-width`)),
      side,
    );

  test('stacked, a line runs across above the example', async ({ page }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await border(result, 'top')).toBeGreaterThan(0);
  });

  test('side by side, the line stands between the columns, only once an example is open', async ({
    page,
  }) => {
    await page.goto('/?layout=desk');
    const main = page.locator('.app__main');
    expect(await border(main, 'left')).toBe(0);
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await border(main, 'left')).toBeGreaterThan(0);
    expect(await border(result, 'top')).toBe(0);
  });
});
