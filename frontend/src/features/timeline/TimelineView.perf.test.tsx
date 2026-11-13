import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import * as notation from '@/shared/lib/notation';
import type { Timeline } from '@/shared/api/client';
import { TimelineView } from './TimelineView';

/** Names a note per note drawn: counts how often the piano roll draws (#500). */
vi.mock('@/shared/lib/notation', async (importOriginal) => {
  const real = await importOriginal<typeof notation>();
  return { ...real, pitchName: vi.fn(real.pitchName) };
});

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

/** A second of playing: the player reports the position about 40 times (#500). */
const FRAMES = 40;
/**
 * A move of the playhead costs at most this share of drawing the whole
 * timeline again (#500). A share, not milliseconds, so a slow machine does
 * not fail it; the full redraw was about 1.1 and the playhead move is
 * about 0.45 of it.
 */
const FRAME_SHARE = 0.75;

/** Milliseconds per call of draw, over n calls. */
function timePer(n: number, draw: (i: number) => void) {
  const t0 = performance.now();
  for (let i = 1; i <= n; i += 1) draw(i);
  return (performance.now() - t0) / n;
}

describe('TimelineView rendering budget', () => {
  it(`renders ${NOTES} notes within ${BUDGET_MS} ms`, () => {
    const tl = bigTimeline();
    const t0 = performance.now();
    const { container } = render(<TimelineView timeline={tl} positionTick={0} />);
    const first = performance.now() - t0;
    expect(container.querySelectorAll('.timeline__note')).toHaveLength(NOTES);
    expect(first).toBeLessThan(BUDGET_MS);
  });

  it(`draws again only what moved over ${FRAMES} moves of the playhead (#500)`, () => {
    const tl = bigTimeline();
    const { container, rerender } = render(<TimelineView timeline={tl} positionTick={0} playing />);
    const step = 1920 / FRAMES; // a bar a second, at 120 BPM in 4/4
    timePer(10, (f) => {
      rerender(<TimelineView timeline={tl} positionTick={f * step} playing />); // warm up
    });
    const move = timePer(FRAMES, (f) => {
      rerender(<TimelineView timeline={tl} positionTick={f * step} playing />);
    });
    // Still marks the note under the playhead after the last move.
    expect(container.querySelectorAll('[data-sounding="true"]')).toHaveLength(1);
    // A new timeline object draws everything again.
    const full = timePer(10, () => {
      rerender(<TimelineView timeline={{ ...tl }} positionTick={1920} playing />);
    });
    expect(move / full).toBeLessThan(FRAME_SHARE);
  });

  it('does not draw the notes again while the same notes sound (#500)', () => {
    const tl = bigTimeline();
    const { rerender } = render(<TimelineView timeline={tl} positionTick={0} playing />);
    vi.mocked(notation.pitchName).mockClear();
    for (let tick = 3; tick < 120; tick += 3) {
      rerender(<TimelineView timeline={tl} positionTick={tick} playing />); // within the first note
    }
    expect(notation.pitchName).not.toHaveBeenCalled();
    rerender(<TimelineView timeline={tl} positionTick={120} playing />); // the next note
    expect(notation.pitchName).toHaveBeenCalled();
  });
});
