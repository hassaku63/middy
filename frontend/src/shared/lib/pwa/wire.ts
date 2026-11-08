import type { InstallPrompt, PwaStatus } from './status';

type RegisterSW = (options: {
  immediate?: boolean;
  onNeedRefresh?: () => void;
}) => (reloadPage?: boolean) => Promise<void>;

/**
 * Connects the service worker registration and window events to the
 * status store. Called once from main.tsx; `registerSW` and `target` are
 * parameters so this can be tested without a browser.
 */
export function wirePwa(
  status: PwaStatus,
  registerSW: RegisterSW,
  target: EventTarget = window,
): void {
  const updateSW = registerSW({
    immediate: true,
    // Being ready offline is the normal state and needs no notice (BL-071).
    onNeedRefresh: () => status.needRefresh(() => updateSW(true)),
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
