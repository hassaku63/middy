import { describe, expect, it, vi } from 'vitest';
import { PwaStatus } from './status';
import { CHECK_EVERY_MS, wirePwa } from './wire';

describe('wirePwa', () => {
  it('feeds service worker callbacks and window events into the store', async () => {
    const status = new PwaStatus();
    const updateSW = vi.fn().mockResolvedValue(undefined);
    let callbacks: { onNeedRefresh?: () => void } = {};
    const registerSW = vi.fn((opts: typeof callbacks) => {
      callbacks = opts;
      return updateSW;
    });
    const target = new EventTarget();

    wirePwa(status, registerSW, target);

    callbacks.onNeedRefresh?.();
    await status.applyUpdate();
    expect(updateSW).toHaveBeenCalledWith(true);

    target.dispatchEvent(new Event('offline'));
    expect(status.getSnapshot().online).toBe(false);
    target.dispatchEvent(new Event('online'));
    expect(status.getSnapshot().online).toBe(true);

    const install = new Event('beforeinstallprompt', { cancelable: true });
    Object.assign(install, {
      prompt: vi.fn().mockResolvedValue(undefined),
      userChoice: Promise.resolve({ outcome: 'dismissed' }),
    });
    target.dispatchEvent(install);
    expect(install.defaultPrevented).toBe(true);
    expect(status.getSnapshot().installable).toBe(true);
    target.dispatchEvent(new Event('appinstalled'));
    expect(status.getSnapshot().installable).toBe(false);
  });

  it('asks for a new version every hour while online (#500)', () => {
    const status = new PwaStatus();
    const update = vi.fn().mockResolvedValue(undefined);
    let checks: { check: () => void; ms: number }[] = [];
    const registerSW = vi.fn(
      (opts: { onRegisteredSW?: (u: string, r: ServiceWorkerRegistration) => void }) => {
        opts.onRegisteredSW?.('/sw.js', { update } as unknown as ServiceWorkerRegistration);
        return vi.fn().mockResolvedValue(undefined);
      },
    );
    const target = new EventTarget();
    wirePwa(status, registerSW, target, (check, ms) => {
      checks = [...checks, { check, ms }];
    });
    expect(checks.map((c) => c.ms)).toEqual([CHECK_EVERY_MS]);

    checks[0]?.check();
    expect(update).toHaveBeenCalledTimes(1);
    target.dispatchEvent(new Event('offline'));
    checks[0]?.check();
    expect(update).toHaveBeenCalledTimes(1); // not while offline
  });
});
