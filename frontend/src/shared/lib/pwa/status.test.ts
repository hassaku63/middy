import { describe, expect, it, vi } from 'vitest';
import { PwaStatus, type InstallPrompt } from './status';

describe('PwaStatus', () => {
  it('notifies subscribers and keeps snapshots immutable', () => {
    const s = new PwaStatus();
    const listener = vi.fn();
    const unsubscribe = s.subscribe(listener);
    const before = s.getSnapshot();

    s.setOnline(false);
    s.installAvailable({ prompt: vi.fn(), userChoice: Promise.resolve({ outcome: 'dismissed' }) });

    expect(listener).toHaveBeenCalledTimes(2);
    expect(s.getSnapshot()).toMatchObject({ installable: true, online: false });
    expect(before.online).toBe(true);
    unsubscribe();
    s.setOnline(true);
    expect(listener).toHaveBeenCalledTimes(2);
    expect(s.getSnapshot().online).toBe(true);
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

  it('offers the update again when activating it fails (review)', async () => {
    const s = new PwaStatus();
    const apply = vi.fn().mockRejectedValueOnce(new Error('activate')).mockResolvedValue(undefined);
    s.needRefresh(apply);
    await s.applyUpdate();
    expect(s.getSnapshot().updateReady).toBe(true);
    await s.applyUpdate();
    expect(apply).toHaveBeenCalledTimes(2);
    expect(s.getSnapshot().updateReady).toBe(false);
  });

  it('opens the install dialog once even when clicked twice (review)', async () => {
    const s = new PwaStatus();
    let choose: (c: { outcome: 'accepted' | 'dismissed' }) => void = () => undefined;
    const prompt: InstallPrompt = {
      prompt: vi.fn().mockResolvedValue(undefined),
      userChoice: new Promise((resolve) => (choose = resolve)),
    };
    s.installAvailable(prompt);
    const first = s.promptInstall();
    const second = s.promptInstall();
    choose({ outcome: 'dismissed' });
    expect(await first).toBe(false);
    expect(await second).toBe(false);
    expect(prompt.prompt).toHaveBeenCalledTimes(1);
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
