import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { DownloadMidiButton } from './DownloadMidiButton';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 7, mode: 'major' },
  tracks: [],
  harmony: [{ start: 0, duration: 1, chord: { root: 7, quality: 'maj' }, romanNumeral: 'I' }],
};

describe('DownloadMidiButton', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('is disabled without a timeline', () => {
    render(<DownloadMidiButton timeline={null} saveFile={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'MIDI を保存' })).toBeDisabled();
  });

  it('fetches the file and saves it under a descriptive name', async () => {
    const bytes = new Uint8Array([0x4d, 0x54, 0x68, 0x64]);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(bytes, { status: 200 })));
    const saveFile = vi.fn();
    const user = userEvent.setup();
    render(<DownloadMidiButton timeline={timeline} saveFile={saveFile} />);

    await user.click(screen.getByRole('button', { name: 'MIDI を保存' }));

    await waitFor(() => expect(saveFile).toHaveBeenCalledTimes(1));
    const [blob, name] = saveFile.mock.calls[0] as [Blob, string];
    expect(name).toBe('middy-G-major-I.mid');
    expect(new Uint8Array(await blob.arrayBuffer())).toEqual(bytes);
    expect(screen.getByRole('button', { name: 'MIDI を保存' })).toBeEnabled();
  });

  it('says the export failed, in Japanese, and does not save', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ code: 'invalid_request', detail: 'ppq must be positive' }), {
          status: 400,
        }),
      ),
    );
    const saveFile = vi.fn();
    const user = userEvent.setup();
    render(<DownloadMidiButton timeline={timeline} saveFile={saveFile} />);

    await user.click(screen.getByRole('button', { name: 'MIDI を保存' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('MIDI の書き出しに失敗しました');
    expect(saveFile).not.toHaveBeenCalled();
  });
});
