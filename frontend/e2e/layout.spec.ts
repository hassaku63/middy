import { expect, test, type Locator, type Page } from '@playwright/test';
import { openExample } from './helpers';

/**
 * BL-061 (docs/product/experience.md): while an example plays, the timeline
 * and the explanation of the chord that is sounding are seen together.
 */
async function inViewport(page: Page, locator: Locator) {
  const box = await locator.boundingBox();
  const vp = page.viewportSize();
  expect(box).not.toBeNull();
  expect(vp).not.toBeNull();
  if (!box || !vp) return;
  // Layout lands on fractional pixels (800.0156 for a box that ends at the
  // bottom edge); allow the rounding, not a visible overflow.
  const px = 1;
  expect(box.y).toBeGreaterThanOrEqual(-px);
  expect(box.y + box.height).toBeLessThanOrEqual(vp.height + px);
  expect(box.x + box.width).toBeLessThanOrEqual(vp.width + px);
}

test.describe('at 1280×800', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('the timeline and the sounding chord are visible together while playing', async ({
    page,
  }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    await page.getByRole('button', { name: '再生' }).click();
    const timeline = result.getByRole('group', { name: 'タイムライン' });
    await timeline.scrollIntoViewIfNeeded();
    await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
    await inViewport(page, timeline);
    await inViewport(page, result.getByRole('heading', { level: 2 }));
    await page.getByRole('button', { name: '停止' }).click();
  });
});

test.describe('at phone width', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('nothing scrolls sideways and the timeline keeps its proportions', async ({ page }) => {
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
      390,
    );
    const box = await result.getByRole('group', { name: 'タイムライン' }).boundingBox();
    const [w, h] = await result
      .getByRole('group', { name: 'タイムライン' })
      .evaluate((el) => [Number(el.getAttribute('width')), Number(el.getAttribute('height'))]);
    expect(box).not.toBeNull();
    // Scaled down to fit, the drawing keeps its aspect ratio instead of a tall empty box.
    expect(Math.abs((box?.height ?? 0) / (box?.width ?? 1) - h / w)).toBeLessThan(0.05);
  });
});

test.describe('in the light theme', () => {
  test.use({ colorScheme: 'light' });

  test('an example card keeps its spec items apart and unshaded (BL-069)', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel('イメージ', { exact: true }).fill('雨の日曜の午後');
    await page.getByRole('button', { name: '送る' }).click();
    const spec = page
      .getByRole('article', { name: '窓辺の雨' })
      .locator('.candidate__spec')
      .first();
    await expect(spec).toHaveCSS('display', 'flex');
    await expect(spec).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
  });
});

/**
 * BL-212: a line marks the boundary between the conversation and the open
 * example, across the page when they are stacked and upright when they sit
 * side by side; before an example opens there is nothing to divide.
 */
test.describe('the boundary between the conversation and the example', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  // The width of a border, read in the page (the side goes as an argument).
  const border = (locator: Locator, side: 'top' | 'left') =>
    locator.evaluate(
      (el, s) => parseFloat(getComputedStyle(el).getPropertyValue(`border-${s}-width`)),
      side,
    );

  test('stacked below 1024 px, a line runs across above the example', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/');
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await border(result, 'top')).toBeGreaterThan(0);
  });

  test('side by side, the line stands between the columns, only once an example is open', async ({
    page,
  }) => {
    await page.goto('/');
    const main = page.locator('.app__main');
    expect(await border(main, 'left')).toBe(0);
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await border(main, 'left')).toBeGreaterThan(0);
    expect(await border(result, 'top')).toBe(0);
  });
});

/**
 * #353, #360: the arrangement follows the width the stage has, not the
 * window's. Wide, the chord's frame is a band above the timeline and the
 * notes one column below; narrow, both below the timeline. A line of an
 * explanation stays a sentence (at most 40 characters, WCAG 1.4.8 for
 * CJK) at any width.
 */
async function longestLine(page: Page) {
  return page.evaluate(() =>
    Math.max(
      ...Array.from(document.querySelectorAll('.explanation p'), (el) => {
        const size = parseFloat(getComputedStyle(el).fontSize);
        return el.getBoundingClientRect().width / size;
      }),
    ),
  );
}

for (const { width, height, band } of [
  { width: 960, height: 1080, band: true },
  { width: 1024, height: 800, band: false },
  { width: 1280, height: 800, band: true },
  { width: 1920, height: 1080, band: true },
  { width: 2560, height: 1440, band: true },
]) {
  test.describe(`at ${width}×${height}`, () => {
    test.use({ viewport: { width, height } });

    test(`the chord frame ${band ? 'above' : 'below'} the timeline, lines of 40 characters at most`, async ({
      page,
    }) => {
      await page.goto('/');
      const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
      const timeline = await result.getByRole('group', { name: 'タイムライン' }).boundingBox();
      const frame = await result.getByRole('region', { name: 'コードの解説' }).boundingBox();
      const notes = await result.getByRole('region', { name: '解説', exact: true }).boundingBox();
      expect(timeline && frame && notes).toBeTruthy();
      if (!timeline || !frame || !notes) return;
      const below = (a: { y: number; height: number }, b: { y: number }) =>
        b.y >= a.y + a.height - 1;
      if (band) {
        expect(below(frame, timeline)).toBe(true);
      } else {
        expect(below(timeline, frame)).toBe(true);
      }
      expect(below(band ? timeline : frame, notes)).toBe(true);
      expect(await longestLine(page)).toBeLessThanOrEqual(40.5);
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
        width,
      );
    });
  });
}

/**
 * #419: on a wide stage the chord's frame and the whole example's notes stop
 * at what they hold, instead of stretching with the right half empty; their
 * right edges line up, and the unpin button stays by the chord's heading.
 */
for (const { width, height } of [
  { width: 1920, height: 1080 },
  { width: 2560, height: 1440 },
]) {
  test.describe(`the explanations' width at ${width}×${height}`, () => {
    test.use({ viewport: { width, height } });

    test('end near their text and line up', async ({ page }) => {
      await page.goto('/');
      const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
      await result.getByRole('listbox', { name: 'コード' }).getByRole('option').first().click();
      const frameRegion = result.getByRole('region', { name: 'コードの解説' });
      const frame = await frameRegion.boundingBox();
      const notes = await result.getByRole('region', { name: '解説', exact: true }).boundingBox();
      const unpin = await frameRegion.getByRole('button', { name: /固定を解除/ }).boundingBox();
      const text = await frameRegion.locator('.explanation__chord p').first().boundingBox();
      expect(frame && notes && unpin && text).toBeTruthy();
      if (!frame || !notes || !unpin || !text) return;
      const right = (b: { x: number; width: number }) => b.x + b.width;
      expect(Math.abs(right(frame) - right(notes))).toBeLessThanOrEqual(1);
      expect(right(frame) - right(unpin)).toBeLessThanOrEqual(48);
      // The text's box is at most 40 characters wide; the frame ends within
      // its padding of where those 40 characters would end.
      expect(right(frame) - (text.x + 40 * 16)).toBeLessThanOrEqual(48);
    });
  });
}

/**
 * ADR-0041 (#353): on a wide screen the page uses its width up to QHD and
 * the timeline's height is capped, so that the timeline and the frame of
 * the sounding chord are in view together while it plays.
 */
for (const { width, height } of [
  { width: 1920, height: 1080 },
  { width: 2560, height: 1440 },
]) {
  test.describe(`playing at ${width}×${height}`, () => {
    test.use({ viewport: { width, height } });

    test('the timeline and the chord frame are in view together', async ({ page }) => {
      await page.goto('/');
      const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
      await page.getByRole('button', { name: '再生' }).click();
      await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
      const timeline = result.getByRole('group', { name: 'タイムライン' });
      await inViewport(page, timeline);
      await inViewport(page, result.getByRole('region', { name: 'コードの解説' }));
      const box = await timeline.boundingBox();
      expect(box?.height ?? Infinity).toBeLessThanOrEqual(height * 0.45 + 1);
      // The page grows past the old 1440 px.
      const stage = await result.boundingBox();
      expect(stage?.width ?? 0).toBeGreaterThan(1440 - 340);
      await page.getByRole('button', { name: '停止' }).click();
    });
  });
}
