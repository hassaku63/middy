import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import {
  BAR_WIDTH,
  beatAt,
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

describe('the grid (#435)', () => {
  it('draws each bar and, fainter, each beat under the notes', () => {
    const { container } = render(<TimelineView timeline={timeline} positionTick={0} />);
    const roll = container.querySelector('.timeline__note')?.parentElement;
    const grid = roll?.querySelector('.timeline__grid');
    // Three bars: four bar lines, and three beats inside each bar.
    expect(grid?.querySelectorAll('.timeline__bar-line')).toHaveLength(4);
    expect(grid?.querySelectorAll('.timeline__beat')).toHaveLength(9);
    // Over the rows, under the notes.
    const children = Array.from(roll?.children ?? []);
    const at = (cls: string) => children.findIndex((c) => c.classList.contains(cls));
    expect(at('timeline__grid')).toBeGreaterThan(-1);
    expect(at('timeline__grid')).toBeLessThan(at('timeline__note'));
  });

  it('counts a compound meter in dotted quarters', () => {
    const six = { ...timeline, timeSignature: { numerator: 6, denominator: 8 } };
    const { container } = render(<TimelineView timeline={six} positionTick={0} />);
    const grid = container.querySelector('.timeline__grid');
    const bars = grid?.querySelectorAll('.timeline__bar-line').length ?? 0;
    // A 6/8 bar has two beats: one beat line in each bar.
    expect(grid?.querySelectorAll('.timeline__beat')).toHaveLength(bars - 1);
  });

  it('leaves the beats off a stretched overview strip', () => {
    const { container } = render(<TimelineView timeline={timeline} positionTick={0} stretch />);
    expect(container.querySelectorAll('.timeline__beat')).toHaveLength(0);
    expect(container.querySelectorAll('.timeline__bar-line').length).toBeGreaterThan(0);
  });
});

describe('the bar numbers (#435)', () => {
  it('numbers each bar above the chord lane, and not on a thin strip', () => {
    const { container, rerender } = render(<TimelineView timeline={timeline} positionTick={0} />);
    const numbers = Array.from(
      container.querySelectorAll('.timeline__bar-number'),
      (t) => t.textContent,
    );
    expect(numbers).toEqual(['1', '2', '3']);
    rerender(<TimelineView timeline={timeline} positionTick={0} stretch />);
    expect(container.querySelectorAll('.timeline__bar-number')).toHaveLength(0);
  });
});

describe('clicking to move the playhead (BL-328)', () => {
  const layout = computeLayout(timeline); // 40 px of names, 160 px a bar, 3 bars
  it.each([
    ['over the row names: nothing', 20, null],
    ['the first beat', 41, 0],
    ['the middle of a beat: its start', 40 + 160 + 60, 1920 + 480],
    ['past the last beat: the last beat', 40 + 3 * 160 + 30, 3 * 1920 - 480],
  ] as const)('%s', (_, x, want) => {
    expect(beatAt(layout, x)).toBe(want);
  });

  it('counts a compound meter in dotted quarters', () => {
    const six = computeLayout({ ...timeline, timeSignature: { numerator: 6, denominator: 8 } });
    // A 6/8 bar is 1440 ticks of two 720-tick beats, 160 px wide.
    expect(beatAt(six, 40 + 100)).toBe(720);
  });

  /** The drawing at half its size, so a click is scaled to the drawing. */
  function renderSeekable(onSeek = vi.fn<(tick: number) => void>(), onSelect = vi.fn()) {
    const { container } = render(
      <TimelineView
        timeline={timeline}
        positionTick={0}
        onSeek={onSeek}
        onSelectHarmony={onSelect}
      />,
    );
    const svg = container.querySelector('svg.timeline');
    if (!(svg instanceof SVGSVGElement)) throw new Error('no drawing');
    const half = layout.width / 2;
    svg.getBoundingClientRect = () => new DOMRect(100, 0, half, layout.height / 2);
    return { svg, onSeek, onSelect };
  }

  it('a click on the drawing moves the playhead to the start of the beat under it', () => {
    const { svg, onSeek } = renderSeekable();
    // 40 + 160 + 60 in drawing units: bar 2, past its second beat's start.
    fireEvent.click(svg, { clientX: 100 + (40 + 160 + 60) / 2 });
    expect(onSeek).toHaveBeenCalledWith(1920 + 480);
    expect(svg.classList.contains('timeline--seekable')).toBe(true);
  });

  it('a click on a chord pins it and leaves the playhead alone', async () => {
    const { onSeek, onSelect } = renderSeekable();
    await userEvent.click(screen.getByRole('option', { name: /V G/ }));
    expect(onSelect).toHaveBeenCalledWith(1);
    expect(onSeek).not.toHaveBeenCalled();
  });

  it('a mouse shows the beat a click would go to; a finger does not', () => {
    const { svg } = renderSeekable();
    const x = 100 + (40 + 330) / 2; // bar 3, beat 1 (40 + 320 + 10)
    fireEvent.pointerMove(svg, { clientX: x, pointerType: 'mouse' });
    expect(svg.querySelector('.timeline__seek-hint')?.getAttribute('x1')).toBe(
      String(layout.x(3840)),
    );
    fireEvent.pointerMove(svg, { clientX: x, pointerType: 'touch' });
    expect(svg.querySelector('.timeline__seek-hint')).toBeNull();
  });

  it('a mouse moves the playhead on the press and drags it along (BL-335)', () => {
    const onSeek = vi.fn<(tick: number) => void>();
    const onScrub = vi.fn();
    const { container } = render(
      <TimelineView timeline={timeline} positionTick={0} onSeek={onSeek} onScrub={onScrub} />,
    );
    const svg = container.querySelector('svg.timeline');
    if (!(svg instanceof SVGSVGElement)) throw new Error('no drawing');
    svg.getBoundingClientRect = () => new DOMRect(100, 0, layout.width / 2, layout.height / 2);
    svg.setPointerCapture = vi.fn();
    const x = (drawing: number) => 100 + drawing / 2;
    fireEvent.pointerDown(svg, { clientX: x(40 + 10), pointerType: 'mouse', button: 0 });
    expect(onScrub).toHaveBeenLastCalledWith('start');
    expect(onSeek).toHaveBeenLastCalledWith(0);
    fireEvent.pointerMove(svg, { clientX: x(40 + 160 + 10), pointerType: 'mouse' });
    expect(onSeek).toHaveBeenLastCalledWith(1920);
    fireEvent.pointerMove(svg, { clientX: x(40 + 160 + 20), pointerType: 'mouse' }); // same beat
    expect(onSeek).toHaveBeenCalledTimes(2);
    fireEvent.pointerUp(svg, { pointerType: 'mouse' });
    expect(onScrub).toHaveBeenLastCalledWith('end');
    // The click that follows the press does not seek again.
    fireEvent.click(svg, { clientX: x(40 + 160 + 20) });
    expect(onSeek).toHaveBeenCalledTimes(2);
  });

  it('a finger only clicks: a press does not move the playhead', () => {
    const { svg, onSeek } = renderSeekable();
    fireEvent.pointerDown(svg, { clientX: 100 + 300 / 2, pointerType: 'touch', button: 0 });
    expect(onSeek).not.toHaveBeenCalled();
  });

  it('without onSeek, a click does nothing and shows no hint', () => {
    const { container } = render(<TimelineView timeline={timeline} positionTick={0} />);
    const svg = container.querySelector('svg.timeline');
    expect(svg?.classList.contains('timeline--seekable')).toBe(false);
  });
});
