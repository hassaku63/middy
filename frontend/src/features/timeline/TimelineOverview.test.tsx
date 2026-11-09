import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { TimelineOverview } from './TimelineOverview';

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

describe('TimelineOverview (BL-080)', () => {
  it('draws a short timeline once, whole', () => {
    const { container } = render(<TimelineOverview timeline={bars(4)} positionTick={0} />);
    expect(container.querySelectorAll('svg.timeline')).toHaveLength(1);
    expect(container.querySelector('.timeline--overview')).toBeNull();
  });

  it('draws a long one as an overview and a four-bar detail that follows the playhead', () => {
    const { container, rerender } = render(
      <TimelineOverview timeline={bars(16)} positionTick={0} onSelectHarmony={vi.fn()} />,
    );
    const overview = container.querySelector('svg.timeline--overview');
    expect(container.querySelectorAll('svg.timeline')).toHaveLength(2);
    expect(overview?.getAttribute('aria-hidden')).toBe('true');
    // The duplicate drawing's chords can be clicked but are no tab stops.
    expect(overview?.querySelectorAll('[role=option][tabindex]')).toHaveLength(0);
    expect(overview?.querySelector('.timeline__highlight')?.getAttribute('x')).toBe('40');

    rerender(
      <TimelineOverview timeline={bars(16)} positionTick={5 * 1920} onSelectHarmony={vi.fn()} />,
    );
    // Bar 6 is on the second page: bars 5–8.
    expect(
      container.querySelector('svg.timeline--overview .timeline__highlight')?.getAttribute('x'),
    ).toBe(String(40 + 4 * 160));
  });
});
