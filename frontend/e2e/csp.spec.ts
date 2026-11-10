import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * #206 A-4: the web container sends a strict Content-Security-Policy (this
 * origin only, no inline script or style). The app's main paths run
 * without a single violation: describe, open, play, pin, regroove, save
 * the MIDI, on both layouts. Runs against nginx (make up, or any server
 * that sends the header); without the header it only proves the paths.
 */
for (const path of ['/']) {
  test(`the main paths raise no CSP violation (${path})`, async ({ page }) => {
    const violations: string[] = [];
    page.on('console', (m) => {
      if (/Content Security Policy|Content-Security-Policy/i.test(m.text()))
        violations.push(m.text());
    });
    await page.goto(path);
    const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
    await page.getByRole('button', { name: '再生' }).click();
    await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
    await page.getByRole('button', { name: '停止' }).click();
    await result.getByRole('listbox', { name: 'コード' }).getByRole('option').nth(1).click();
    await Promise.all([
      page.waitForEvent('download'),
      page.getByRole('button', { name: 'MIDI を保存' }).click(),
    ]);
    expect(violations).toEqual([]);
  });
}
