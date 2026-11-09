import { fireEvent, render, screen } from '@testing-library/react';
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

  describe('following the playhead (B, BL-127)', () => {
    const detailX = (container: HTMLElement) =>
      Number(
        container.querySelector('svg.timeline--overview .timeline__highlight')?.getAttribute('x'),
      );

    it('keeps the playhead at the centre while playing, inside the piece', () => {
      const { container, rerender } = render(
        <TimelineOverview follow timeline={bars(16)} positionTick={6 * 1920} playing />,
      );
      // Bar 7's start (x = 40 + 6·160 = 1000) at the centre of a 640-wide detail.
      expect(detailX(container)).toBe(1000 - 320);
      rerender(<TimelineOverview follow timeline={bars(16)} positionTick={0} playing />);
      expect(detailX(container)).toBe(40); // not before the first bar
      rerender(
        <TimelineOverview follow timeline={bars(16)} positionTick={16 * 1920 - 1} playing />,
      );
      expect(detailX(container)).toBe(40 + 16 * 160 - 640); // nor past the last
      expect(screen.queryByRole('slider', { name: '拡大の位置' })).toBeNull();
    });

    it('lets a slider move it while stopped', () => {
      const { container } = render(
        <TimelineOverview follow timeline={bars(16)} positionTick={0} playing={false} />,
      );
      const slider = screen.getByRole('slider', { name: '拡大の位置' });
      fireEvent.change(slider, { target: { value: String(8 * 1920) } });
      expect(detailX(container)).toBe(40 + 8 * 160 - 320);
    });

    it('centres a pinned chord while stopped', () => {
      const { container } = render(
        <TimelineOverview
          follow
          timeline={bars(16)}
          positionTick={0}
          playing={false}
          selectedHarmony={10}
        />,
      );
      expect(detailX(container)).toBe(40 + 10.5 * 160 - 320);
      expect(screen.queryByRole('slider', { name: '拡大の位置' })).toBeNull();
    });
  });
});
