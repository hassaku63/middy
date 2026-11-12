import { expect, test, type Page } from '@playwright/test';

/**
 * #360: a long timeline is one drawing that scrolls sideways in its own
 * box, with the whole piece as a thin strip below that outlines the part
 * in view. The fake LLM repeats its progression four times on 長い (16
 * chords, 16 bars at 4 beats each). The page never scrolls sideways.
 */
test.use({ viewport: { width: 1280, height: 800 } });

async function send(page: Page, image: string, path = '/') {
  await page.goto(path);
  await page.getByLabel('イメージ', { exact: true }).fill(image);
  await page.getByRole('button', { name: '送る' }).click();
  // An answer opens nothing by itself (BL-219): open its first example.
  await page.getByRole('article').first().getByRole('button').click();
  // Choosing plays it (#315); stopped, the detail stays where it is put.
  const stop = page.getByRole('button', { name: '停止' });
  await expect(stop).toBeVisible();
  if (await stop.isEnabled()) await stop.click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('group', { name: 'タイムライン' }).first()).toBeVisible();
  const width = page.viewportSize()?.width ?? 0;
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
    width,
  );
  return result;
}

test('a long timeline scrolls in its box; pinning a chord in the strip brings it into view', async ({
  page,
}) => {
  const result = await send(page, '長い雨の日');
  await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
  const box = result.locator('.timeline-overview__scroll');
  await expect(box.getByRole('group', { name: 'タイムライン' })).toHaveCount(1);
  const strip = result.locator('svg.timeline--overview');
  expect((await strip.boundingBox())?.height ?? 99).toBeLessThanOrEqual(64);
  // The strip is a duplicate drawing: its chords are no tab stops.
  await expect(strip.locator('[role=option][tabindex]')).toHaveCount(0);
  // Pin the twelfth chord in the strip: the box scrolls to put it in view.
  await strip.locator('[role=option]').nth(11).click();
  const twelfth = box.getByRole('option').nth(11);
  await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
  const [b, c] = await Promise.all([box.boundingBox(), twelfth.boundingBox()]);
  expect(b && c && c.x >= b.x - 1 && c.x + c.width <= b.x + b.width + 1).toBe(true);
});

test('a short timeline is drawn once, whole', async ({ page }) => {
  const result = await send(page, '雨の日曜');
  await expect(result.getByRole('group', { name: 'タイムライン' })).toHaveCount(1);
  await expect(result.locator('svg.timeline--overview')).toHaveCount(0);
});

test('scrolling the box moves the outline in the strip', async ({ page }) => {
  const result = await send(page, '長い雨の日');
  await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
  const highlight = result.locator('svg.timeline--overview .timeline__highlight');
  const x = async () => Number(await highlight.getAttribute('x'));
  const start = await x();
  await result.locator('.timeline-overview__scroll').evaluate((el) => {
    el.scrollLeft = 1000;
  });
  await expect.poll(x).toBeGreaterThan(start);
});

test.describe('at 1920×1080', () => {
  test.use({ viewport: { width: 1920, height: 1080 } });

  test('a long timeline and the chord frame are in one screen', async ({ page }) => {
    const result = await send(page, '長い雨の日');
    await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
    await expect(result.getByRole('group', { name: 'タイムライン' })).toBeInViewport();
    await expect(result.getByRole('region', { name: 'コードの解説' })).toBeInViewport({ ratio: 1 });
  });
});
