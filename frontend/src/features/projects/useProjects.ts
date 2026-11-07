import { useCallback, useEffect, useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import {
  defaultProjectStore,
  newProjectId,
  type Project,
  type ProjectSource,
  type ProjectStore,
} from '@/shared/lib/storage/projects';

export type Projects = {
  projects: Project[];
  save: (name: string, timeline: Timeline, source: ProjectSource) => Promise<Project>;
  remove: (id: string) => Promise<void>;
  error: string | null;
};

/** Owns the saved-project list; the store is injectable for tests. */
export function useProjects(store: ProjectStore = sharedStore()): Projects {
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      setProjects(await store.list());
      setError(null);
    } catch {
      setError('保存したプロジェクトを読み込めませんでした');
    }
  }, [store]);

  // Initial load: setState only in the promise callbacks, never synchronously in the effect.
  useEffect(() => {
    let cancelled = false;
    store
      .list()
      .then((list) => {
        if (!cancelled) setProjects(list);
      })
      .catch(() => {
        if (!cancelled) setError('保存したプロジェクトを読み込めませんでした');
      });
    return () => {
      cancelled = true;
    };
  }, [store]);

  const save = useCallback(
    async (name: string, timeline: Timeline, source: ProjectSource) => {
      const now = Date.now();
      const project: Project = {
        id: newProjectId(),
        name: name.trim() || defaultName(source),
        createdAt: now,
        updatedAt: now,
        timeline,
        source,
      };
      await store.save(project);
      await refresh();
      return project;
    },
    [store, refresh],
  );

  const remove = useCallback(
    async (id: string) => {
      await store.remove(id);
      await refresh();
    },
    [store, refresh],
  );

  return { projects, save, remove, error };
}

function defaultName(source: ProjectSource): string {
  return source.kind === 'idea'
    ? source.candidateTitle
    : source.request.progression.map((p) => p.numeral).join('-');
}

let cached: ProjectStore | null = null;
function sharedStore(): ProjectStore {
  cached ??= defaultProjectStore();
  return cached;
}
