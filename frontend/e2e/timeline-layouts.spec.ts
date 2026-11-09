import { expect, test, type Page } from '@playwright/test';

/**
 * BL-080: a long timeline shows the whole piece small on top and four bars
 * large below. The fake LLM repeats its progression four times on 長い (16
 * chords, 16 bars at 4 beats each). The page never scrolls sideways.
 */
test.use({ viewport: { width: 1280, height: 800 } });

async function send(page: Page, image: string) {
  await page.goto('/');
  await page.getByLabel('イメージ', { exact: true }).fill(image);
  await page.getByRole('button', { name: '送る' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('img', { name: 'タイムライン' }).first()).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(1280);
  return result;
}

test('a long timeline shows the whole piece and a four-bar detail that follows a pinned chord', async ({
  page,
}) => {
  const result = await send(page, '長い雨の日');
  await expect(result).toContainText('候補: 窓辺の雨（長め）');
  const overview = result.locator('svg.timeline--overview');
  await expect(overview).toBeVisible();
  await expect(overview.locator('.timeline__highlight')).toHaveAttribute('x', '40');
  // Pin the ninth chord in the overview: the detail moves to bars 9–12.
  await overview.locator('[role=option]').nth(8).click();
  await expect(overview.locator('.timeline__highlight')).toHaveAttribute('x', String(40 + 8 * 160));
  // The overview is a duplicate drawing: its chords are no tab stops.
  await expect(overview.locator('[role=option][tabindex]')).toHaveCount(0);
});

test('a short timeline is drawn once, whole', async ({ page }) => {
  const result = await send(page, '雨の日曜');
  await expect(result.locator('svg.timeline--overview')).toHaveCount(0);
  await expect(result.getByRole('img', { name: 'タイムライン' })).toHaveCount(1);
});
