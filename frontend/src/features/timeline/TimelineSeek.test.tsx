import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { beatAt, computeLayout } from './layout';
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

  it('a finger holding the playhead pulls the score: right goes back (BL-337)', () => {
    const onSeek = vi.fn<(tick: number) => void>();
    const onScrub = vi.fn();
    const { container } = render(
      <TimelineView timeline={timeline} positionTick={3840} onSeek={onSeek} onScrub={onScrub} />,
    );
    const svg = container.querySelector('svg.timeline');
    const grip = container.querySelector('[data-testid=head-grip]');
    if (!(svg instanceof SVGSVGElement) || !(grip instanceof SVGElement)) throw new Error();
    // At its own size: 160 px a bar, 40 px a beat.
    svg.getBoundingClientRect = () => new DOMRect(0, 0, layout.width, layout.height);
    grip.setPointerCapture = vi.fn();
    fireEvent.pointerDown(grip, { clientX: 400, pointerType: 'touch' });
    expect(onScrub).toHaveBeenLastCalledWith('start', 'head');
    fireEvent.pointerMove(grip, { clientX: 400 + 90, pointerType: 'touch' }); // 2¼ beats back
    expect(onSeek).toHaveBeenLastCalledWith(3840 - 3 * 480);
    fireEvent.pointerMove(grip, { clientX: 400 - 40, pointerType: 'touch' }); // a beat ahead
    expect(onSeek).toHaveBeenLastCalledWith(3840 + 480);
    fireEvent.pointerUp(grip, { pointerType: 'touch' });
    expect(onScrub).toHaveBeenLastCalledWith('end', 'head');
    // The tap's click does not seek.
    fireEvent.click(grip, { clientX: 360 });
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
