import { expect, type Page } from '@playwright/test';

/**
 * The one entry (docs/product/experience.md): describe an image and open
 * one of the examples. Against LLM_PROVIDER=fake the answers are fixed:
 * words like 雨 or 夜 give 窓辺の雨 (A minor) and 遠い記憶 (E minor);
 * anything else gives 朝の光 (C major) and 出発の朝 (G major).
 */
export async function openExample(page: Page, image: string, title: string) {
  await page.getByLabel('イメージ', { exact: true }).fill(image);
  await page.getByRole('button', { name: '送る' }).click();
  const card = page.getByRole('article', { name: title });
  await card.getByRole('button', { name: 'この候補を開く' }).click();
  const result = page.getByRole('region', { name: '結果' });
  await expect(result).toBeVisible();
  return { card, result };
}
