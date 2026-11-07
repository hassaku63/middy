import { expect, test } from '@playwright/test';

/**
 * Local persistence (BL-024): a saved project survives a reload because it
 * lives in the browser's IndexedDB, and opening it restores the timeline.
 */
test('save a project, reload, reopen it, delete it', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: '進行から' }).click();
  await page.getByLabel('進行（ローマ数字）').fill('ii - V7 - I');
  await page.getByRole('button', { name: '生成' }).click();
  await expect(page.getByRole('region', { name: '結果' })).toBeVisible();

  await page.getByLabel('プロジェクト名').fill('ツーファイブ');
  await page.getByRole('button', { name: '保存', exact: true }).click();
  const list = page.getByRole('list', { name: 'プロジェクト一覧' });
  await expect(list.getByRole('listitem')).toHaveCount(1);

  await page.reload();
  await expect(page.getByRole('region', { name: '結果' })).toHaveCount(0);
  const item = page.getByRole('list', { name: 'プロジェクト一覧' }).getByRole('listitem');
  await expect(item).toContainText('ツーファイブ');
  await expect(item).toContainText('C major');

  await item.getByRole('button', { name: '開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('option')).toHaveText([/ii7?.*Dm/, /V7.*G7/, /I.*C/]);
  await expect(result).toContainText('プロジェクト: ツーファイブ');

  await item.getByRole('button', { name: 'ツーファイブ を削除' }).click();
  await expect(page.getByRole('list', { name: 'プロジェクト一覧' })).toHaveCount(0);
});
