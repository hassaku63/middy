import { expect, test, type Page } from '@playwright/test';

/**
 * BL-328 (the owner's rules): clicking the timeline anywhere but on a
 * chord moves the playhead to the start of the beat under it. Stopped, the
 * next play starts there; playing, play jumps there. A chord still pins.
 * BL-334: ⏸ stays where it is, ■ goes back to the beginning. The fake LLM's 長い例 is 16 bars of 4/4.
 */
async function open(page: Page) {
  await page.goto('/');
  await page.getByLabel('イメージ', { exact: true }).fill('長い雨の日');
  await page.getByRole('button', { name: '送る' }).click();
  await page.getByRole('article').first().getByRole('button').click();
  // Choosing plays it (#315): stop, so each test starts at the beginning.
  await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
  await page.getByRole('button', { name: '停止', exact: true }).click();
  const result = page.getByRole('region', { name: '結果' });
  const drawing = result.locator('.timeline-overview__scroll svg.timeline');
  await expect(drawing).toBeVisible();
  return { result, drawing, position: page.getByLabel('再生位置') };
}

/** Clicks a bar's number in the ruler: the first beat of that bar. */
async function clickBar(page: Page, bar: number) {
  await page
    .locator('.timeline-overview__scroll .timeline__bar-number', {
      hasText: new RegExp(`^${String(bar)}$`),
    })
    .click();
}

for (const viewport of [
  { width: 1280, height: 800 },
  { width: 390, height: 844 },
]) {
  test.describe(`at ${String(viewport.width)} px`, () => {
    test.use({ viewport });

    test('stopped, a click sets where play starts; pause stays, stop goes to the beginning', async ({
      page,
    }) => {
      const { position } = await open(page);
      await clickBar(page, 3);
      await expect(position).toHaveText('3.1');
      await page.getByRole('button', { name: '再生' }).click();
      await expect(position).toHaveText(/^3\.[2-4]$/, { timeout: 5000 });
      await page.getByRole('button', { name: '一時停止' }).click();
      const paused = await position.textContent();
      expect(paused).toMatch(/^3\.[2-4]$/);
      await page.getByRole('button', { name: '再生' }).click();
      await expect(position).not.toHaveText(paused ?? '');
      await expect(position).toHaveText(/^[34]\./);
      await page.getByRole('button', { name: '停止', exact: true }).click();
      await expect(position).toHaveText('1.1');
      await expect(page.getByRole('button', { name: '停止', exact: true })).toBeDisabled();
    });

    test('playing, a click moves play there', async ({ page }) => {
      const { position } = await open(page);
      await page.getByRole('button', { name: '再生' }).click();
      await expect(position).toHaveText(/^1\.[2-4]$/, { timeout: 5000 });
      await clickBar(page, 2);
      await expect(position).toHaveText(/^2\./);
      await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
    });

    test('a chord still pins and leaves the playhead alone', async ({ page }) => {
      const { result, position } = await open(page);
      const box = result.locator('.timeline-overview__scroll');
      const second = box.getByRole('option').nth(1);
      await second.click();
      await expect(second).toHaveAttribute('aria-selected', 'true');
      await expect(position).toHaveText('1.1');
    });
  });
}

test.describe('on a touch screen', () => {
  test.use({ viewport: { width: 390, height: 844 }, hasTouch: true, isMobile: true });

  test('panning the box sideways does not move the playhead', async ({ page }) => {
    const { result, position } = await open(page);
    const box = result.locator('.timeline-overview__scroll');
    const b = await box.boundingBox();
    if (!b) throw new Error('no box');
    const y = b.y + b.height * 0.7;
    const cdp = await page.context().newCDPSession(page);
    const touch = (type: 'touchStart' | 'touchMove' | 'touchEnd', x: number) =>
      cdp.send('Input.dispatchTouchEvent', {
        type,
        touchPoints: type === 'touchEnd' ? [] : [{ x, y }],
      });
    await touch('touchStart', b.x + b.width * 0.8);
    for (let i = 1; i <= 10; i += 1) await touch('touchMove', b.x + b.width * (0.8 - i * 0.06));
    await touch('touchEnd', 0);
    await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
    await expect(position).toHaveText('1.1');
  });
});

/**
 * BL-329: the keyboard's way to the same thing (WCAG 2.1.1): ← → pin a
 * chord, and the frame's 「ここから再生」 plays from its start.
 */
test('with the keyboard alone, a pinned chord plays from its start', async ({ page }) => {
  const { result, position } = await open(page);
  await page.keyboard.press('ArrowRight');
  await page.keyboard.press('ArrowRight');
  const box = result.locator('.timeline-overview__scroll');
  const pinned = box.locator('[role=option][aria-selected=true]');
  await expect(pinned).toHaveCount(1);
  const bar =
    (await box
      .getByRole('option')
      .evaluateAll((all) => all.findIndex((o) => o.getAttribute('aria-selected') === 'true'))) + 1; // one chord a bar
  const play = result.getByRole('region', { name: 'コードの解説' }).getByRole('button', {
    name: 'ここから再生',
  });
  const size = await play.boundingBox();
  expect(size && size.height >= 24 && size.width >= 24).toBe(true);
  await play.focus();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
  await expect(position).toHaveText(new RegExp(`^${String(bar)}\\.`));
  await page.getByRole('button', { name: '停止', exact: true }).click();
  await expect(position).toHaveText('1.1');
});

/**
 * ADR-0069: playing, a click on the strip brings the box there and pauses
 * the following; the playhead goes on where it was.
 */
test('playing, a click on the strip moves the view, not the playhead', async ({ page }) => {
  const { result, position } = await open(page);
  await page.getByRole('button', { name: '再生' }).click();
  const strip = result.locator('svg.timeline--overview');
  const s = await strip.boundingBox();
  if (!s) throw new Error('no strip');
  await strip.click({ position: { x: s.width * 0.75, y: s.height / 2 } });
  await expect(result.getByRole('button', { name: '再生位置へ戻る' })).toBeVisible();
  await expect(position).toHaveText(/^1\./);
  const box = result.locator('.timeline-overview__scroll');
  expect(await box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(1000);
  await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
});

/** BL-334: ■ takes the playhead back to the beginning, and the box with it. */
test('stop brings the box back to the beginning', async ({ page }) => {
  const { result, position } = await open(page);
  const box = result.locator('.timeline-overview__scroll');
  await clickBar(page, 3);
  await page.getByRole('button', { name: '再生' }).click();
  await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
  await page.getByRole('button', { name: '停止', exact: true }).click();
  await expect(position).toHaveText('1.1');
  await expect.poll(() => box.evaluate((el) => el.scrollLeft)).toBe(0);
});

/**
 * BL-331 (ADR-0069): dragging the strip's frame scrolls the main box along;
 * the playhead stays. Playing, the following pauses.
 */
test('dragging the frame on the strip scrolls the box, not the playhead', async ({ page }) => {
  const { result, position } = await open(page);
  const box = result.locator('.timeline-overview__scroll');
  const frame = result.locator('svg.timeline--overview .timeline__highlight');
  await result.locator('svg.timeline--overview').scrollIntoViewIfNeeded();
  // The frame is filled faintly, so it reads as something to hold (BL-336).
  expect(await frame.evaluate((el) => getComputedStyle(el).fill)).toBe('rgba(245, 158, 11, 0.15)');
  const f = await frame.boundingBox();
  if (!f) throw new Error('no frame');
  const y = f.y + f.height / 2;
  await page.mouse.move(f.x + f.width / 2, y);
  await page.mouse.down();
  await page.mouse.move(f.x + f.width / 2 + 100, y, { steps: 5 });
  const mid = await box.evaluate((el) => el.scrollLeft);
  expect(mid).toBeGreaterThan(0);
  await page.mouse.move(f.x + f.width / 2 + 200, y, { steps: 5 });
  await page.mouse.up();
  expect(await box.evaluate((el) => el.scrollLeft)).toBeGreaterThan(mid);
  await expect(position).toHaveText('1.1');
  // The frame was held where it was pressed: it did not jump to centre.
  const after = await frame.boundingBox();
  expect(Math.abs((after?.x ?? 0) - (f.x + 200))).toBeLessThanOrEqual(12);

  // Playing, a drag on the strip pauses the following.
  await page.getByRole('button', { name: '再生' }).click();
  const g = await frame.boundingBox();
  if (!g) throw new Error('no frame');
  await page.mouse.move(g.x + g.width / 2, y);
  await page.mouse.down();
  await page.mouse.move(g.x + g.width / 2 - 80, y, { steps: 4 });
  await page.mouse.up();
  await expect(result.getByRole('button', { name: '再生位置へ戻る' })).toBeVisible();
});

/**
 * BL-335: with a mouse, pressing the drawing moves the playhead and a drag
 * carries it along until the release. Playing, the sound holds while
 * dragging and goes on from where the drag ended.
 */
test.describe('dragging the playhead with a mouse', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  async function dragBars(page: Page, from: number, to: number, during?: () => Promise<void>) {
    const bar = (n: number) =>
      page
        .locator('.timeline-overview__scroll .timeline__bar-number', {
          hasText: new RegExp(`^${String(n)}$`),
        })
        .boundingBox();
    const [a, b] = await Promise.all([bar(from), bar(to)]);
    if (!a || !b) throw new Error('no bar numbers');
    await page.mouse.move(a.x + 2, a.y + a.height / 2);
    await page.mouse.down();
    await page.mouse.move(b.x + 2, b.y + b.height / 2, { steps: 8 });
    await during?.();
    await page.mouse.up();
  }

  test('stopped, a drag sets where play starts', async ({ page }) => {
    const { position } = await open(page);
    await dragBars(page, 2, 4, async () => {
      await expect(position).toHaveText('4.1');
    });
    await expect(position).toHaveText('4.1');
    await expect(page.getByRole('button', { name: '再生' })).toBeVisible();
  });

  test('playing, the sound holds while dragging and goes on from the release', async ({ page }) => {
    const { position } = await open(page);
    await page.getByRole('button', { name: '再生' }).click();
    await expect(position).toHaveText(/^1\.[2-4]$/, { timeout: 5000 });
    await dragBars(page, 2, 3, async () => {
      await expect(page.getByRole('button', { name: '再生' })).toBeVisible();
      await expect(position).toHaveText('3.1');
    });
    await expect(page.getByRole('button', { name: '一時停止' })).toBeVisible();
    await expect(position).toHaveText(/^3\./);
  });
});

/**
 * BL-337 (ADR-0072): on a touch screen a finger holds the playhead by a
 * wide band and pulls the score: the playhead stays in the middle and a
 * swipe to the left brings later bars under it.
 */
test.describe('holding the playhead with a finger', () => {
  test.use({ viewport: { width: 390, height: 844 }, hasTouch: true, isMobile: true });

  test('a swipe on the playhead moves it a bar on, and it stays in the middle', async ({
    page,
  }) => {
    const { result, position } = await open(page);
    const box = result.locator('.timeline-overview__scroll');
    await page
      .locator('.timeline-overview__scroll .timeline__bar-number', { hasText: /^2$/ })
      .tap();
    await expect(position).toHaveText('2.1');
    const grip = box.getByTestId('head-grip');
    const g = await grip.boundingBox();
    if (!g) throw new Error('no grip');
    const x = g.x + g.width / 2;
    const y = g.y + g.height / 2;
    const cdp = await page.context().newCDPSession(page);
    const touch = (type: 'touchStart' | 'touchMove' | 'touchEnd', at: number) =>
      cdp.send('Input.dispatchTouchEvent', {
        type,
        touchPoints: type === 'touchEnd' ? [] : [{ x: at, y }],
      });
    await touch('touchStart', x);
    for (let i = 1; i <= 8; i += 1) await touch('touchMove', x - i * 20);
    await touch('touchEnd', 0);
    await expect(position).toHaveText(/^3\./);
    const offset = await box.evaluate((el) => {
      const head = el.querySelector('[data-testid=playhead]')?.getBoundingClientRect();
      const labels = el.querySelector('.timeline-labels')?.getBoundingClientRect();
      const rect = el.getBoundingClientRect();
      if (!head || !labels) return Infinity;
      return Math.abs(head.left - (labels.right + (rect.right - labels.right) / 2));
    });
    expect(offset).toBeLessThanOrEqual(8);
  });
});

/**
 * #500: the player speaks only while the music is silent: a pause and a
 * stop are said in words, and nothing is in the region while it plays.
 */
test('the player is heard only when the music stops', async ({ page }) => {
  const { result } = await open(page);
  const said = result.getByRole('group', { name: '再生' }).getByRole('status').last();
  await clickBar(page, 3);
  await expect(said).toHaveText('3 小節目 1 拍目へ');
  await page.getByRole('button', { name: '再生' }).click();
  await expect(said).toHaveText('');
  await page.getByRole('button', { name: '一時停止' }).click();
  await expect(said).toHaveText(/^3 小節目 \d 拍目で一時停止$/);
  await page.getByRole('button', { name: '停止', exact: true }).click();
  await expect(said).toHaveText('停止。先頭に戻りました');
  await expect(page.getByLabel('再生位置')).toHaveAccessibleName('再生位置 1 小節目 1 拍目');
});
