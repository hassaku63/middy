import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { PlayerControls } from './PlayerControls';
import type { Synth } from './TimelinePlayer';
import { usePlayer, type AudioEngine } from './usePlayer';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 120,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [
    {
      name: 'chords',
      channel: 0,
      program: 0,
      notes: [{ pitch: 60, velocity: 80, start: 0, duration: 1920 * 2 }],
    },
  ],
  harmony: [],
};

function Harness({ engine, tl }: { engine: () => AudioEngine; tl: Timeline | null }) {
  const player = usePlayer(engine);
  return <PlayerControls timeline={tl} player={player} />;
}

describe('PlayerControls + usePlayer', () => {
  let created: number;
  let synth: Synth & { played: number; allOffCalls: number };
  const engine = (): AudioEngine => {
    created += 1;
    return { synth, clock: { now: () => Date.now() / 1000 }, resume: () => Promise.resolve() };
  };

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    created = 0;
    synth = {
      played: 0,
      allOffCalls: 0,
      play() {
        this.played += 1;
      },
      allOff() {
        this.allOffCalls += 1;
      },
    };
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('is disabled without a timeline', () => {
    render(<Harness engine={engine} tl={null} />);
    expect(screen.getByRole('button', { name: '再生' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '停止' })).toBeDisabled();
    expect(screen.getByLabelText(/^再生位置/)).toHaveTextContent('-');
    expect(created).toBe(0);
  });

  it('creates the audio engine only when play is pressed, then plays and stops', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<Harness engine={engine} tl={timeline} />);
    expect(created).toBe(0);
    expect(screen.getByLabelText(/^再生位置/)).toHaveTextContent('1.1');

    await user.click(screen.getByRole('button', { name: '再生' }));
    expect(created).toBe(1);
    expect(synth.played).toBe(1);
    // Play turns into pause in the same place (#402 B); stop is beside it (BL-334).
    expect(screen.queryByRole('button', { name: '再生' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '一時停止' })).toBeEnabled();
    expect(screen.getByRole('button', { name: '停止' })).toBeEnabled();
    expect(screen.getAllByRole('button')).toHaveLength(2);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1100); // just over two beats at 120 BPM
    });
    expect(screen.getByLabelText(/^再生位置/)).toHaveTextContent('1.3');

    // Pause stays at 1.3; stop goes back to the beginning (BL-334).
    await user.click(screen.getByRole('button', { name: '一時停止' }));
    expect(synth.allOffCalls).toBe(1);
    expect(screen.getByLabelText(/^再生位置/)).toHaveTextContent('1.3');
    await user.click(screen.getByRole('button', { name: '停止' }));
    expect(screen.getByLabelText(/^再生位置/)).toHaveTextContent('1.1');
    expect(screen.getByRole('button', { name: '再生' })).toBeEnabled();

    // Second play reuses the engine.
    await user.click(screen.getByRole('button', { name: '再生' }));
    expect(created).toBe(1);
  });
});
