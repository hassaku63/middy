import { useCallback, useEffect, useRef, useState } from 'react';
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
  /** Resolves to the saved project, or null when the store refused it (see error). */
  save: (name: string, timeline: Timeline, source: ProjectSource) => Promise<Project | null>;
  remove: (id: string) => Promise<void>;
  error: string | null;
};

/** Owns the saved-project list; the store is injectable for tests. */
export function useProjects(store: ProjectStore = sharedStore()): Projects {
  const [projects, setProjects] = useState<Project[]>([]);
  const [error, setError] = useState<string | null>(null);
  // Numbers each list() so only the latest one lands: a slow first load
  // must not put back a list from before a save or delete.
  const reads = useRef(0);

  const refresh = useCallback(async () => {
    const read = ++reads.current;
    try {
      const list = await store.list();
      if (read !== reads.current) return;
      setProjects(list);
      setError(null);
    } catch {
      if (read === reads.current) setError('保存した例を読み込めませんでした');
    }
  }, [store]);

  // Initial load: setState only in the promise callbacks, never synchronously in the effect.
  useEffect(() => {
    let cancelled = false;
    const read = ++reads.current;
    const current = () => !cancelled && read === reads.current;
    store
      .list()
      .then((list) => {
        if (current()) setProjects(list);
      })
      .catch(() => {
        if (current()) setError('保存した例を読み込めませんでした');
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
      try {
        await store.save(project);
      } catch {
        setError('保存できませんでした（ブラウザの保存領域を確かめてください）');
        return null;
      }
      await refresh();
      return project;
    },
    [store, refresh],
  );

  const remove = useCallback(
    async (id: string) => {
      try {
        await store.remove(id);
      } catch {
        setError('削除できませんでした');
        return;
      }
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
