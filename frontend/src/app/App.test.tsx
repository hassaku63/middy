import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { App } from './App';

describe('App', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('shows backend status once the health check succeeds', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          new Response(JSON.stringify({ status: 'ok', version: '1.2.3' }), { status: 200 }),
        ),
    );

    render(<App />);

    expect(screen.getByRole('heading', { name: 'middy' })).toBeInTheDocument();
    expect(await screen.findByRole('status')).toHaveTextContent('バックエンド: ok (1.2.3)');
  });

  it('shows an error message when the backend is unreachable', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('network')));

    render(<App />);

    expect(await screen.findByRole('status')).toHaveTextContent('バックエンドに接続できません');
  });
});
