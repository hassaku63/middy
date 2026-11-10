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

/**
 * BL-185 (#279): on the desk, one chord's explanation sits in a frame of
 * fixed size above the timeline. Pinning shows by the frame's colour and
 * by ▶／■ on the chord lane; ← → move and pin, Esc follows again.
 */
test('the chord frame keeps its size, and the keys pin and unpin (BL-185)', async ({ page }) => {
  await page.goto('/?layout=desk');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  const frame = result.getByRole('region', { name: 'コードの解説' });
  await expect(frame).toHaveClass(/is-following/);

  // Playing through chords of different length and text, the frame stays put.
  const heights = new Set<number>();
  await page.getByRole('button', { name: '再生' }).click();
  for (let i = 0; i < 8; i++) {
    heights.add(Math.round((await frame.boundingBox())?.height ?? 0));
    await page.waitForTimeout(400);
  }
  await page.getByRole('button', { name: '停止' }).click();
  expect(heights.size).toBe(1);
  // 141 characters of the widest kind fit at 1280 px without scrolling.
  const fits = await frame.evaluate((f) => {
    const text = f.querySelector('.explanation__text');
    if (text) text.textContent = 'あ'.repeat(141);
    return f.scrollHeight <= f.clientHeight;
  });
  expect(fits).toBe(true);

  await page.locator('body').click({ position: { x: 5, y: 5 } });
  await page.keyboard.press('ArrowRight');
  await expect(frame).toHaveClass(/is-pinned/);
  await expect(result.locator('.timeline__mark').filter({ hasText: '■' }).first()).toBeVisible();
  const first = await frame.getByRole('heading', { level: 2 }).textContent();
  await page.keyboard.press('ArrowRight');
  await expect(frame.getByRole('heading', { level: 2 })).not.toHaveText(first ?? '');

  await page.keyboard.press('Escape');
  await expect(frame).toHaveClass(/is-following/);
  await expect(result.locator('.timeline__mark').filter({ hasText: '■' })).toHaveCount(0);

  // Typing in the input keeps the arrows for the input.
  await page.getByLabel('イメージ', { exact: true }).focus();
  await page.keyboard.press('ArrowLeft');
  await expect(frame).toHaveClass(/is-following/);
});

/**
 * BL-186 (#279): on the desk, earlier exchanges become one line per
 * example. Four exchanges fit without scrolling; an earlier example opens
 * in one step and comes back with the chord that was pinned; 1／2／3 open
 * the latest answer's examples.
 */
test('the column lists earlier examples and brings one back as it was left (BL-186)', async ({
  page,
}) => {
  await page.goto('/?layout=desk');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  await result.getByRole('option', { name: 'VII G' }).click();
  await openExample(page, 'もう少し明るく', '朝の光');
  await openExample(page, '夜の帰り道', '窓辺の雨');
  await openExample(page, '晴れた朝', '朝の光');

  const earlier = page.getByRole('list', { name: 'これまでの例' });
  const lines = earlier.getByRole('button');
  await expect(lines).toHaveCount(6);
  for (const line of await lines.all()) await expect(line).toBeInViewport({ ratio: 1 });

  // The same name in two exchanges, told apart by the exchange's number.
  await earlier.getByRole('button', { name: /^1 回目：\s*窓辺の雨/ }).click();
  await expect(result).toContainText('候補: 窓辺の雨');
  const frame = result.getByRole('region', { name: 'コードの解説' });
  await expect(frame).toHaveClass(/is-pinned/);
  await expect(frame.getByRole('heading', { level: 2 })).toContainText('VII');
  await expect(earlier.getByRole('button', { name: /^1 回目：\s*窓辺の雨/ })).toHaveAttribute(
    'aria-pressed',
    'true',
  );

  // The keys open the latest answer's examples, not the earlier ones.
  await page.locator('body').click({ position: { x: 5, y: 5 } });
  await page.keyboard.press('2');
  await expect(result).toContainText('候補: 出発の朝');
  await page.keyboard.press('1');
  await expect(result).toContainText('候補: 朝の光');
  await expect(frame).toHaveClass(/is-following/);
});

/**
 * BL-197 (#314): the wide desk puts the timeline across the whole stage,
 * with the chord's frame and the other explanations side by side below it,
 * so the timeline is wider than on the desk.
 */
test('the wide desk gives the timeline the whole stage, the explanations below (BL-197)', async ({
  page,
}) => {
  const timelineWidth = async (layout: string) => {
    await page.goto(`/?layout=${layout}`);
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    const box = await result.getByRole('group', { name: 'タイムライン' }).boundingBox();
    return { result, box };
  };
  const desk = await timelineWidth('desk');
  const wide = await timelineWidth('wide');
  expect(wide.box && desk.box && wide.box.width > desk.box.width * 1.3).toBe(true);

  const frame = await wide.result.getByRole('region', { name: 'コードの解説' }).boundingBox();
  const notes = await wide.result.getByRole('region', { name: '解説', exact: true }).boundingBox();
  expect(frame && wide.box && frame.y >= wide.box.y + wide.box.height).toBe(true);
  expect(
    frame && notes && frame.x + frame.width <= notes.x && Math.abs(frame.y - notes.y) < 2,
  ).toBe(true);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(1280);
});

/**
 * #315: on the desk, choosing an example plays it and choosing it again
 * stops it; the switch beside the player turns this off and is
 * remembered; Space plays and stops. An arriving answer stays silent.
 */
test('choosing an example on the desk plays it; Space plays and stops (#315)', async ({ page }) => {
  await page.goto('/?layout=desk');
  await page.getByLabel('イメージ', { exact: true }).fill('雨の日曜の午後');
  await page.getByRole('button', { name: '送る' }).click();
  const result = page.getByRole('region', { name: '結果' });
  // The answer opens nothing by itself (BL-219).
  await expect(page.getByRole('article', { name: '遠い記憶' })).toBeVisible();
  await expect(result).toHaveCount(0);

  await page.getByRole('article', { name: '遠い記憶' }).click();
  await expect(page.getByRole('button', { name: '停止' })).toBeEnabled();
  await page.getByRole('article', { name: '遠い記憶' }).click();
  await expect(page.getByRole('button', { name: '停止' })).toBeDisabled();

  await page.locator('body').click({ position: { x: 5, y: 5 } });
  await page.keyboard.press('Space');
  await expect(page.getByRole('button', { name: '停止' })).toBeEnabled();
  await page.keyboard.press('Space');
  await expect(page.getByRole('button', { name: '停止' })).toBeDisabled();

  await page.getByLabel('候補を選んだら再生').uncheck();
  await page.reload();
  await openExample(page, '雨の日曜の午後', '窓辺の雨');
  await expect(page.getByLabel('候補を選んだら再生')).not.toBeChecked();
  await page.getByRole('article', { name: '遠い記憶' }).click();
  await expect(result).toContainText('候補: 遠い記憶');
  await expect(page.getByRole('button', { name: '停止' })).toBeDisabled();
});

/**
 * BL-218, BL-222: the desk's input comes right after the conversation,
 * high up while the conversation is short, and stays in view at the
 * column's foot once it fills the column.
 */
test('the desk input rises into a short conversation and sticks at the foot of a long one', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  const input = page.getByLabel('イメージ', { exact: true });

  await page.goto('/?layout=desk');
  expect((await input.boundingBox())?.y ?? 9999).toBeLessThan(400);

  await openExample(page, '雨の日曜の午後', '窓辺の雨');
  await openExample(page, 'もっと明るく', '朝の光');
  await openExample(page, 'もう少し静かに', '朝の光');
  await expect(input).toBeInViewport();
});
