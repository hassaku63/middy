/**
 * App-wide PWA state (service worker lifecycle, install prompt,
 * connectivity) as a tiny external store. main.tsx feeds it from the
 * service-worker registration and window events; React reads it with
 * useSyncExternalStore. Kept framework-free so it is trivially testable.
 */

/** The subset of the beforeinstallprompt event we use (Chromium only). */
export type InstallPrompt = {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
};

export type PwaState = {
  /** A new version is waiting; call applyUpdate() to switch to it. */
  updateReady: boolean;
  /** The app shell is cached and the app works offline. */
  offlineReady: boolean;
  /** The browser offered to install the app. */
  installable: boolean;
  online: boolean;
};

type Listener = () => void;

export class PwaStatus {
  private state: PwaState;
  private readonly listeners = new Set<Listener>();
  private update: (() => Promise<void>) | null = null;
  private install: InstallPrompt | null = null;

  constructor(online = true) {
    this.state = { updateReady: false, offlineReady: false, installable: false, online };
  }

  getSnapshot = (): PwaState => this.state;

  subscribe = (listener: Listener): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  /** Service worker found a new version; `apply` activates it and reloads. */
  needRefresh(apply: () => Promise<void>): void {
    this.update = apply;
    this.set({ updateReady: true });
  }

  offlineIsReady(): void {
    this.set({ offlineReady: true });
  }

  dismissOfflineReady(): void {
    this.set({ offlineReady: false });
  }

  installAvailable(prompt: InstallPrompt): void {
    this.install = prompt;
    this.set({ installable: true });
  }

  installed(): void {
    this.install = null;
    this.set({ installable: false });
  }

  setOnline(online: boolean): void {
    this.set({ online });
  }

  async applyUpdate(): Promise<void> {
    const apply = this.update;
    if (!apply) return;
    this.update = null;
    this.set({ updateReady: false });
    await apply();
  }

  /** Shows the browser's install dialog; resolves true when accepted. */
  async promptInstall(): Promise<boolean> {
    const install = this.install;
    if (!install) return false;
    await install.prompt();
    const { outcome } = await install.userChoice;
    this.installed();
    return outcome === 'accepted';
  }

  private set(patch: Partial<PwaState>): void {
    this.state = { ...this.state, ...patch };
    for (const l of this.listeners) l();
  }
}

/** The process-wide instance main.tsx wires up and the UI reads. */
export const pwaStatus = new PwaStatus(typeof navigator === 'undefined' ? true : navigator.onLine);
