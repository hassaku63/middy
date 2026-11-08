import { expect, test } from '@playwright/test';

/**
 * BL-080 trial: long timelines in four layouts, chosen with ?timeline=<id>.
 * The fake LLM repeats its progression four times on 長い (16 chords, 16
 * bars at 4 beats each). Each layout must keep the page from scrolling
 * sideways and keep chords clickable.
 */
test.use({ viewport: { width: 1280, height: 800 } });

async function openLong(page: import('@playwright/test').Page, layout: string) {
  await page.goto(`/?timeline=${layout}`);
  await page.getByLabel('イメージ', { exact: true }).fill('長い雨の日');
  await page.getByRole('button', { name: '送る' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result).toContainText('候補: 窓辺の雨（長め）');
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(1280);
  return result;
}

test('fit draws one timeline (today)', async ({ page }) => {
  const result = await openLong(page, 'fit');
  await expect(result.getByRole('img', { name: 'タイムライン' })).toHaveCount(1);
});

test('scroll keeps 160 px a bar and scrolls inside its box', async ({ page }) => {
  const result = await openLong(page, 'scroll');
  const box = result.getByTestId('timeline-scroll');
  const [scrollWidth, clientWidth] = await box.evaluate((el) => [el.scrollWidth, el.clientWidth]);
  expect(scrollWidth).toBeGreaterThan(clientWidth);
  expect(scrollWidth).toBeGreaterThanOrEqual(16 * 160);
});

test('rows wraps four bars a row, and a chord in row 2 can be pinned', async ({ page }) => {
  const result = await openLong(page, 'rows');
  await expect(result.getByRole('img', { name: /タイムライン（\d+〜\d+ 小節）/ })).toHaveCount(4);
  const row2 = result.getByRole('img', { name: 'タイムライン（5〜8 小節）' });
  // Each row draws the whole piece and shows a slice; chords 5–8 are the visible ones.
  await row2.getByRole('option').nth(5).click();
  await expect(result.getByRole('heading', { level: 2 })).toContainText('VI');
});

test('overview shows the whole piece and a four-bar detail that follows a pinned chord', async ({
  page,
}) => {
  const result = await openLong(page, 'overview');
  const overview = result.locator('svg.timeline--overview');
  await expect(overview).toBeVisible();
  await expect(overview.locator('.timeline__highlight')).toHaveAttribute('x', '40');
  // Pin the ninth chord in the overview: the detail moves to bars 9–12.
  await overview.locator('[role=option]').nth(8).click();
  await expect(overview.locator('.timeline__highlight')).toHaveAttribute('x', String(40 + 8 * 160));
});

test('?lab shows a switcher that rewrites the URL', async ({ page }) => {
  await page.goto('/?lab');
  await page.getByLabel('イメージ', { exact: true }).fill('長い雨の日');
  await page.getByRole('button', { name: '送る' }).click();
  await page.getByLabel('表示（試験）').selectOption('rows');
  await expect(page).toHaveURL(/timeline=rows/);
  await expect(page.getByRole('img', { name: /タイムライン（\d+〜\d+ 小節）/ })).toHaveCount(4);
});
