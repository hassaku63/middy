import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { BAR_WIDTH, harmonyAt } from './layout';
import { TimelineView } from './TimelineView';

/** C | G | Am (one bar each, block triads) at 120 BPM. */
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
      notes: [
        [60, 64, 67],
        [67, 71, 74],
        [69, 72, 76],
      ].flatMap((ps, bar) =>
        ps.map((pitch) => ({ pitch, velocity: 80, start: bar * 1920, duration: 1920 })),
      ),
    },
  ],
  harmony: [
    { start: 0, duration: 1920, chord: { root: 0, quality: 'maj' }, romanNumeral: 'I' },
    { start: 1920, duration: 1920, chord: { root: 7, quality: 'maj' }, romanNumeral: 'V' },
    { start: 3840, duration: 1920, chord: { root: 9, quality: 'min' }, romanNumeral: 'vi' },
  ],
};

describe('harmonyAt', () => {
  it('finds the chord sounding at a tick', () => {
    expect(harmonyAt(timeline.harmony, 0)).toBe(0);
    expect(harmonyAt(timeline.harmony, 1919)).toBe(0);
    expect(harmonyAt(timeline.harmony, 1920)).toBe(1);
    expect(harmonyAt(timeline.harmony, 5760)).toBe(-1);
  });
});

describe('TimelineView', () => {
  it('renders one chord block per harmony event with numeral and symbol', () => {
    render(<TimelineView timeline={timeline} positionTick={0} />);
    const chords = screen.getAllByRole('option');
    expect(chords.map((c) => c.getAttribute('aria-label'))).toEqual(['I (C)', 'V (G)', 'vi (Am)']);
  });

  it('highlights the chord and notes under the playhead', () => {
    const { container, rerender } = render(
      <TimelineView timeline={timeline} positionTick={2000} />,
    );
    expect(screen.getByRole('option', { current: true })).toHaveAttribute('aria-label', 'V (G)');
    const sounding = [...container.querySelectorAll('[data-sounding="true"]')].map((n) =>
      n.getAttribute('aria-label'),
    );
    expect(sounding).toEqual(['G4', 'B4', 'D5']);

    rerender(<TimelineView timeline={timeline} positionTick={4000} />);
    expect(screen.getByRole('option', { current: true })).toHaveAttribute('aria-label', 'vi (Am)');
  });

  it('moves the playhead one bar width per bar', () => {
    const { rerender } = render(<TimelineView timeline={timeline} positionTick={0} />);
    const x0 = Number(screen.getByTestId('playhead').getAttribute('x1'));
    rerender(<TimelineView timeline={timeline} positionTick={1920} />);
    const x1 = Number(screen.getByTestId('playhead').getAttribute('x1'));
    expect(x1 - x0).toBe(BAR_WIDTH);
  });

  it('reports clicks on chords and marks the selected one', async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(
      <TimelineView
        timeline={timeline}
        positionTick={0}
        selectedHarmony={2}
        onSelectHarmony={onSelect}
      />,
    );
    expect(screen.getByRole('option', { selected: true })).toHaveAttribute('aria-label', 'vi (Am)');
    await user.click(screen.getByRole('option', { name: 'V (G)' }));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('draws percussion in a drum lane instead of the piano roll', () => {
    const withDrums: Timeline = {
      ...timeline,
      tracks: [
        ...timeline.tracks,
        {
          name: 'drums',
          channel: 9,
          program: 0,
          notes: [
            { pitch: 36, velocity: 100, start: 0, duration: 240 },
            { pitch: 38, velocity: 96, start: 480, duration: 240 },
            { pitch: 42, velocity: 60, start: 0, duration: 120 },
          ],
        },
      ],
    };
    const { container } = render(<TimelineView timeline={withDrums} positionTick={0} />);
    const lane = screen.getByTestId('drum-lane');
    expect(lane).toBeInTheDocument();
    expect(lane.querySelectorAll('.timeline__drum-hit')).toHaveLength(3);
    expect(lane.querySelectorAll('.timeline__drum-hit.is-sounding')).toHaveLength(2); // kick + hat at 0
    // Drum notes must not appear as pitched notes nor stretch the pitch range.
    expect(container.querySelectorAll('.timeline__note')).toHaveLength(9);
    expect(screen.getByText('Kick')).toBeInTheDocument();
    expect(screen.getByText('Snare')).toBeInTheDocument();
  });

  it('names the bodhrán and grid drums: toms and side stick (review)', () => {
    const bodhran: Timeline = {
      ...timeline,
      tracks: [
        ...timeline.tracks,
        {
          name: 'drums',
          channel: 9,
          program: 0,
          notes: [45, 47, 50, 37].map((pitch, i) => ({
            pitch,
            velocity: 90,
            start: i * 240,
            duration: 240,
          })),
        },
      ],
    };
    render(<TimelineView timeline={bodhran} positionTick={0} />);
    for (const name of ['Tom L', 'Tom M', 'Tom H', 'Stick']) {
      expect(screen.getByText(name)).toBeInTheDocument();
    }
  });

  it('renders an empty timeline as a single empty bar', () => {
    const empty: Timeline = { ...timeline, tracks: [], harmony: [] };
    render(<TimelineView timeline={empty} positionTick={0} />);
    expect(screen.getByRole('img', { name: 'タイムライン' })).toBeInTheDocument();
    expect(screen.queryAllByRole('option')).toHaveLength(0);
  });
});
