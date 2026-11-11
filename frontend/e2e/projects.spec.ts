import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * Local persistence (BL-024): a saved project survives a reload because it
 * lives in the browser's IndexedDB, and opening it restores the timeline.
 * It is saved in one click, under the candidate's title (#440).
 */
test('save a project, reload, reopen it, delete it', async ({ page }) => {
  await page.goto('/');
  await openExample(page, '晴れた朝', '朝の光'); // C major, I - V7/vi - vi - IV

  await page.getByRole('button', { name: '保存', exact: true }).click();
  // The saved list stays folded until asked for (BL-060).
  const fold = page.getByText('保存した例（1）');
  await expect(fold).toBeVisible();
  await expect(page.getByRole('list', { name: '保存した例の一覧' })).toBeHidden();

  await page.reload();
  await expect(page.getByRole('region', { name: '結果' })).toHaveCount(0);
  // Nothing is open, so there is nothing to save.
  await expect(page.getByRole('button', { name: '保存', exact: true })).toHaveCount(0);
  await page.getByText('保存した例（1）').click();
  const item = page.getByRole('list', { name: '保存した例の一覧' }).getByRole('listitem');
  await expect(item).toContainText('朝の光');
  await expect(item).toContainText('C major');

  await item.getByRole('button', { name: '開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /I.*C/,
    /V7\/vi.*E7/,
    /vi.*Am/,
    /IV.*F/,
  ]);
  await expect(result).toContainText('保存した例: 朝の光');

  await item.getByRole('button', { name: '朝の光 を削除' }).click();
  await expect(page.getByRole('region', { name: '保存した例' })).toHaveCount(0);
});
