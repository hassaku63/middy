import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import {
  BAR_WIDTH,
  computeLayout,
  DRUM_NAMES,
  fitBarWidth,
  harmonyAt,
  LABEL_WIDTH,
} from './layout';
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
    expect(chords.map((c) => c.getAttribute('aria-label'))).toEqual(['I C', 'V G', 'vi Am']);
  });

  it('highlights the chord and notes under the playhead', () => {
    const { container, rerender } = render(
      <TimelineView timeline={timeline} positionTick={2000} />,
    );
    expect(screen.getByRole('option', { current: true })).toHaveAttribute('aria-label', 'V G');
    const sounding = [...container.querySelectorAll('[data-sounding="true"]')].map((n) =>
      n.getAttribute('data-pitch'),
    );
    expect(sounding).toEqual(['G4', 'B4', 'D5']);

    rerender(<TimelineView timeline={timeline} positionTick={4000} />);
    expect(screen.getByRole('option', { current: true })).toHaveAttribute('aria-label', 'vi Am');
  });

  it('runs the playhead under the chord lane, so it never hides a numeral (#402 I)', () => {
    render(<TimelineView timeline={timeline} positionTick={0} />);
    const playhead = screen.getByTestId('playhead');
    const lane = screen.getByRole('listbox', { name: 'コード' });
    const laneBottom = Math.max(
      ...Array.from(lane.querySelectorAll('rect')).map(
        (r) => Number(r.getAttribute('y')) + Number(r.getAttribute('height')),
      ),
    );
    expect(Number(playhead.getAttribute('y1'))).toBeGreaterThanOrEqual(laneBottom);
  });

  it('widens the bars to fill a box wider than the drawing, never narrows them (#402 F)', () => {
    const natural = computeLayout(timeline);
    // A box exactly the drawing's proportions keeps BAR_WIDTH.
    expect(fitBarWidth(natural, { width: natural.width, height: natural.height })).toBe(BAR_WIDTH);
    // Twice as wide at the same height: the bars take the extra width.
    const wide = fitBarWidth(natural, { width: natural.width * 2, height: natural.height });
    const fitted = computeLayout(timeline, wide);
    expect(fitted.width).toBeCloseTo(natural.width * 2);
    expect(fitted.height).toBe(natural.height); // the rows keep their size
    // A narrower box shrinks the drawing as before; the bars stay at BAR_WIDTH.
    expect(fitBarWidth(natural, { width: natural.width / 2, height: natural.height })).toBe(
      BAR_WIDTH,
    );
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
    expect(screen.getByRole('option', { selected: true })).toHaveAttribute('aria-label', 'vi Am');
    await user.click(screen.getByRole('option', { name: 'V G' }));
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
    expect(screen.getByText('キック')).toBeInTheDocument();
    expect(screen.getByText('スネア')).toBeInTheDocument();
  });

  it("names the drums in the rhythm notes' words, in a column wide enough (#402 E)", () => {
    const withAll: Timeline = {
      ...timeline,
      tracks: [
        {
          name: 'drums',
          channel: 9,
          program: 0,
          notes: Object.keys(DRUM_NAMES).map((pitch) => ({
            pitch: Number(pitch),
            velocity: 80,
            start: 0,
            duration: 120,
          })),
        },
      ],
    };
    const layout = computeLayout(withAll);
    const longest = Math.max(...Object.values(DRUM_NAMES).map((n) => n.length));
    expect(layout.labelWidth).toBeGreaterThanOrEqual(longest * 9 + 4);
    expect(computeLayout(timeline).labelWidth).toBe(LABEL_WIDTH); // no drums: the pitch names'
    expect(layout.x(0)).toBe(layout.labelWidth);
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
    for (const name of ['低いタム', '中くらいのタム', '高いタム', 'サイドスティック']) {
      expect(screen.getByText(name)).toBeInTheDocument();
    }
  });

  it('marks where a pushed chord starts to sound, left of its block (#385)', () => {
    const pushed: Timeline = {
      ...timeline,
      harmony: timeline.harmony.map((h, i) => (i === 1 ? { ...h, push: 'eighth' as const } : h)),
    };
    const { container } = render(<TimelineView timeline={pushed} positionTick={0} />);
    const tabs = container.querySelectorAll('.timeline__push');
    expect(tabs).toHaveLength(1);
    const block = container.querySelectorAll('.timeline__chord > rect:first-child')[1];
    const tab = tabs[0];
    const tabX = Number(tab?.getAttribute('x'));
    const tabEnd = tabX + Number(tab?.getAttribute('width'));
    expect(tabEnd).toBeCloseTo(Number(block?.getAttribute('x')));
    expect(tabX).toBeLessThan(tabEnd);
    // The tab is drawn only; the option's name says what it shows.
    expect(screen.getByRole('option', { name: /押し/ })).toHaveAttribute(
      'aria-label',
      'V G、8 分音符ひとつ早く入る（押し）',
    );
  });

  it('tells the foot hi-hat from the closed hi-hat (#385)', () => {
    const jazz: Timeline = {
      ...timeline,
      tracks: [
        ...timeline.tracks,
        {
          name: 'drums',
          channel: 9,
          program: 0,
          notes: [42, 44, 51].map((pitch, i) => ({
            pitch,
            velocity: 60,
            start: i * 240,
            duration: 120,
          })),
        },
      ],
    };
    render(<TimelineView timeline={jazz} positionTick={0} />);
    for (const name of ['ハイハット', '足のハイハット', 'ライド']) {
      expect(screen.getByText(name)).toBeInTheDocument();
    }
  });

  it('marks nothing as sounding while stopped, though the playhead sits at the start', () => {
    const { container } = render(
      <TimelineView timeline={timeline} positionTick={0} playing={false} />,
    );
    expect(container.querySelector('.timeline__chord.is-active')).toBeNull();
    expect(container.querySelector('[aria-current="true"]')).toBeNull();
    expect(container.querySelector('.is-sounding')).toBeNull();
    expect(container.querySelector('.timeline__playhead')).not.toBeNull();
  });

  it('renders an empty timeline as a single empty bar', () => {
    const empty: Timeline = { ...timeline, tracks: [], harmony: [] };
    render(<TimelineView timeline={empty} positionTick={0} />);
    expect(screen.getByRole('group', { name: 'タイムライン' })).toBeInTheDocument();
    expect(screen.queryAllByRole('option')).toHaveLength(0);
  });
});

describe('marks (BL-185)', () => {
  const marks = (container: HTMLElement) =>
    Array.from(container.querySelectorAll('.timeline__mark')).map((m) => m.textContent);

  it('marks the chord followed with ▶, and only when asked', () => {
    const { container, rerender } = render(
      <TimelineView timeline={timeline} positionTick={2000} />,
    );
    expect(marks(container)).toEqual([]);
    rerender(<TimelineView timeline={timeline} positionTick={2000} marks />);
    expect(marks(container)).toEqual(['▶']);
  });

  it('marks the pinned chord with ■ and hides the marks from assistive tech', () => {
    const { container } = render(
      <TimelineView timeline={timeline} positionTick={2000} selectedHarmony={2} marks />,
    );
    expect(marks(container)).toEqual(['■']);
    expect(container.querySelector('.timeline__mark')?.getAttribute('aria-hidden')).toBe('true');
    expect(screen.getByRole('option', { name: 'vi Am' })).toBeInTheDocument();
  });
});
