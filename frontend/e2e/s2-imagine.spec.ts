import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

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
  // The model said there was nothing it could not express (ADR-0061).
  await expect(log.getByText(/表せなかったもの/)).toHaveCount(0);

  // Nothing is chosen until the user chooses (BL-219); pressing the card opens it.
  const result = page.getByRole('region', { name: '結果' });
  await expect(result).toHaveCount(0);
  await card.getByRole('button', { name: '窓辺の雨' }).click();
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /i.*Am/,
    /VI.*F/,
    /III.*C/,
    /VII.*G/,
  ]);
  await expect(result.getByRole('heading', { name: '窓辺の雨' })).toBeVisible();
  // The fake's sample carries patterns (arpeggio + ballad drums): a drum lane and rhythm notes appear.
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  await expect(result.getByRole('heading', { name: 'リズム' })).toBeVisible();
  await expect(card).toContainText('伴奏: アルペジオ');
  await expect(card.getByRole('button', { name: '窓辺の雨' })).toHaveAttribute(
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
  // A refined answer leaves the chosen example open; the user opens a new one.
  await expect(result.getByRole('heading', { name: '窓辺の雨' })).toBeVisible();
  await page.getByRole('article', { name: '朝の光' }).getByRole('button').click();
  await expect(result.getByRole('heading', { name: '朝の光' })).toBeVisible();
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /I.*C/,
    /V7\/vi.*E7/,
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
  await card.getByRole('button').click();
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
  // What the engine cannot express is noted under the summary (ADR-0061).
  await expect(page.getByRole('list', { name: '会話' })).toContainText(
    'この例では表せなかったもの（AI の見立て）: フィドルの装飾音',
  );
  await card.getByRole('button').click();
  const result = page.getByRole('region', { name: '結果' });
  // The dorian IV is a major chord, spelled for the key, and explained by its sixth.
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'IV G' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText(
    'ドリアンの特徴音（主音から長 6 度の B）',
  );
  // The whole example is explained above the chord: the mode and the chord that makes it (BL-098).
  const overview = result.getByLabel('全体と流れ');
  await expect(overview).toContainText('D dorian（ドリアン）');
  await expect(overview).toContainText('短調なら、ここは iv（Gm）になる');
  await expect(overview).toContainText('繰り返すと、最後の VII から頭の i へ、導音を使わずに戻る');
  await page.getByRole('button', { name: /固定を解除/ }).click();
  // Switching to the mixolydian example keeps the key selector in that mode.
  await page.getByRole('article', { name: '港町のジグ' }).click();
  await expect(result.getByLabel('キー')).toHaveValue('7');
  await expect(
    result.getByLabel('キー').locator('option', { hasText: 'G mixolydian' }),
  ).toHaveCount(1);
});

test('a borrowed iv in a major example is spelled for the key and explained (BL-094)', async ({
  page,
}) => {
  await page.goto('/');
  await openExample(page, '出発の朝、晴れ', '出発の朝');
  const card = page.getByRole('article', { name: '出発の朝' });
  await expect(card).toContainText('I - IV - iv - I');
  const result = page.getByRole('region', { name: '結果' });
  // G major's borrowed iv is Cm, taken from G minor.
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'iv Cm' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText(
    '同主短調（G minor）から借りた和音',
  );
});

test('a secondary dominant names the chord it pulls towards (BL-074)', async ({ page }) => {
  await page.goto('/');
  const { result } = await openExample(page, '晴れた朝', '朝の光');
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'V7/vi E7' })
    .click();
  const text = result.locator('.explanation__text').first();
  await expect(text).toContainText('vi（Am）へ向かうセカンダリードミナント');
  await expect(text).toContainText('次の vi へ解決する');
});

test('a waltz image gets a 3/4 example that stays in 3/4 when the key changes (BL-101)', async ({
  page,
}) => {
  await page.goto('/');
  const { card, result } = await openExample(page, 'ウィーンの舞踏会のワルツ', 'ウィーンの円舞曲');
  await expect(card).toContainText('拍子: 3/4');
  await expect(card).toContainText('伴奏: ワルツ');
  const overview = result.getByLabel('全体と流れ');
  await expect(overview).toContainText('3 拍子（3/4）で、1 拍目が強く 2・3 拍目が弱い');
  await expect(overview).toContainText('4 個の和音で 4 小節');
  // The waltz accompaniment names three beats.
  await expect(result.getByLabel('リズム')).toContainText('拍の頭（1・2・3）');
  // F major to D major: the numerals and the meter stay.
  await result.getByLabel('キー').selectOption('2');
  await expect(result.getByLabel('キー')).toHaveValue('2');
  await expect(overview).toContainText('3 拍子（3/4）');
  await expect(overview).toContainText('D major');
});

test('a rocking lullaby gets a 6/8 example counted in two beats, kept when the key changes (BL-102)', async ({
  page,
}) => {
  await page.goto('/');
  const { card, result } = await openExample(page, 'ゆりかごの子守唄', 'ゆりかごの 6/8');
  await expect(card).toContainText('拍子: 6/8');
  await expect(card).toContainText('伴奏: 6/8 の揺れ');
  await expect(card).toContainText('56 BPM');
  const overview = result.getByLabel('全体と流れ');
  await expect(overview).toContainText('6/8 拍子で、付点 4 分の 2 拍に数え');
  await expect(overview).toContainText('4 個の和音で 4 小節');
  // The rocking accompaniment names two beats, each divided in three.
  await expect(result.getByLabel('リズム')).toContainText('拍の頭（1・2）');
  // G major to D major: the meter and the chord lengths stay.
  await result.getByLabel('キー').selectOption('2');
  await expect(result.getByLabel('キー')).toHaveValue('2');
  await expect(overview).toContainText('6/8 拍子');
  await expect(overview).toContainText('4 個の和音で 4 小節');
  await expect(overview).toContainText('D major');
});

test('a Celtic reel holds its tonic chord for two bars, kept when the key changes (BL-109)', async ({
  page,
}) => {
  await page.goto('/');
  const { card, result } = await openExample(
    page,
    'アイルランドのパブのケルト音楽',
    '緑の丘のリール',
  );
  await expect(card).toContainText('長さ: 8・4・2・2 拍');
  await expect(card).toContainText('スタイル: ケルト'); // BL-119
  const overview = result.getByLabel('全体と流れ');
  await expect(overview).toContainText('4 個の和音で 4 小節');
  // D dorian to E dorian: the numerals and each chord's length stay.
  await result.getByLabel('キー').selectOption('4');
  await expect(result.getByLabel('キー')).toHaveValue('4');
  await expect(overview).toContainText('E dorian');
  await expect(overview).toContainText('4 個の和音で 4 小節');
});

test('a graduation song walks its bass down with slash chords, kept when the key changes (BL-116)', async ({
  page,
}) => {
  await page.goto('/');
  const { card, result } = await openExample(page, '卒業式の合唱曲', '一歩ずつ下がるベース');
  await expect(card).toContainText('I - I/7 - vi - vi/5 - IV - I/3 - ii7 - V7');
  const overview = result.getByLabel('全体と流れ');
  await expect(overview).toContainText('ベースが C → B → A → G → F → E → D と順に下がる');
  // The slash chord names its bass (and the ballad's add9, BL-307), and the
  // explanation says what the bass does.
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'I/7 C(add9)/B' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText('経過音');
  await page.getByRole('button', { name: /固定を解除/ }).click();
  // C major to D major: the numerals keep their basses, now spelled for D.
  await result.getByLabel('キー').selectOption('2');
  await expect(result.getByLabel('キー')).toHaveValue('2');
  await expect(overview).toContainText('ベースが D → C# → B → A → G → F# → E と順に下がる');
  await expect(
    result.getByRole('listbox', { name: 'コード' }).getByRole('option', { name: 'I/7 D(add9)/C#' }),
  ).toHaveCount(1);
});

test('a driving image pushes a chord early and explains the push (BL-085)', async ({ page }) => {
  await page.goto('/');
  const { result } = await openExample(page, '勢いのあるロック', '朝の光');
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'IV F' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText(
    '8 分音符ひとつ早く（4 拍目の裏で）入る（押し）',
  );
  // A key change keeps the push.
  await result.getByLabel('キー').selectOption('2');
  await result
    .getByRole('listbox', { name: 'コード' })
    .getByRole('option', { name: 'IV G' })
    .click();
  await expect(result.locator('.explanation__text').first()).toContainText('（押し）');
});

test('a half-time image writes its drums as a grid and explains it (BL-077)', async ({ page }) => {
  await page.goto('/');
  const { result } = await openExample(page, 'ハーフタイムのポップ', '朝の光');
  const rhythm = result.getByLabel('リズム');
  await expect(rhythm).toContainText('ハーフタイム');
  await expect(rhythm).toContainText('格子');
  await expect(result.getByTestId('drum-lane')).toBeVisible();
  // A key change sends the grid back.
  await result.getByLabel('キー').selectOption('2');
  await expect(result.getByLabel('キー')).toHaveValue('2');
  await expect(rhythm).toContainText('ハーフタイム');
});
