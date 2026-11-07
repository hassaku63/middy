import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * Local persistence (BL-024): a saved project survives a reload because it
 * lives in the browser's IndexedDB, and opening it restores the timeline.
 */
test('save a project, reload, reopen it, delete it', async ({ page }) => {
  await page.goto('/');
  await openExample(page, '晴れた朝', '朝の光'); // C major, I - V - vi - IV

  await page.getByLabel('プロジェクト名').fill('朝の練習');
  await page.getByRole('button', { name: '保存', exact: true }).click();
  const list = page.getByRole('list', { name: 'プロジェクト一覧' });
  await expect(list.getByRole('listitem')).toHaveCount(1);

  await page.reload();
  await expect(page.getByRole('region', { name: '結果' })).toHaveCount(0);
  const item = page.getByRole('list', { name: 'プロジェクト一覧' }).getByRole('listitem');
  await expect(item).toContainText('朝の練習');
  await expect(item).toContainText('C major');

  await item.getByRole('button', { name: '開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('option')).toHaveText([/I.*C/, /V.*G/, /vi.*Am/, /IV.*F/]);
  await expect(result).toContainText('プロジェクト: 朝の練習');

  await item.getByRole('button', { name: '朝の練習 を削除' }).click();
  await expect(page.getByRole('list', { name: 'プロジェクト一覧' })).toHaveCount(0);
});
