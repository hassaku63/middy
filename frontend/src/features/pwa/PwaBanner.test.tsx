import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { InstallButton, PwaBanner } from './PwaBanner';
import type { Pwa } from './usePwa';

function pwa(partial: Partial<Pwa>): Pwa {
  return {
    updateReady: false,
    installable: false,
    online: true,
    applyUpdate: vi.fn().mockResolvedValue(undefined),
    promptInstall: vi.fn().mockResolvedValue(true),
    ...partial,
  };
}

describe('PwaBanner', () => {
  it('renders nothing in the normal state', () => {
    const { container } = render(<PwaBanner pwa={pwa({})} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('offers the update, and no install line (#439)', async () => {
    const p = pwa({ updateReady: true, installable: true });
    const user = userEvent.setup();
    render(<PwaBanner pwa={p} />);

    await user.click(screen.getByRole('button', { name: '更新して再読み込み' }));
    expect(p.applyUpdate).toHaveBeenCalledTimes(1);
    expect(screen.queryByText(/アプリとして使えます/)).not.toBeInTheDocument();
  });

  it('offers to install by a small button only while the browser can (#439)', async () => {
    const { rerender } = render(<InstallButton pwa={pwa({})} />);
    expect(screen.queryByRole('button')).not.toBeInTheDocument();

    const p = pwa({ installable: true });
    rerender(<InstallButton pwa={p} />);
    await userEvent.setup().click(screen.getByRole('button', { name: 'アプリにする' }));
    expect(p.promptInstall).toHaveBeenCalledTimes(1);
  });

  it('explains what still works offline', () => {
    render(<PwaBanner pwa={pwa({ online: false })} />);
    expect(
      screen.getByText(/オフラインです。とっておいた例は開いて再生できます/),
    ).toBeInTheDocument();
  });
});
