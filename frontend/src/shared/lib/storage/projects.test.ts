import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { MemoryProjectStore, newProjectId, type Project } from './projects';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [],
};

function project(name: string, updatedAt: number): Project {
  return {
    id: newProjectId(),
    name,
    createdAt: updatedAt,
    updatedAt,
    timeline,
    source: { kind: 'idea', image: '雨', candidateTitle: name },
  };
}

describe('MemoryProjectStore', () => {
  it('lists newest first, gets, overwrites and removes', async () => {
    const store = new MemoryProjectStore();
    const a = project('a', 1);
    const b = project('b', 2);
    await store.save(a);
    await store.save(b);
    expect((await store.list()).map((p) => p.name)).toEqual(['b', 'a']);
    expect(await store.get(a.id)).toEqual(a);

    await store.save({ ...a, name: 'a2', updatedAt: 3 });
    expect((await store.list()).map((p) => p.name)).toEqual(['a2', 'b']);

    await store.remove(b.id);
    expect(await store.list()).toHaveLength(1);
    expect(await store.get(b.id)).toBeUndefined();
  });

  it('generates distinct ids', () => {
    expect(newProjectId()).not.toBe(newProjectId());
  });
});
