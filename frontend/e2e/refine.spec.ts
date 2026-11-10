import { expect, test } from '@playwright/test';

/**
 * BL-068 (docs/product/experience.md, 言い直す): refining is the loop, so the
 * conversation must not push playback out of sight. Earlier turns fold away;
 * the latest answer, the input and 再生 stay on the first screen.
 */
test.use({ viewport: { width: 1280, height: 800 } });

test('after three refinements, 再生 is still on the first screen', async ({ page }) => {
  await page.goto('/');
  const say = async (text: string, title: string) => {
    await page.getByLabel('イメージ', { exact: true }).fill(text);
    await page.getByRole('button', { name: '送る' }).click();
    await expect(page.getByRole('region', { name: '結果' })).toContainText(`候補: ${title}`);
  };
  await say('雨の日曜の午後', '窓辺の雨');
  await say('もっと晴れた感じに', '朝の光');
  await say('やっぱり夜の雰囲気で', '窓辺の雨');
  await say('明るく軽やかに', '朝の光');

  await page.evaluate(() => window.scrollTo(0, 0));
  const play = await page.getByRole('button', { name: '再生' }).boundingBox();
  expect(play).not.toBeNull();
  expect((play?.y ?? 9999) + (play?.height ?? 0)).toBeLessThanOrEqual(800);
  await expect(page.getByLabel('イメージ', { exact: true })).toBeInViewport();

  // Earlier turns are folded, not lost.
  const earlier = page.getByText(/これまでのやり取り（3）/);
  await expect(earlier).toBeVisible();
  await earlier.click();
  await expect(page.getByText('雨の日曜の午後')).toBeVisible();
  // The fake repeats titles across answers; only the open example is marked (BL-070).
  await expect(page.getByRole('button', { name: /./, pressed: true })).toHaveCount(1);
});
