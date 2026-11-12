import type { FromProgressionRequest, IdeaCandidate, Timeline } from '@/shared/api/client';

/** Where a project came from, so the UI can show it and later re-run it. */
export type ProjectSource =
  | { kind: 'progression'; request: FromProgressionRequest }
  | {
      kind: 'idea';
      image: string;
      candidateTitle: string;
      /**
       * The card as the AI proposed it, kept with the example so its reason
       * and patterns can be read again (BL-299). Absent in examples kept
       * before it was stored.
       */
      candidate?: IdeaCandidate;
    };

export type Project = {
  id: string;
  name: string;
  createdAt: number;
  updatedAt: number;
  timeline: Timeline;
  source: ProjectSource;
};

/** Minimal persistence port; implementations must not throw on empty stores. */
export type ProjectStore = {
  list(): Promise<Project[]>;
  get(id: string): Promise<Project | undefined>;
  save(project: Project): Promise<void>;
  remove(id: string): Promise<void>;
};

export function newProjectId(): string {
  return typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

/** In-memory store for tests and for browsers without IndexedDB. */
export class MemoryProjectStore implements ProjectStore {
  private readonly items = new Map<string, Project>();

  list(): Promise<Project[]> {
    return Promise.resolve([...this.items.values()].sort((a, b) => b.updatedAt - a.updatedAt));
  }
  get(id: string): Promise<Project | undefined> {
    return Promise.resolve(this.items.get(id));
  }
  save(project: Project): Promise<void> {
    this.items.set(project.id, project);
    return Promise.resolve();
  }
  remove(id: string): Promise<void> {
    this.items.delete(id);
    return Promise.resolve();
  }
}

const DB_NAME = 'middy';
const DB_VERSION = 1;
const STORE = 'projects';

/**
 * IndexedDB-backed store: one object store keyed by id. Kept dependency-free;
 * every call opens a short transaction so there is no connection to manage.
 */
export class IndexedDbProjectStore implements ProjectStore {
  constructor(private readonly factory: IDBFactory = indexedDB) {}

  static isAvailable(): boolean {
    return typeof indexedDB !== 'undefined';
  }

  async list(): Promise<Project[]> {
    const all = await this.run<Project[]>('readonly', (store) => store.getAll());
    return all.sort((a, b) => b.updatedAt - a.updatedAt);
  }
  get(id: string): Promise<Project | undefined> {
    return this.run<Project | undefined>('readonly', (store) => store.get(id));
  }
  async save(project: Project): Promise<void> {
    await this.run('readwrite', (store) => store.put(project));
  }
  async remove(id: string): Promise<void> {
    await this.run('readwrite', (store) => store.delete(id));
  }

  private open(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
      const req = this.factory.open(DB_NAME, DB_VERSION);
      req.onupgradeneeded = () => {
        if (!req.result.objectStoreNames.contains(STORE)) {
          req.result.createObjectStore(STORE, { keyPath: 'id' });
        }
      };
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error ?? new Error('indexedDB open failed'));
    });
  }

  private async run<T>(
    mode: IDBTransactionMode,
    op: (store: IDBObjectStore) => IDBRequest<T> | IDBRequest<IDBValidKey> | IDBRequest<undefined>,
  ): Promise<T> {
    const db = await this.open();
    try {
      // Settle on the transaction, not the request: a write can succeed as
      // a request and still abort at commit (QuotaExceededError).
      return await new Promise<T>((resolve, reject) => {
        const tx = db.transaction(STORE, mode);
        const req = op(tx.objectStore(STORE));
        tx.oncomplete = () => resolve(req.result as T);
        tx.onerror = () => reject(tx.error ?? req.error ?? new Error('indexedDB request failed'));
        tx.onabort = () => reject(tx.error ?? new Error('indexedDB transaction aborted'));
      });
    } finally {
      db.close();
    }
  }
}

/** Picks IndexedDB when the browser has it, else memory (lost on reload). */
export function defaultProjectStore(): ProjectStore {
  return IndexedDbProjectStore.isAvailable()
    ? new IndexedDbProjectStore()
    : new MemoryProjectStore();
}
