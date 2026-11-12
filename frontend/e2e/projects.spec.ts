import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * Local persistence (BL-024): a kept example survives a reload because it
 * lives in the browser's IndexedDB, and opening it restores the timeline.
 * It is kept with the star beside its title, under the candidate's title
 * (#440, BL-296).
 */
test('keep an example, reload, reopen it, let it go, delete it', async ({ page }) => {
  await page.goto('/');
  await openExample(page, '晴れた朝', '朝の光'); // C major, I - V7/vi - vi - IV
  const result = page.getByRole('region', { name: '結果' });
  await expect(result.getByRole('heading', { name: '朝の光' })).toBeVisible();

  const star = result.getByRole('button', { name: 'とっておく' });
  await expect(star).toHaveAttribute('aria-pressed', 'false');
  await expect(star).toHaveAttribute('title', 'とっておく');
  // The symbol is the only visible label, so its target is at least 24 px (WCAG 2.5.8).
  const box = await star.boundingBox();
  expect(box?.width).toBeGreaterThanOrEqual(24);
  expect(box?.height).toBeGreaterThanOrEqual(24);
  await star.click();
  await expect(star).toHaveAttribute('aria-pressed', 'true');
  await expect(star).toHaveAttribute('title', 'とっておいた例から外す');
  // The list stays folded until asked for (BL-060).
  const fold = page.getByText('とっておいた例（1）');
  await expect(fold).toBeVisible();
  await expect(page.getByRole('list', { name: 'とっておいた例の一覧' })).toBeHidden();

  // The same candidate in another key is another example: not kept yet.
  await result.getByLabel('キー').selectOption('2');
  await expect(star).toHaveAttribute('aria-pressed', 'false');
  await star.click();
  await expect(page.getByText('とっておいた例（2）')).toBeVisible();

  await page.reload();
  await expect(result).toHaveCount(0);
  // Nothing is open, so there is nothing to keep.
  await expect(page.getByRole('button', { name: 'とっておく' })).toHaveCount(0);
  await page.getByText('とっておいた例（2）').click();
  const items = page.getByRole('list', { name: 'とっておいた例の一覧' }).getByRole('listitem');
  const c = items.filter({ hasText: 'C major' });
  await expect(c).toContainText('朝の光');
  // The card was kept with it: its progression and reason show in the list (BL-299).
  await expect(c).toContainText('I - V7/vi - vi - IV');
  await expect(c).toContainText('親しみやすいポップ進行');

  await c.getByRole('button', { name: '開く' }).click();
  await expect(result.getByRole('listbox', { name: 'コード' }).getByRole('option')).toHaveText([
    /I.*C/,
    /V7\/vi.*E7/,
    /vi.*Am/,
    /IV.*F/,
  ]);
  await expect(result.getByRole('heading', { name: '朝の光' })).toBeVisible();
  // ...and the explanation opens with it, after a reload too.
  await expect(result.getByLabel('この例（AI の提案）')).toContainText('イメージ: 晴れた朝');
  await expect(result.getByLabel('この例（AI の提案）')).toContainText('親しみやすいポップ進行');
  // Opened from the list, it is kept; pressing the star again lets it go.
  await expect(star).toHaveAttribute('aria-pressed', 'true');
  await star.click();
  await expect(star).toHaveAttribute('aria-pressed', 'false');
  await expect(page.getByText('とっておいた例（1）')).toBeVisible();

  await items.getByRole('button', { name: '朝の光 を削除' }).click();
  await expect(page.getByRole('region', { name: 'とっておいた例' })).toHaveCount(0);
});
