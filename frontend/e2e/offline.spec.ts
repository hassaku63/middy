import { expect, test } from '@playwright/test';

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

  await page.getByRole('button', { name: '進行から' }).click();
  await page.getByLabel('進行（ローマ数字）').fill('I - IV - V - I');
  await page.getByRole('button', { name: '生成' }).click();
  await expect(page.getByRole('region', { name: '結果' })).toBeVisible();
  await page.getByLabel('プロジェクト名').fill('オフライン用');
  await page.getByRole('button', { name: '保存', exact: true }).click();
  await expect(page.getByRole('list', { name: 'プロジェクト一覧' })).toContainText('オフライン用');

  await context.setOffline(true);
  await page.reload();
  await expect(
    page.getByText(/オフラインです。保存したプロジェクトは開いて再生できます/),
  ).toBeVisible();

  const item = page.getByRole('list', { name: 'プロジェクト一覧' }).getByRole('listitem');
  await item.getByRole('button', { name: '開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('option')).toHaveText([/I.*C/, /IV.*F/, /V.*G/, /I.*C/]);
  await page.getByRole('button', { name: '再生' }).click();
  await expect(result.locator('[role=option][aria-current=true]')).toBeVisible();
  await page.getByRole('button', { name: '停止' }).click();

  await context.setOffline(false);
  await expect(page.getByText(/オフラインです/)).toHaveCount(0);
});
