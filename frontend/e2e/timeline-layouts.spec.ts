import { expect, test, type Page } from '@playwright/test';

/**
 * #360: a long timeline is one drawing that scrolls sideways in its own
 * box, with the whole piece as a thin strip below that outlines the part
 * in view. The fake LLM repeats its progression four times on 長い (16
 * chords, 16 bars at 4 beats each). The page never scrolls sideways.
 */
test.use({ viewport: { width: 1280, height: 800 } });

async function send(page: Page, image: string, path = '/') {
  await page.goto(path);
  await page.getByLabel('イメージ', { exact: true }).fill(image);
  await page.getByRole('button', { name: '送る' }).click();
  // An answer opens nothing by itself (BL-219): open its first example.
  await page.getByRole('article').first().getByRole('button').click();
  // Choosing plays it (#315); stopped, the detail stays where it is put.
  const stop = page.getByRole('button', { name: '停止' });
  await expect(stop).toBeVisible();
  if (await stop.isEnabled()) await stop.click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('group', { name: 'タイムライン' }).first()).toBeVisible();
  const width = page.viewportSize()?.width ?? 0;
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
    width,
  );
  return result;
}

test('a long timeline scrolls in its box; a click on the strip moves the playhead there', async ({
  page,
}) => {
  const result = await send(page, '長い雨の日');
  await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
  const box = result.locator('.timeline-overview__scroll');
  await expect(box.getByRole('group', { name: 'タイムライン' })).toHaveCount(1);
  const strip = result.locator('svg.timeline--overview');
  expect((await strip.boundingBox())?.height ?? 99).toBeLessThanOrEqual(64);
  // The strip is a duplicate drawing: its chords are no tab stops.
  await expect(strip.locator('[role=option][tabindex]')).toHaveCount(0);
  // Stopped, a click on the twelfth bar in the strip (BL-330) moves the
  // playhead there, and the box scrolls to put it in view; nothing pins.
  await strip
    .locator('[role=option]')
    .nth(11)
    .click({ position: { x: 1, y: 4 } });
  await expect(page.getByLabel('再生位置')).toHaveText('12.1');
  await expect(box.locator('[role=option][aria-selected=true]')).toHaveCount(0);
  await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
  const [b, head] = await Promise.all([
    box.boundingBox(),
    box.getByTestId('playhead').boundingBox(),
  ]);
  expect(b && head && head.x >= b.x && head.x <= b.x + b.width).toBe(true);
});

test('a short timeline is drawn once, whole', async ({ page }) => {
  const result = await send(page, '雨の日曜');
  await expect(result.getByRole('group', { name: 'タイムライン' })).toHaveCount(1);
  await expect(result.locator('svg.timeline--overview')).toHaveCount(0);
});

test('scrolling the box moves the outline in the strip', async ({ page }) => {
  const result = await send(page, '長い雨の日');
  await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
  const highlight = result.locator('svg.timeline--overview .timeline__highlight');
  const x = async () => Number(await highlight.getAttribute('x'));
  const start = await x();
  await result.locator('.timeline-overview__scroll').evaluate((el) => {
    el.scrollLeft = 1000;
  });
  await expect.poll(x).toBeGreaterThan(start);
});

test.describe('at 1920×1080', () => {
  test.use({ viewport: { width: 1920, height: 1080 } });

  test('a long timeline and the chord frame are in one screen', async ({ page }) => {
    const result = await send(page, '長い雨の日');
    await expect(result.getByRole('heading', { name: '窓辺の雨（長め）' })).toBeVisible();
    await expect(result.getByRole('group', { name: 'タイムライン' })).toBeInViewport();
    await expect(result.getByRole('region', { name: 'コードの解説' })).toBeInViewport({ ratio: 1 });
  });
});

/**
 * BL-326 (BL-127, the owner's choice): playing, the playhead keeps to the
 * middle of the drawing in view and the drawing flows under it; when the
 * person asked the system for less motion, the box turns a page instead.
 */
for (const viewport of [
  { width: 1920, height: 1080 },
  { width: 390, height: 844 },
]) {
  test.describe(`following the playhead at ${String(viewport.width)} px`, () => {
    test.use({ viewport });

    test('playing, the playhead stays in the middle of the drawing in view', async ({ page }) => {
      const result = await send(page, '長い雨の日');
      const box = result.locator('.timeline-overview__scroll');
      await page.getByRole('button', { name: '再生' }).click();
      // Into the second bar at 72 BPM, past where the scroll can start.
      await expect
        .poll(() => box.evaluate((el) => el.scrollLeft), { timeout: 15000 })
        .toBeGreaterThan(0);
      const offset = await box.evaluate((el) => {
        const head = el.querySelector('[data-testid=playhead]')?.getBoundingClientRect();
        const labels = el.querySelector('.timeline-labels')?.getBoundingClientRect();
        const rect = el.getBoundingClientRect();
        if (!head || !labels) return Infinity;
        const middle = labels.right + (rect.right - labels.right) / 2;
        return Math.abs(head.left - middle);
      });
      expect(offset).toBeLessThanOrEqual(8);
    });
  });
}

test.describe('following the playhead with less motion', () => {
  test.use({ viewport: { width: 1920, height: 1080 }, reducedMotion: 'reduce' });

  test('the box turns a page and holds still in between', async ({ page }) => {
    const result = await send(page, '長い雨の日');
    const box = result.locator('.timeline-overview__scroll');
    await page.getByRole('button', { name: '再生' }).click();
    // Past the middle of the first page (centring would scroll here), the
    // drawing still holds still: a page turns only near the right edge.
    await expect
      .poll(
        () =>
          box.evaluate((el) => {
            const head = el.querySelector('[data-testid=playhead]')?.getBoundingClientRect();
            const labels = el.querySelector('.timeline-labels')?.getBoundingClientRect();
            const rect = el.getBoundingClientRect();
            if (!head || !labels) return false;
            return head.left > labels.right + (rect.right - labels.right) / 2 + 20;
          }),
        { timeout: 15000 },
      )
      .toBe(true);
    expect(await box.evaluate((el) => el.scrollLeft)).toBe(0);
  });
});

/**
 * BL-332 (ADR-0068, as a DAW's follow): playing, scrolling the box by hand
 * pauses the following; 「再生位置へ戻る」 takes it up again.
 */
test.describe('scrolling by hand while playing', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('pauses the following until the person asks for it back', async ({ page }) => {
    const result = await send(page, '長い雨の日');
    const box = result.locator('.timeline-overview__scroll');
    await page.getByRole('button', { name: '再生' }).click();
    const back = result.getByRole('button', { name: '再生位置へ戻る' });
    await expect(back).toHaveCount(0);
    // Scroll far ahead of the playhead, as a person reading ahead would.
    await box.hover();
    await page.mouse.wheel(1200, 0);
    await expect(back).toBeVisible();
    const left = await box.evaluate((el) => el.scrollLeft);
    expect(left).toBeGreaterThan(600);
    await page.waitForTimeout(500);
    expect(await box.evaluate((el) => el.scrollLeft)).toBe(left);
    await back.click();
    await expect(back).toHaveCount(0);
    // Back near the start of the piece: the playhead is in view again.
    await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBeLessThan(left);
    const [b, head] = await Promise.all([
      box.boundingBox(),
      box.getByTestId('playhead').boundingBox(),
    ]);
    expect(b && head && head.x >= b.x && head.x <= b.x + b.width).toBe(true);
  });
});
