import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { TimelineOverview } from './TimelineOverview';
import { scrollFor } from './useTimelineScroll';

/** bars one-bar chords (I each) at 120 BPM. */
function bars(n: number): Timeline {
  return {
    ppq: 480,
    tempoBpm: 120,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 0, mode: 'major' },
    tracks: [],
    harmony: Array.from({ length: n }, (_, i) => ({
      start: i * 1920,
      duration: 1920,
      chord: { root: 0, quality: 'maj' as const },
      romanNumeral: 'I',
    })),
  };
}

describe('TimelineOverview (#360)', () => {
  it('draws a short timeline once, whole, with the label column held over it (#402 A)', () => {
    const { container } = render(<TimelineOverview timeline={bars(4)} positionTick={0} />);
    expect(container.querySelectorAll('svg.timeline')).toHaveLength(1);
    const labels = container.querySelector('.timeline-overview__scroll > svg.timeline-labels');
    expect(labels?.getAttribute('aria-hidden')).toBe('true');
    expect(labels?.getAttribute('viewBox')).toBe('0 0 40 182');
  });

  it('draws a long one whole in a scrolling box, with a strip of the whole piece', () => {
    const { container } = render(
      <TimelineOverview timeline={bars(16)} positionTick={0} onSelectHarmony={vi.fn()} />,
    );
    const main = container.querySelector('.timeline-overview__scroll svg.timeline');
    const strip = container.querySelector('svg.timeline--overview');
    expect(container.querySelectorAll('svg.timeline')).toHaveLength(2);
    // The scrolled drawing is the whole piece, not a slice.
    expect(main?.getAttribute('viewBox')).toBe(
      `0 0 ${40 + 16 * 160} ${main?.getAttribute('height')}`,
    );
    expect(main?.getAttribute('aria-hidden')).toBeNull();
    expect(strip?.getAttribute('aria-hidden')).toBe('true');
    expect(strip?.getAttribute('preserveAspectRatio')).toBe('none');
    // The duplicate drawing's chords can be clicked but are no tab stops.
    expect(strip?.querySelectorAll('[role=option][tabindex]')).toHaveLength(0);
    expect(strip?.querySelector('.timeline__highlight')).not.toBeNull();
    expect(container.querySelector('input[type=range]')).toBeNull();
  });
});

describe('scrollFor (#360)', () => {
  // A 1000 px view scrolled to 2000: the lead before the point is 100 px.
  it.each([
    ['inside the view: stays', 2500, 'page', null],
    ['near the right edge: turns the page', 2950, 'page', 2850],
    ['before the view: turns back', 1500, 'page', 1400],
    ['near the left edge: turns back a little', 2050, 'page', 1950],
    ['centre: always to the middle', 2500, 'centre', 2000],
  ] as const)('%s', (_, x, keep, want) => {
    expect(scrollFor(2000, 1000, x, keep)).toBe(want);
  });
});
