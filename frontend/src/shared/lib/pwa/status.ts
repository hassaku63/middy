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
    this.state = { updateReady: false, installable: false, online };
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

  /** Activates the waiting version; if that fails, the offer stays so it can be tried again. */
  async applyUpdate(): Promise<void> {
    const apply = this.update;
    if (!apply) return;
    this.update = null;
    this.set({ updateReady: false });
    try {
      await apply();
    } catch {
      this.needRefresh(apply);
    }
  }

  /**
   * Shows the browser's install dialog; resolves true when accepted. The
   * offer is withdrawn before the dialog opens, because the browser shows
   * it only once: a second click must not call prompt() again.
   */
  async promptInstall(): Promise<boolean> {
    const install = this.install;
    if (!install) return false;
    this.installed();
    await install.prompt();
    const { outcome } = await install.userChoice;
    return outcome === 'accepted';
  }

  private set(patch: Partial<PwaState>): void {
    this.state = { ...this.state, ...patch };
    for (const l of this.listeners) l();
  }
}

/** The process-wide instance main.tsx wires up and the UI reads. */
export const pwaStatus = new PwaStatus(typeof navigator === 'undefined' ? true : navigator.onLine);
