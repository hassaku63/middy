import { expect, test } from '@playwright/test';

/**
 * Scenario S2 (docs/product/vision.md) against the server's default
 * LLM_PROVIDER=fake: the unscripted fake answers with a deterministic,
 * mood-aware sample intent, so this proves the whole chain (chat UI ->
 * /api/v1/ideas -> intent -> engine -> player/view/panel) without a key.
 */
test('describe an image, get candidates, open one, refine', async ({ page }) => {
  await page.goto('/');
  // The image is the only entry (docs/product/experience.md).
  await expect(page.getByRole('button', { name: '進行から' })).toHaveCount(0);

  await page.getByLabel('イメージ', { exact: true }).fill('雨の日曜の午後');
  await page.getByRole('button', { name: '送る' }).click();

  const log = page.getByRole('list', { name: '会話' });
  await expect(log.getByText('雨の日曜の午後')).toBeVisible();
  const card = page.getByRole('article', { name: '窓辺の雨' }); // the fake's minor-mood sample
  await expect(card).toBeVisible();
  await expect(card).toContainText('A minor');
  await expect(card).toContainText('i - VI - III - VII');

  // The first example is open without another click (BL-064): one press of 再生 makes sound.
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /i.*Am/,
    /VI.*F/,
    /III.*C/,
    /VII.*G/,
  ]);
  await expect(result).toContainText('候補: 窓辺の雨');
  // The fake's sample carries patterns (arpeggio + ballad drums): a drum lane and rhythm notes appear.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  await expect(result.getByRole('heading', { name: 'リズム' })).toBeVisible();
  await expect(card).toContainText('伴奏: アルペジオ');
  await expect(card.getByRole('button', { name: '開いています' })).toHaveAttribute(
    'aria-pressed',
    'true',
  );

  // Playback and explanation work on the example (vision.md S1).
  await page.getByRole('button', { name: '再生' }).click();
  await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
  await page.getByRole('button', { name: '停止' }).click();

  // Refining sends the whole history; the fake switches mood on keywords.
  await page.getByLabel('イメージ', { exact: true }).fill('もっと晴れた感じに');
  await page.getByRole('button', { name: '送る' }).click();
  await expect(page.getByRole('article', { name: '朝の光' })).toBeVisible();
  // A refined answer opens its first example again.
  await expect(result).toContainText('候補: 朝の光');
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /I.*C/,
    /V.*G/,
    /vi.*Am/,
    /IV.*F/,
  ]);
});

test('a swing image swings the example and says so (BL-076)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('イメージ', { exact: true }).fill('夜のジャズバーでスウィングする感じ');
  await page.getByRole('button', { name: '送る' }).click();
  const card = page.getByRole('article', { name: '窓辺の雨' });
  await expect(card).toContainText('ノリ: スウィング');
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('region', { name: 'ノリ: スウィング' })).toBeVisible();
});

test('a Celtic image gets a modal example with its characteristic tone explained (BL-091)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('イメージ', { exact: true }).fill('アイルランドのパブのケルト音楽');
  await page.getByRole('button', { name: '送る' }).click();
  const card = page.getByRole('article', { name: '緑の丘のリール' });
  await expect(card).toContainText('D dorian');
  await expect(card).toContainText('i - IV - i - VII');
  const result = page.getByRole('region', { name: '結果' });
  // The dorian IV is a major chord, spelled for the key, and explained by its sixth.
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'IV (G)' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText(
    'ドリアンの特徴音（主音から長 6 度の B）',
  );
  await page.getByRole('button', { name: /固定を解除/ }).click();
  // Switching to the mixolydian example keeps the key selector in that mode.
  await page
    .getByRole('article', { name: '港町のジグ' })
    .getByRole('button', { name: 'この候補を開く' })
    .click();
  await expect(result.getByLabel('キー')).toHaveValue('7');
  await expect(
    result.getByLabel('キー').locator('option', { hasText: 'G mixolydian' }),
  ).toHaveCount(1);
});
