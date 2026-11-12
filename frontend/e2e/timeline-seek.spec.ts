import { expect, test, type Page } from '@playwright/test';

/**
 * BL-328 (the owner's rules): clicking the timeline anywhere but on a
 * chord moves the playhead to the start of the beat under it. Stopped, the
 * next play starts there and stop comes back to it; playing, play jumps
 * there. A chord still pins. The fake LLM's 長い例 is 16 bars of 4/4.
 */
async function open(page: Page) {
  await page.goto('/');
  await page.getByLabel('イメージ', { exact: true }).fill('長い雨の日');
  await page.getByRole('button', { name: '送る' }).click();
  await page.getByRole('article').first().getByRole('button').click();
  // Choosing plays it (#315): stop, so each test starts stopped.
  const stop = page.getByRole('button', { name: '停止' });
  await expect(stop).toBeVisible();
  if (await stop.isEnabled()) await stop.click();
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

    test('stopped, a click sets where play starts; stop comes back to it', async ({ page }) => {
      const { position } = await open(page);
      await clickBar(page, 3);
      await expect(position).toHaveText('3.1');
      await page.getByRole('button', { name: '再生' }).click();
      await expect(position).toHaveText(/^3\.[2-4]$/, { timeout: 5000 });
      await page.getByRole('button', { name: '停止' }).click();
      await expect(position).toHaveText('3.1');
    });

    test('playing, a click moves play there', async ({ page }) => {
      const { position } = await open(page);
      await page.getByRole('button', { name: '再生' }).click();
      await expect(position).toHaveText(/^1\.[2-4]$/, { timeout: 5000 });
      await clickBar(page, 2);
      await expect(position).toHaveText(/^2\./);
      await expect(page.getByRole('button', { name: '停止' })).toBeVisible();
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
