import { expect, test, type Page } from '@playwright/test';

/**
 * BL-080: a long timeline shows the whole piece small on top and four bars
 * large below. The fake LLM repeats its progression four times on 長い (16
 * chords, 16 bars at 4 beats each). The page never scrolls sideways.
 */
test.use({ viewport: { width: 1280, height: 800 } });

async function send(page: Page, image: string, path = '/') {
  await page.goto(path);
  await page.getByLabel('イメージ', { exact: true }).fill(image);
  await page.getByRole('button', { name: '送る' }).click();
  // An answer opens nothing by itself (BL-219): open its first example.
  await page.getByRole('article').first().getByRole('button').click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('group', { name: 'タイムライン' }).first()).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(1280);
  return result;
}

test('a long timeline shows the whole piece and a four-bar detail that follows a pinned chord (A, ?zoom=page)', async ({
  page,
}) => {
  const result = await send(page, '長い雨の日', '/?zoom=page');
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
  await expect(result.getByRole('group', { name: 'タイムライン' })).toHaveCount(1);
});

test('BL-127 trial: the detail follows by default (B); ?lab switches to pages (A), kept in the URL', async ({
  page,
}) => {
  const result = await send(page, '長い雨の日', '/?lab');
  await expect(result).toContainText('候補: 窓辺の雨（長め）');
  // B, stopped: a slider moves the detail; the overview's outline follows it.
  const slider = result.getByRole('slider', { name: '拡大の位置' });
  await slider.fill(String(8 * 1920));
  await expect(result.locator('svg.timeline--overview .timeline__highlight')).toHaveAttribute(
    'x',
    String(40 + 8 * 160 - 320),
  );

  await page.getByLabel('拡大の動き（試験）').selectOption('page');
  await expect(page).toHaveURL(/zoom=page/);
  await expect(result.getByRole('slider', { name: '拡大の位置' })).toHaveCount(0);
});
