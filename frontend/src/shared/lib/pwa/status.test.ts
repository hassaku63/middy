import { describe, expect, it, vi } from 'vitest';
import { PwaStatus, type InstallPrompt } from './status';

describe('PwaStatus', () => {
  it('notifies subscribers and keeps snapshots immutable', () => {
    const s = new PwaStatus();
    const listener = vi.fn();
    const unsubscribe = s.subscribe(listener);
    const before = s.getSnapshot();

    s.offlineIsReady();
    s.setOnline(false);

    expect(listener).toHaveBeenCalledTimes(2);
    expect(s.getSnapshot()).toMatchObject({ offlineReady: true, online: false });
    expect(before.offlineReady).toBe(false);
    unsubscribe();
    s.dismissOfflineReady();
    expect(listener).toHaveBeenCalledTimes(2);
    expect(s.getSnapshot().offlineReady).toBe(false);
  });

  it('applies a waiting update once', async () => {
    const s = new PwaStatus();
    const apply = vi.fn().mockResolvedValue(undefined);
    s.needRefresh(apply);
    expect(s.getSnapshot().updateReady).toBe(true);

    await s.applyUpdate();
    await s.applyUpdate();
    expect(apply).toHaveBeenCalledTimes(1);
    expect(s.getSnapshot().updateReady).toBe(false);
  });

  it('prompts for install and reports the outcome', async () => {
    const s = new PwaStatus();
    expect(await s.promptInstall()).toBe(false);

    const prompt: InstallPrompt = {
      prompt: vi.fn().mockResolvedValue(undefined),
      userChoice: Promise.resolve({ outcome: 'accepted' }),
    };
    s.installAvailable(prompt);
    expect(s.getSnapshot().installable).toBe(true);
    expect(await s.promptInstall()).toBe(true);
    expect(prompt.prompt).toHaveBeenCalledTimes(1);
    expect(s.getSnapshot().installable).toBe(false);
  });
});
