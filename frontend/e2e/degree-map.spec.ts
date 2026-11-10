import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * BL-099: the explanation shows the seven chords of the example's mode, the
 * ones the example plays stand out, and the explained chord lights up. In
 * a mode they are grouped by role (ADR-0048). It must fit a phone without a
 * horizontal scroll.
 */
test.describe('degree map on a phone', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('groups the dorian chords by role, lights the pinned IV, and does not overflow', async ({
    page,
  }) => {
    await page.goto('/');
    const { result } = await openExample(page, 'アイルランドのパブのケルト音楽', '緑の丘のリール');
    const map = result.getByRole('group', { name: '度数の地図（和音の役割）' });
    await expect(map.getByRole('listitem')).toHaveText([
      'iDm',
      'iiEm',
      'IVG',
      'IIIF',
      'vAm',
      'VIIC',
      'vi°Bdim',
    ]);
    await expect(map.getByRole('list', { name: '特徴和音' }).getByRole('listitem')).toHaveText([
      'iiEm',
      'IVG',
    ]);
    await result
      .getByRole('listbox', { name: 'コード' })
      .getByRole('option', { name: 'IV G' })
      .click();
    await expect(map.locator('[aria-current="true"]')).toHaveText('IVG');
    await expect(result.getByText('IV G 特徴和音')).toHaveAttribute('aria-live', 'polite');
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
      390,
    );
  });
});
