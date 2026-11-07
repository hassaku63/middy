import { expect, test } from '@playwright/test';

/**
 * Scenario S1 (docs/product/vision.md): choose a key and a progression,
 * hear it, watch the chord and notes light up, read why it works, take
 * the MIDI home. Runs against the real server, so it also proves the
 * single-binary serving, the API contract and the PWA shell.
 */
test('generate, play, follow, pin, download', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('status')).toHaveText('バックエンド: ok');
  await page.getByRole('button', { name: '進行から' }).click();

  // A minor, i - iv - V7 - i at 120 BPM: two seconds per bar.
  await page.getByLabel('調').selectOption('9');
  await page.getByLabel('モード').selectOption('minor');
  await page.getByLabel('進行（ローマ数字）').fill('i - iv - V7 - i');
  await page.getByLabel('テンポ (BPM)').fill('120');
  await page.getByLabel('伴奏').selectOption('arpeggio');
  await page.getByLabel('ドラム').selectOption('eight-beat');
  await page.getByRole('button', { name: '生成' }).click();

  const result = page.getByRole('region', { name: '結果' });
  // Patterns are applied server side: the drum lane appears and the roll holds eighth notes.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  await expect(result.locator('.timeline__note')).toHaveCount(4 * 8);
  await expect(result.getByRole('option')).toHaveText([
    /i.*Am/,
    /iv.*Dm/,
    /V7.*E7/, // uppercase V in minor is the harmonic-minor dominant
    /i.*Am/,
  ]);
  await expect(result.getByRole('heading')).toContainText('i');

  // Play: after ~2.3 s the playhead is in bar 2 and the panel follows the chord.
  await page.getByRole('button', { name: '再生' }).click();
  await expect(page.getByLabel('再生位置')).toHaveText(/^2\./, { timeout: 5_000 });
  await expect(result.locator('[role=option][aria-current=true]')).toHaveAttribute(
    'aria-label',
    'iv (Dm)',
  );
  await expect(result.locator('.explanation__text')).toContainText('サブドミナント');
  await page.getByRole('button', { name: '停止' }).click();
  await expect(page.getByLabel('再生位置')).toHaveText('1.1');

  // Pin a chord by clicking it; the panel stops following playback.
  await result.getByRole('option', { name: 'V7 (E7)' }).click();
  await expect(result.getByRole('heading')).toContainText('V7');
  await expect(result.locator('.explanation__text')).toContainText('和声的短音階');
  await expect(result.locator('.explanation__cadence')).toHaveText('半終止');
  await page.getByRole('button', { name: /固定を解除/ }).click();

  // Download the MIDI file with a descriptive name.
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.getByRole('button', { name: 'MIDI を保存' }).click(),
  ]);
  expect(download.suggestedFilename()).toBe('middy-A-minor-i-iv-V7-i.mid');
  const path = await download.path();
  expect(path).not.toBeNull();
});

test('invalid roman numerals are reported from the API, not swallowed', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: '進行から' }).click();
  await page.getByLabel('進行（ローマ数字）').fill('I - X');
  await page.getByRole('button', { name: '生成' }).click();
  await expect(page.getByRole('alert')).toContainText('"X" is not a roman numeral');
});

test('deep links fall back to the app shell and the service worker installs', async ({ page }) => {
  await page.goto('/some/deep/route');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('middy');
  // `ready` resolves while the worker may still be "activating"; poll until it settles.
  await expect
    .poll(
      () =>
        page.evaluate(async () => {
          const reg = await navigator.serviceWorker.ready;
          return reg.active?.state ?? 'none';
        }),
      { timeout: 10_000 },
    )
    .toBe('activated');
});
