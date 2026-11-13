import { expect, test } from '@playwright/test';
import { openExample } from './helpers';

/**
 * #498: on a phone held sideways (short and wide), opening an example
 * folds the conversation into a drawer, and the timeline and the chord's
 * frame share the screen. 「会話」 brings the conversation back with what
 * was typed in it; closing it returns the focus to 「会話」.
 */
for (const viewport of [
  { width: 844, height: 390 },
  { width: 932, height: 430 },
]) {
  test.describe(`a phone sideways at ${String(viewport.width)}×${String(viewport.height)}`, () => {
    test.use({ viewport, isMobile: true, hasTouch: true });

    test('the timeline and the chord frame share the screen; the conversation is a drawer', async ({
      page,
    }) => {
      await page.goto('/');
      const { result } = await openExample(page, '長い雨の日', '窓辺の雨（長め）');
      const conversation = page.locator('#conversation');
      await expect(conversation).toBeHidden();
      await expect(result.getByRole('group', { name: 'タイムライン' }).first()).toBeInViewport();
      const frame = result.getByRole('region', { name: 'コードの解説' });
      await expect(frame.getByRole('heading', { level: 2 })).toBeInViewport();
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
        viewport.width,
      );
      expect(await page.evaluate(() => document.documentElement.scrollHeight)).toBeLessThanOrEqual(
        viewport.height,
      );

      // The drawer opens over the result, with what was typed still there.
      const toggle = page.getByRole('button', { name: '会話', exact: true });
      await expect(toggle).toHaveAttribute('aria-expanded', 'false');
      await toggle.click();
      await expect(conversation).toBeVisible();
      await expect(page.getByRole('button', { name: '会話を閉じる' })).toBeFocused();
      await page.getByLabel('イメージ', { exact: true }).fill('もう少し明るく');
      await page.getByRole('button', { name: '会話を閉じる' }).click();
      await expect(conversation).toBeHidden();
      await expect(toggle).toBeFocused();
      await toggle.click();
      await expect(page.getByLabel('イメージ', { exact: true })).toHaveValue('もう少し明るく');
    });
  });
}

test.describe('a phone held upright', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('keeps the conversation in the page and shows no drawer controls', async ({ page }) => {
    await page.goto('/');
    await openExample(page, '長い雨の日', '窓辺の雨（長め）');
    await expect(page.locator('#conversation')).toBeVisible();
    await expect(page.getByRole('button', { name: '会話', exact: true })).toHaveCount(0);
  });
});
