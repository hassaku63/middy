import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { computeLayout } from './layout';
import { TimelineView } from './TimelineView';

/** Four bars of C at 120 BPM. */
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
      notes: [{ pitch: 60, velocity: 80, start: 0, duration: 4 * 1920 }],
    },
  ],
  harmony: [
    { start: 0, duration: 4 * 1920, chord: { root: 0, quality: 'maj' }, romanNumeral: 'I' },
  ],
};
const layout = computeLayout(timeline);

function setup() {
  const onSeek = vi.fn<(tick: number) => void>();
  const onScrub = vi.fn();
  const view = render(
    <TimelineView timeline={timeline} positionTick={3840} onSeek={onSeek} onScrub={onScrub} />,
  );
  const svg = view.container.querySelector('svg.timeline');
  const grip = view.container.querySelector('[data-testid=head-grip]');
  if (!(svg instanceof SVGSVGElement) || !(grip instanceof SVGElement)) throw new Error();
  svg.getBoundingClientRect = () => new DOMRect(0, 0, layout.width, layout.height);
  svg.setPointerCapture = vi.fn();
  grip.setPointerCapture = vi.fn();
  return { ...view, svg, grip, onSeek, onScrub };
}

describe('pointers on the timeline (#500)', () => {
  it('a second finger neither takes the playhead again nor ends the hold', () => {
    const { grip, onSeek, onScrub } = setup();
    const finger = (id: number, clientX: number) => ({
      pointerId: id,
      clientX,
      pointerType: 'touch',
    });
    fireEvent.pointerDown(grip, finger(1, 400));
    fireEvent.pointerDown(grip, finger(2, 300));
    expect(onScrub).toHaveBeenCalledTimes(1);
    fireEvent.pointerMove(grip, finger(2, 300 + 90));
    expect(onSeek).not.toHaveBeenCalled();
    fireEvent.pointerUp(grip, finger(2, 300 + 90));
    expect(onScrub).toHaveBeenCalledTimes(1);
    fireEvent.pointerMove(grip, finger(1, 400 + 90));
    expect(onSeek).toHaveBeenLastCalledWith(3840 - 3 * 480);
    fireEvent.pointerUp(grip, finger(1, 400 + 90));
    expect(onScrub).toHaveBeenLastCalledWith('end', 'head');
  });

  it('a hold whose capture is lost ends, so play goes on', () => {
    const { grip, onScrub } = setup();
    fireEvent.pointerDown(grip, { pointerId: 1, clientX: 400, pointerType: 'touch' });
    fireEvent(
      grip,
      Object.assign(new Event('lostpointercapture', { bubbles: true }), { pointerId: 9 }),
    );
    expect(onScrub).toHaveBeenCalledTimes(1); // another pointer's loss is not this hold's
    fireEvent(
      grip,
      Object.assign(new Event('lostpointercapture', { bubbles: true }), { pointerId: 1 }),
    );
    expect(onScrub).toHaveBeenLastCalledWith('end', 'head');
  });

  it("a mouse drag ends when its capture is lost, and ignores another pointer's release", () => {
    const { svg, onScrub } = setup();
    fireEvent.pointerDown(svg, { pointerId: 1, clientX: 60, pointerType: 'mouse', button: 0 });
    fireEvent.pointerUp(svg, { pointerId: 2, pointerType: 'pen' });
    expect(onScrub).toHaveBeenCalledTimes(1);
    fireEvent(
      svg,
      Object.assign(new Event('lostpointercapture', { bubbles: true }), { pointerId: 1 }),
    );
    expect(onScrub).toHaveBeenLastCalledWith('end');
  });

  it('a drag ends when the drawing is taken away under it', () => {
    const { svg, grip, onScrub, unmount } = setup();
    fireEvent.pointerDown(svg, { pointerId: 1, clientX: 60, pointerType: 'mouse', button: 0 });
    fireEvent.pointerDown(grip, { pointerId: 2, clientX: 400, pointerType: 'touch' });
    unmount();
    expect(onScrub).toHaveBeenCalledWith('end');
    expect(onScrub).toHaveBeenCalledWith('end', 'head');
  });
});
