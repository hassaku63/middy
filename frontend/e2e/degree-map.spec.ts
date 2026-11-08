import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * BL-099 (trial): ?degrees shows the seven chords of the example's mode as
 * a band, the ones the example plays stand out, and the explained chord
 * lights up. It must fit a phone without a horizontal scroll.
 */
test.describe('degree map on a phone', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('shows the dorian chords, lights the pinned IV, and does not overflow', async ({ page }) => {
    await page.goto('/?degrees');
    const { result } = await openExample(page, 'アイルランドのパブのケルト音楽', '緑の丘のリール');
    const map = result.getByRole('list', { name: '度数の地図' });
    await expect(map.getByRole('listitem')).toHaveText([
      'iDm',
      'iiEm',
      'IIIF',
      'IVG',
      'vAm',
      'vi°Bdim',
      'VIIC',
    ]);
    await result
      .getByRole('listbox', { name: 'コード' })
      .getByRole('option', { name: 'IV (G)' })
      .click();
    await expect(map.locator('[aria-current="true"]')).toHaveText('IVG');
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
      390,
    );
  });

  test('is not shown without ?degrees', async ({ page }) => {
    await page.goto('/');
    const { result } = await openExample(page, 'アイルランドのパブのケルト音楽', '緑の丘のリール');
    await expect(result.getByLabel('全体と流れ')).toBeVisible();
    await expect(result.getByRole('list', { name: '度数の地図' })).toHaveCount(0);
  });
});
