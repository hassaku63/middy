import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * BL-184 (#279): the desk under trial puts the conversation in a left
 * column beside the stage on wide screens, so that the input, the
 * timeline and the explanation are in view together. Today's layout stays
 * the default; ?lab switches, and the choice is kept in the URL.
 */
test.use({ viewport: { width: 1280, height: 800 } });

test('the desk shows the input, the timeline and the explanation at once (?layout=desk)', async ({
  page,
}) => {
  await page.goto('/?layout=desk');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  await expect(page.getByLabel('イメージ', { exact: true })).toBeInViewport({ ratio: 1 });
  await expect(result.getByRole('group', { name: 'タイムライン' })).toBeInViewport();
  await expect(result.getByRole('heading', { level: 2 })).toBeInViewport({ ratio: 1 });
  // The column sits beside the stage, not above it.
  const side = await page.getByRole('region', { name: 'イメージから' }).boundingBox();
  const stage = await result.boundingBox();
  expect(side && stage && side.x + side.width <= stage.x).toBe(true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(1280);

  // The stage works as in today's layout: play, pin, unpin.
  await page.getByRole('button', { name: '再生' }).click();
  await expect(page.getByLabel('再生位置')).toHaveText(/^2\./, { timeout: 7_000 });
  await page.getByRole('button', { name: '停止' }).click();
  await result.getByRole('option', { name: 'VII G' }).click();
  await expect(result.getByRole('heading', { level: 2 })).toContainText('VII');
  await page.getByRole('button', { name: /固定を解除/ }).click();

  // A second answer arrives in the column; the stage stays beside it.
  await openExample(page, 'もう少し明るく', '朝の光');
  await expect(page.getByLabel('イメージ', { exact: true })).toBeInViewport({ ratio: 1 });
  await expect(result).toContainText('候補: 朝の光');
});

test("today's layout stays the default; ?lab switches to the desk and keeps it in the URL", async ({
  page,
}) => {
  await page.goto('/?lab');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  const stacked = await result.boundingBox();
  const chat = await page.getByRole('region', { name: 'イメージから' }).boundingBox();
  expect(chat && stacked && chat.y + chat.height <= stacked.y).toBe(true);

  await page.getByLabel('画面の配置（試験）').selectOption('desk');
  await expect(page).toHaveURL(/layout=desk/);
  const side = await page.getByRole('region', { name: 'イメージから' }).boundingBox();
  const stage = await result.boundingBox();
  expect(side && stage && side.x + side.width <= stage.x).toBe(true);
});

test('below 1024 px the desk stacks as today', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/?layout=desk');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  const chat = await page.getByRole('region', { name: 'イメージから' }).boundingBox();
  const stage = await result.boundingBox();
  expect(chat && stage && chat.y + chat.height <= stage.y).toBe(true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
});
