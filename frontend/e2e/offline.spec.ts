import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * The offline promise of the PWA (BL-025): once the service worker has
 * cached the shell, a saved project opens and plays with no network, and
 * the banner says what still works.
 */
test('a saved project opens offline', async ({ page, context }) => {
  await page.goto('/');
  // Wait until the service worker controls the page, so reloads come from cache.
  await expect
    .poll(() => page.evaluate(async () => (await navigator.serviceWorker.ready).active?.state))
    .toBe('activated');
  await page.reload();
  await expect
    .poll(() => page.evaluate(() => navigator.serviceWorker.controller !== null))
    .toBe(true);

  await openExample(page, '晴れた朝', '朝の光'); // C major, I - V7/vi - vi - IV
  await page.getByLabel('例の名前').fill('オフライン用');
  await page.getByRole('button', { name: '保存', exact: true }).click();
  await expect(page.getByText('保存した例（1）')).toBeVisible();

  await context.setOffline(true);
  await page.reload();
  await expect(page.getByText(/オフラインです。保存した例は開いて再生できます/)).toBeVisible();

  await page.getByText('保存した例（1）').click();
  const item = page.getByRole('list', { name: '保存した例の一覧' }).getByRole('listitem');
  await item.getByRole('button', { name: '開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /I.*C/,
    /V7\/vi.*E7/,
    /vi.*Am/,
    /IV.*F/,
  ]);
  await page.getByRole('button', { name: '再生' }).click();
  await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
  await page.getByRole('button', { name: '停止' }).click();

  await context.setOffline(false);
  await expect(page.getByText(/オフラインです/)).toHaveCount(0);
});
