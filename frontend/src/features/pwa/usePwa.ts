import { useSyncExternalStore } from 'react';
import { pwaStatus, type PwaStatus } from '@/shared/lib/pwa/status';

/** Live PWA state plus the two actions the banner offers. */
export function usePwa(status: PwaStatus = pwaStatus) {
  const state = useSyncExternalStore(status.subscribe, status.getSnapshot, status.getSnapshot);
  return {
    ...state,
    applyUpdate: () => status.applyUpdate(),
    promptInstall: () => status.promptInstall(),
  };
}

export type Pwa = ReturnType<typeof usePwa>;
