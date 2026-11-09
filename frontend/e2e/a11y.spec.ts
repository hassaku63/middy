import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';
import { openExample } from './helpers';

/**
 * Automated accessibility checks with axe-core (BL-162, #248) on the main
 * screens: the first screen, an opened example, and the explanation of a
 * pinned chord. axe finds part of WCAG's failures by machine; the rest stays
 * with the manual checks in docs/design/ux-principles.md.
 */
async function violations(page: Page) {
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
    .analyze();
  return results.violations.map((v) => `${v.id}: ${v.help} (${String(v.nodes.length)})`);
}

test('the first screen has no axe violations', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByLabel('イメージ', { exact: true })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});

test('an opened example has no axe violations', async ({ page }) => {
  await page.goto('/');
  await openExample(page, '雨の日曜の午後', '窓辺の雨');
  expect(await violations(page)).toEqual([]);
});

test('the explanation of a pinned chord has no axe violations', async ({ page }) => {
  await page.goto('/');
  const { result } = await openExample(page, '雨の日曜の午後', '遠い記憶');
  await result.getByRole('option', { name: 'V7 B7' }).click();
  await expect(page.getByRole('button', { name: /固定を解除/ })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});

test('the desk layout under trial has no axe violations (BL-184)', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/?layout=desk');
  const { result } = await openExample(page, '雨の日曜の午後', '窓辺の雨');
  expect(await violations(page)).toEqual([]);
  // Pinned, the chord frame and the ■ on the chord lane (BL-185).
  await result.getByRole('option', { name: 'VII G' }).click();
  await expect(result.getByRole('region', { name: 'コードの解説' })).toHaveClass(/is-pinned/);
  expect(await violations(page)).toEqual([]);
  // A second answer: the first one becomes the list of earlier examples (BL-186).
  await openExample(page, 'もう少し明るく', '朝の光');
  await expect(page.getByRole('list', { name: 'これまでの例' })).toBeVisible();
  expect(await violations(page)).toEqual([]);
});

test.describe('in the dark theme', () => {
  test.use({ colorScheme: 'dark' });

  test('an opened example has no axe violations', async ({ page }) => {
    await page.goto('/');
    await openExample(page, '雨の日曜の午後', '窓辺の雨');
    expect(await violations(page)).toEqual([]);
  });
});
