import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * The foundation under every example (vision.md S1): hear it, watch the
 * chord and notes light up, read why it works, take the MIDI home. Runs
 * against the real server with the fake LLM, so it also proves the
 * single-binary serving, the API contract and the PWA shell.
 */
test('play an example, follow it, pin a chord, download', async ({ page }) => {
  await page.goto('/');

  // 窓辺の雨: A minor, i - VI - III - VII, 72 BPM, arpeggio and ballad drums.
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  // Patterns are applied server side: the drum lane appears and the roll holds eighth notes.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  await expect(result.locator('.timeline__note')).toHaveCount(4 * 8);
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /i.*Am/,
    /VI.*F/,
    /III.*C/,
    /VII.*G/,
  ]);
  await expect(result.getByRole('heading', { level: 2 })).toContainText('i');

  // Play: at 72 BPM a bar lasts 3.3 s; the panel follows into bar 2.
  await page.getByRole('button', { name: '再生' }).click();
  await expect(page.getByLabel('再生位置')).toHaveText(/^2\./, { timeout: 7_000 });
  await expect(result.locator('[role=option][aria-current=true]')).toHaveAttribute(
    'aria-label',
    'VI F',
  );
  await expect(result.getByRole('heading', { level: 2 })).toContainText('VI');
  await page.getByRole('button', { name: '停止' }).click();
  await expect(page.getByLabel('再生位置')).toHaveText('1.1');

  // 遠い記憶: E minor, ending iiø7 - V7 - i. Pin the dominant by clicking it.
  // The whole card opens it (#313): press the card, not a button in it.
  await page.getByRole('article', { name: '遠い記憶' }).click();
  await result.getByRole('option', { name: 'V7 B7' }).click();
  await expect(result.getByRole('heading', { level: 2 })).toContainText('V7');
  // Uppercase V in minor is the harmonic-minor dominant.
  await expect(result.locator('.explanation__text').first()).toContainText('和声的短音階');
  await page.getByRole('button', { name: /固定を解除/ }).click();

  // Download the MIDI file with a descriptive name.
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.getByRole('button', { name: 'MIDI を保存' }).click(),
  ]);
  expect(download.suggestedFilename()).toMatch(/^middy-E-minor-i-iv-VII-III-VI-.*-V7-i\.mid$/);
  expect(await download.path()).not.toBeNull();
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

test('change the key of an example: the numerals stay, the chords move (BL-078)', async ({
  page,
}) => {
  await page.goto('/');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨'); // A minor, arpeggio + ballad
  await result.getByLabel('キー').selectOption('0'); // C minor
  const chords = result.getByRole('listbox', { name: 'コード' }).getByRole('option');
  // Spelled for C minor (three flats), not with sharps.
  await expect(chords).toHaveText([/i.*Cm/, /VI.*Ab/, /III.*Eb/, /VII.*Bb/]);
  // The patterns come along: the drum lane is still there.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.getByRole('button', { name: 'MIDI を保存' }).click(),
  ]);
  expect(download.suggestedFilename()).toBe('middy-C-minor-i-VI-III-VII.mid');
});

test('play the same example in another style: the harmony stays, the groove changes (BL-087)', async ({
  page,
}) => {
  await page.goto('/');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨'); // A minor, arpeggio + ballad
  const chords = result.getByRole('listbox', { name: 'コード' }).getByRole('option');
  // Named by numeral and chord (the ▶ of the followed chord is no part of it).
  const names = () => chords.evaluateAll((os) => os.map((o) => o.getAttribute('aria-label')));
  const before = await names();
  await result.getByLabel('スタイル').selectOption('jazz');
  await expect(result.getByLabel('スタイル')).toHaveValue('jazz');
  const note = result.getByRole('status').filter({ hasText: '鳴らし直しました' });
  await expect(note).toContainText('ジャズで鳴らし直しました。和声はそのままです。');
  await expect(note).toContainText('伴奏: アルペジオ → 4 ビートのコンピング');
  await expect(note).toContainText('ノリ: ストレート → スウィング');
  await expect.poll(names).toEqual(before);
  await expect(result.getByLabel('リズム')).toContainText('スウィング');
});
