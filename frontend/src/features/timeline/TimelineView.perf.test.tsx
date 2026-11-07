import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { TimelineView } from './TimelineView';

/**
 * Rendering budget (BL-052): a 1,000-note timeline must render, and stay
 * far from pathological. The bound is deliberately loose (jsdom is slow
 * and CI machines vary); `make test-report` shows the actual time on
 * every run, so drift is visible long before this fails.
 */
const NOTES = 1000;
const BUDGET_MS = 3000;

function bigTimeline(): Timeline {
  const notes = Array.from({ length: NOTES }, (_, i) => ({
    pitch: 48 + (i % 24),
    velocity: 80,
    start: i * 120,
    duration: 120,
  }));
  const bars = Math.ceil((NOTES * 120) / 1920);
  return {
    ppq: 480,
    tempoBpm: 120,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 0, mode: 'major' },
    tracks: [{ name: 'chords', channel: 0, program: 0, notes }],
    harmony: Array.from({ length: bars }, (_, b) => ({
      start: b * 1920,
      duration: 1920,
      chord: { root: 0, quality: 'maj' as const },
      romanNumeral: 'I',
    })),
  };
}

describe('TimelineView rendering budget', () => {
  it(`renders ${NOTES} notes within ${BUDGET_MS} ms and re-renders a playhead move quickly`, () => {
    const tl = bigTimeline();
    const t0 = performance.now();
    const { container, rerender } = render(<TimelineView timeline={tl} positionTick={0} />);
    const first = performance.now() - t0;
    expect(container.querySelectorAll('.timeline__note')).toHaveLength(NOTES);

    const t1 = performance.now();
    rerender(<TimelineView timeline={tl} positionTick={60_000} />);
    const move = performance.now() - t1;

    expect(first).toBeLessThan(BUDGET_MS);
    expect(move).toBeLessThan(BUDGET_MS);
  });
});
