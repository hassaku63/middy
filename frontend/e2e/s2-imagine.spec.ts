import { expect, test } from '@playwright/test';

/**
 * Scenario S2 (docs/product/vision.md) against the server's default
 * LLM_PROVIDER=fake: the unscripted fake answers with a deterministic,
 * mood-aware sample intent, so this proves the whole chain (chat UI ->
 * /api/v1/ideas -> intent -> engine -> player/view/panel) without a key.
 */
test('describe an image, get candidates, open one, refine', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'イメージから（AI）' })).toHaveAttribute(
    'aria-pressed',
    'true',
  );

  await page.getByLabel('イメージ', { exact: true }).fill('雨の日曜の午後');
  await page.getByRole('button', { name: '送る' }).click();

  const log = page.getByRole('list', { name: '会話' });
  await expect(log.getByText('雨の日曜の午後')).toBeVisible();
  const card = page.getByRole('article', { name: '窓辺の雨' }); // the fake's minor-mood sample
  await expect(card).toBeVisible();
  await expect(card).toContainText('A minor');
  await expect(card).toContainText('i - VI - III - VII');

  await card.getByRole('button', { name: 'この候補を開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('option')).toHaveText([/i.*Am/, /VI.*F/, /III.*C/, /VII.*G/]);
  await expect(result).toContainText('候補: 窓辺の雨');
  // The fake's sample carries patterns (arpeggio + ballad drums): a drum lane and rhythm notes appear.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  await expect(result.getByRole('heading', { name: 'リズム' })).toBeVisible();
  await expect(card).toContainText('伴奏: アルペジオ');
  await expect(card.getByRole('button', { name: '開いています' })).toHaveAttribute(
    'aria-pressed',
    'true',
  );

  // Playback and explanation work on the candidate exactly as in S1.
  await page.getByRole('button', { name: '再生' }).click();
  await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
  await page.getByRole('button', { name: '停止' }).click();

  // Refining sends the whole history; the fake switches mood on keywords.
  await page.getByLabel('イメージ', { exact: true }).fill('もっと晴れた感じに');
  await page.getByRole('button', { name: '送る' }).click();
  await expect(page.getByRole('article', { name: '朝の光' })).toBeVisible();
});
