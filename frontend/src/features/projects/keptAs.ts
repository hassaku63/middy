import type { Timeline } from '@/shared/api/client';
import type { Project } from '@/shared/lib/storage/projects';

/**
 * The kept example that is this timeline, if any. An example is what it
 * sounds like: the same candidate in another key or style is another
 * example, and keeping it twice is not possible (BL-296).
 */
export function keptAs(kept: Project[], timeline: Timeline): Project | undefined {
  const sound = JSON.stringify(timeline);
  return kept.find((p) => JSON.stringify(p.timeline) === sound);
}
