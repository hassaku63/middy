import type { InstallPrompt, PwaStatus } from './status';

type RegisterSW = (options: {
  immediate?: boolean;
  onNeedRefresh?: () => void;
  onRegisteredSW?: (url: string, registration: ServiceWorkerRegistration | undefined) => void;
}) => (reloadPage?: boolean) => Promise<void>;

type Every = (check: () => void, ms: number) => void;

/**
 * How often an app left open asks for a new version (#500): the browser
 * only looks when a page is opened, and an installed app may stay open
 * for days.
 */
export const CHECK_EVERY_MS = 60 * 60 * 1000;

/**
 * Connects the service worker registration and window events to the
 * status store. Called once from main.tsx; `registerSW`, `target` and
 * `every` are parameters so this can be tested without a browser.
 */
export function wirePwa(
  status: PwaStatus,
  registerSW: RegisterSW,
  target: EventTarget = window,
  every: Every = (check, ms) => {
    window.setInterval(check, ms);
  },
): void {
  const updateSW = registerSW({
    immediate: true,
    // Being ready offline is the normal state and needs no notice (BL-071).
    onNeedRefresh: () => status.needRefresh(() => updateSW(true)),
    // Asks again now and then; a new version shows as the update banner.
    onRegisteredSW: (_url, registration) => {
      if (!registration) return;
      every(() => {
        if (!status.getSnapshot().online) return;
        registration.update().catch(() => undefined); // tried again next time
      }, CHECK_EVERY_MS);
    },
  });

  target.addEventListener('beforeinstallprompt', (event) => {
    // Keep the browser's mini-infobar quiet; we offer our own button.
    event.preventDefault();
    if (isInstallPrompt(event)) status.installAvailable(event);
  });
  target.addEventListener('appinstalled', () => status.installed());
  target.addEventListener('online', () => status.setOnline(true));
  target.addEventListener('offline', () => status.setOnline(false));
}

/** beforeinstallprompt is not in the DOM typings; check the shape instead of casting. */
function isInstallPrompt(event: Event): event is Event & InstallPrompt {
  return 'prompt' in event && typeof event.prompt === 'function' && 'userChoice' in event;
}
